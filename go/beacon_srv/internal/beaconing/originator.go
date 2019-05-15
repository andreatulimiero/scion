// Copyright 2019 Anapaya Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package beaconing

import (
	"context"
	"time"

	"github.com/scionproto/scion/go/beacon_srv/internal/beaconing/metrics"
	"github.com/scionproto/scion/go/beacon_srv/internal/ifstate"
	"github.com/scionproto/scion/go/beacon_srv/internal/onehop"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/ctrl/seg"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/periodic"
	"github.com/scionproto/scion/go/lib/spath"
	"github.com/scionproto/scion/go/lib/util"
	"github.com/scionproto/scion/go/proto"
)

var _ periodic.Task = (*Originator)(nil)

// OriginatorConf is the configuration to create a new originator.
type OriginatorConf struct {
	Config        ExtenderConf
	Sender        *onehop.Sender
	Period        time.Duration
	EnableMetrics bool
}

// Originator originates beacons. It should only be used by core ASes.
type Originator struct {
	*segExtender
	sender  *onehop.Sender
	metrics *metrics.Originator

	// tick is mutable.
	tick tick
}

// New creates a new originator.
func (cfg OriginatorConf) New() (*Originator, error) {

	cfg.Config.task = "originator"
	extender, err := cfg.Config.new()
	if err != nil {
		return nil, err
	}
	o := &Originator{
		sender:      cfg.Sender,
		segExtender: extender,
		tick:        tick{period: cfg.Period},
	}
	if cfg.EnableMetrics {
		o.metrics = metrics.InitOriginator()
	}
	return o, nil
}

// Run originates core and downstream beacons.
func (o *Originator) Run(_ context.Context) {
	o.tick.now = time.Now()
	o.originateBeacons(proto.LinkType_core)
	o.originateBeacons(proto.LinkType_child)
	o.metrics.AddTotalTime(o.tick.now)
	o.tick.updateLast()
}

// originateBeacons creates and sends a beacon for each active interface of
// the specified link type.
func (o *Originator) originateBeacons(linkType proto.LinkType) {
	active, nonActive := sortedIntfs(o.cfg.Intfs, linkType)
	if len(nonActive) > 0 && o.tick.passed() {
		log.Debug("[Originator] Ignore non-active interfaces", "ifids", nonActive)
	}
	intfs := o.needBeacon(active)
	if len(intfs) == 0 {
		return
	}
	infoF := o.createInfoF(o.tick.now)
	s := newSummary()
	for _, ifid := range intfs {
		b := beaconOriginator{
			Originator: o,
			ifId:       ifid,
			infoF:      infoF,
			summary:    s,
		}
		if err := b.originateBeacon(); err != nil {
			log.Error("[Originator] Unable to originate on interface", "ifid", ifid, "err", err)
		}
	}
	o.logSummary(s, linkType)
}

// createInfoF creates the info field.
func (o *Originator) createInfoF(now time.Time) spath.InfoField {
	infoF := spath.InfoField{
		ConsDir: true,
		ISD:     uint16(o.sender.IA.I),
		TsInt:   util.TimeToSecs(now),
	}
	return infoF
}

// needBeacon returns a list of interfaces that need a beacon.
func (o *Originator) needBeacon(active []common.IFIDType) []common.IFIDType {
	if o.tick.passed() {
		return active
	}
	stale := make([]common.IFIDType, 0, len(active))
	for _, ifid := range active {
		intf := o.cfg.Intfs.Get(ifid)
		if intf == nil {
			continue
		}
		if o.tick.now.Sub(intf.LastOriginate()) > o.tick.period {
			stale = append(stale, ifid)
		}
	}
	return stale
}

func (o *Originator) logSummary(s *summary, linkType proto.LinkType) {
	if o.tick.passed() {
		log.Info("[Originator] Originated beacons", "type", linkType.String(), "egIfIds", s.IfIds())
		return
	}
	log.Info("[Originator] Originated beacons on stale interfaces", "type", linkType.String(),
		"egIfIds", s.IfIds())
}

// beaconOriginator originates one beacon on the given interface.
type beaconOriginator struct {
	*Originator
	ifId    common.IFIDType
	infoF   spath.InfoField
	summary *summary
}

// originateBeacon originates a beacon on the given ifid.
func (o *beaconOriginator) originateBeacon() error {
	intf := o.cfg.Intfs.Get(o.ifId)
	if intf == nil {
		o.metrics.IncInternalErr()
		return common.NewBasicError("Interface does not exist", nil)
	}
	topoInfo := intf.TopoInfo()
	msg, err := o.createBeaconMsg(topoInfo.ISD_AS)
	if err != nil {
		o.metrics.IncTotalBeacons(o.ifId, metrics.CreateErr)
		return err
	}
	ov := topoInfo.InternalAddrs.PublicOverlay(topoInfo.InternalAddrs.Overlay)
	if err := o.sender.Send(msg, ov); err != nil {
		o.metrics.IncTotalBeacons(o.ifId, metrics.SendErr)
		return common.NewBasicError("Unable to send packet", err)
	}
	o.onSuccess(intf)
	return nil
}

// createBeaconMsg creates a beacon for the given interface, signs it and
// wraps it in a one-hop message.
func (o *beaconOriginator) createBeaconMsg(remoteIA addr.IA) (*onehop.Msg, error) {
	bseg, err := o.createBeacon()
	if err != nil {
		return nil, common.NewBasicError("Unable to create beacon", err, "ifid", o.ifId)
	}
	return packBeaconMsg(bseg, remoteIA, o.ifId, o.cfg.Signer)
}

func (o *beaconOriginator) createBeacon() (*seg.Beacon, error) {
	bseg, err := seg.NewSeg(&o.infoF)
	if err != nil {
		return nil, common.NewBasicError("Unable to create segment", err)
	}
	if err := o.extend(bseg, 0, o.ifId, nil); err != nil {
		return nil, common.NewBasicError("Unable to extend segment", err)
	}
	return &seg.Beacon{Segment: bseg}, nil
}

func (o *beaconOriginator) onSuccess(intf *ifstate.Interface) {
	intf.Originate(o.tick.now)
	o.summary.AddIfid(o.ifId)
	o.summary.Inc()
	o.metrics.IncTotalBeacons(o.ifId, metrics.Success)
}