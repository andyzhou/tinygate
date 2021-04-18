package iface

import (
	"github.com/andyzhou/gate/json"
	pb "github.com/andyzhou/gate/proto"
)

/*
 * interface for node
 */

 type INode interface {
 	Quit()
 	GetService(address string) IService
 	GetAllService() map[string]IService
 	ClientNodeDown(address string) bool
 	ClientNodeUp(address string, jsonObj *json.NodeJson, stream *pb.GateService_BindStreamServer) bool

 	//set cb for client node down
 	SetCBForClientNodeDown(cb func(remoteAddr string) bool) bool
 }
