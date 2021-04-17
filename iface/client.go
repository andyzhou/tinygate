package iface

import (
	"github.com/andyzhou/gate/json"
	pb "github.com/andyzhou/gate/proto"
)

/*
 * interface for gate client
 */

type IClient interface {
	Quit()

	//send gen request
	SendGenReq(in *pb.GateReq) *pb.GateResp

	//cast data
	CastDataToOneKind(kind string, in *pb.ByteMessage) bool
	CastDataByTag(tag string, in *pb.ByteMessage) bool
	CastData(address string, in *pb.ByteMessage) bool
	CastDataToAll(in *pb.ByteMessage) bool

	//base opt
	BindNodeTags(fromAddr string, bindJson *json.BindJson) bool
	AddGateServer(kind, host string, port int, tags ...string) bool
	SetLog(dir, tag string) bool

	//set cb
	SetCBForStreamReceived(cb func(from string, in *pb.ByteMessage) bool) bool
	SetCBForGateDown(cb func(kind, addr string) bool) bool
}
