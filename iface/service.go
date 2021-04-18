package iface

import (
	pb "github.com/andyzhou/gate/proto"
)

/*
 * interface for sub service
 */

 type IService interface {
 	Quit()
 	SendClientResp(resp *pb.ByteMessage) bool
 	GetRemoteAddr() string
 	GetStream() *pb.GateService_BindStreamServer
 }
