package iface

import (
	"github.com/andyzhou/gate/json"
	pb "github.com/andyzhou/gate/proto"
)

/*
 * interface for sub service node
 */

 type INode interface {
 	Quit()
 	PickNode(kind string) string
 	GetServiceByTag(kind, tag string) IService
 	GetService(address string) IService
 	GetKind(address string) string
 	NodeDown(address string) bool
 	NodeUp(address string, jsonObj *json.NodeJson, stream *pb.GateService_BindStreamServer) bool

 	//set cb for sub service node down
 	SetCBForNodeDown(cb func(service, remoteAddr string) bool) bool
 }
