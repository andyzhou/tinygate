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

	//cast data
	CastDataToOneKind(kind string, in *pb.ByteMessage) bool
	CastDataByTag(tag string, in *pb.ByteMessage) bool
	CastData(address string, in *pb.ByteMessage) bool
	CastDataToAll(in *pb.ByteMessage) bool

	//base opt
	BindNodeTags(fromAddr string, bindJson *json.BindJson) bool
	AddGateServer(tag, host string, port int) bool
	SetLog(dir, tag string) bool

	//set cb
	SetCBForStream(cb func(from string, in *pb.ByteMessage) bool) bool
	SetCBForGateDown(cb func(kind, addr string) bool) bool
}
