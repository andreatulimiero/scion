// Code generated by go/proto/structs_gen_go.sh; DO NOT EDIT.

package proto

import (
	"zombiezen.com/go/capnproto2"

	"github.com/scionproto/scion/go/lib/common"
)

// NewRootStruct calls the appropriate NewRoot<x> function corresponding to the capnp proto type ID,
// and returns the inner capnp.Struct that it receives. This allows the helper
// functions in cereal.go to support generic capnp root struct types.
func NewRootStruct(id ProtoIdType, seg *capnp.Segment) (capnp.Struct, error) {
	var blank capnp.Struct
	switch id {
	case ASEntry_TypeID:
		v, err := NewRootASEntry(seg)
		if err != nil {
			return blank, common.NewBasicError("Error creating new ASEntry capnp struct", err)
		}
		return v.Struct, nil
	case CtrlPld_TypeID:
		v, err := NewRootCtrlPld(seg)
		if err != nil {
			return blank, common.NewBasicError("Error creating new CtrlPld capnp struct", err)
		}
		return v.Struct, nil
	case PathSegment_TypeID:
		v, err := NewRootPathSegment(seg)
		if err != nil {
			return blank, common.NewBasicError("Error creating new PathSegment capnp struct", err)
		}
		return v.Struct, nil
	case PathSegmentSignedData_TypeID:
		v, err := NewRootPathSegmentSignedData(seg)
		if err != nil {
			return blank, common.NewBasicError("Error creating new PathSegmentSignedData capnp struct", err)
		}
		return v.Struct, nil
	case RevInfo_TypeID:
		v, err := NewRootRevInfo(seg)
		if err != nil {
			return blank, common.NewBasicError("Error creating new RevInfo capnp struct", err)
		}
		return v.Struct, nil
	case SignedBlob_TypeID:
		v, err := NewRootSignedBlob(seg)
		if err != nil {
			return blank, common.NewBasicError("Error creating new SignedBlob capnp struct", err)
		}
		return v.Struct, nil
	case SignedCtrlPld_TypeID:
		v, err := NewRootSignedCtrlPld(seg)
		if err != nil {
			return blank, common.NewBasicError("Error creating new SignedCtrlPld capnp struct", err)
		}
		return v.Struct, nil
	case SVCResolutionReply_TypeID:
		v, err := NewRootSVCResolutionReply(seg)
		if err != nil {
			return blank, common.NewBasicError("Error creating new SVCResolutionReply capnp struct", err)
		}
		return v.Struct, nil
	case ColibriRequestPayload_TypeID:
		v, err := NewRootColibriRequestPayload(seg)
		if err != nil {
			return blank, common.NewBasicError("Error creating new ColibriRequestPayload capnp struct", err)
		}
		return v.Struct, nil
	}
	return blank, common.NewBasicError(
		"Unsupported capnp struct type (i.e. not listed in go/proto/structs_gen_go.sh:ROOTTYPES)",
		nil,
		"id", id,
	)
}

func (s ASEntry) GetStruct() capnp.Struct {
	return s.Struct
}
func (s CtrlPld) GetStruct() capnp.Struct {
	return s.Struct
}
func (s PathSegment) GetStruct() capnp.Struct {
	return s.Struct
}
func (s PathSegmentSignedData) GetStruct() capnp.Struct {
	return s.Struct
}
func (s RevInfo) GetStruct() capnp.Struct {
	return s.Struct
}
func (s SignedBlob) GetStruct() capnp.Struct {
	return s.Struct
}
func (s SignedCtrlPld) GetStruct() capnp.Struct {
	return s.Struct
}
func (s SVCResolutionReply) GetStruct() capnp.Struct {
	return s.Struct
}
func (s ColibriRequestPayload) GetStruct() capnp.Struct {
	return s.Struct
}
