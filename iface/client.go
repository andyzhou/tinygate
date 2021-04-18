package iface

import (
	pb "github.com/andyzhou/gate/proto"
)

/*
 * interface for gate client
 */

type IClient interface {
	Quit()

	//send gen request
	SendGenReq(in *pb.GateReq) *pb.GateResp

	//cast stream data
	CastData(address string, in *pb.ByteMessage) bool
	CastDataByKind(kind string, in *pb.ByteMessage) bool
	CastDataToAll(in *pb.ByteMessage) bool

	//base opt
	PickOneGateServer(kind string) IGate
	AddGateServer(kind, host string, port int, tags ...string) bool
	SetLog(dir, tag string) bool

	//set cb func
	SetCBForStreamReceived(cb func(from string, in *pb.ByteMessage) bool) bool
	SetCBForGateServerDown(cb func(kind, addr string) bool) bool
}
