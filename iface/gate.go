package iface

import pb "github.com/andyzhou/gate/proto"

/*
 * interface for gate for client side
 */

type IGate interface {
	Quit()
	SendGenReq(in *pb.GateReq) *pb.GateResp
	CastData(in *pb.ByteMessage) bool
	Connect(isReConn bool) bool

	//get
	GetKind() string //service kind
	GetTags() []string //unique tags
	GetConnStat()string

	//check
	ConnIsNil() bool

	//set cb
	SetCBForStreamReceived(cb func(from string, in *pb.ByteMessage) bool) bool
	SetCBForGateServerDown(cb func(kind, address string) bool) bool
	SetCBForGateServerUp(cb func(kind, address string) bool) bool
}
