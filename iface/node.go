package iface

import (
	pb "github.com/andyzhou/tinygate/proto"
)

/*
 * interface for node
 */

 type INode interface {
 	Quit()
 	GetService(address string) IService
 	GetAllService() map[string]IService
 	ClientNodeDown(address string) bool
 	ClientNodeUp(address string, stream *pb.GateService_BindStreamServer) bool

 	//set cb for client node down
 	SetCBForClientNodeDown(cb func(remoteAddr string) bool) bool
 }
