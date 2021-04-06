package iface

import pb "github.com/andyzhou/gate/proto"

/*
 * interface for gate
 */

type IGate interface {
	Quit()
	CastData(in *pb.ByteMessage) bool
	Connect(isReConn bool) bool

	//get
	GetKind() string
	GetTag() string
	GetConnStat()string

	//check
	ConnIsNil() bool

	//set cb
	SetCBForStreamReceive(cb func(from string, in *pb.ByteMessage) bool) bool
	SetCBForGateDown(cb func(kind, address string) bool) bool
}
