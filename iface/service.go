package iface

import (
	pb "github.com/andyzhou/gate/proto"
)

/*
 * interface for sub service
 */

 type IService interface {
 	Quit()
 	SendClientReq(req *pb.ByteMessage) bool
 	GetRemoteAddr() string
 	GetKind() string
 	GetTag() string
 	GetStream() *pb.GateService_BindStreamServer
 }
