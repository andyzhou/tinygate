package gate

import (
	"fmt"
	"github.com/andyzhou/gate/face"
	"github.com/andyzhou/gate/json"
	"github.com/andyzhou/gate/proto"
	"github.com/andyzhou/gate/rpc"
	"google.golang.org/grpc"
	"log"
	"net"
)

/*
 * gate service
 *
 * - api for service side
 * - base on g-rpc
 */


//service info
type Service struct {
	address string `rpc service address`
	rpc *rpc.Service `rpc service instance`
	service *grpc.Server `g-rpc server`
}

//construct
func NewService(port int) *Service {
	//self init
	address := fmt.Sprintf(":%d", port)
	this := &Service{
		address:address,
		rpc:rpc.NewService(),
		service:nil,
	}
	return this
}

//stop
func (r *Service) Stop() {
	if r.service != nil {
		r.service.Stop()
	}
	if r.rpc != nil {
		r.rpc.Quit()
	}
}

//start
func (r *Service) Start() {
	//create rpc service
	r.createService()
}

//set cb for bind or unbind node
//if sub service send `MessageIdOfBindOrUnbind`, need call the cb
func (r *Service) SetCBForBindUnBindNode(cb func(obj *json.BindJson) bool) bool {
	return r.rpc.SetCBForBindUnBindNode(cb)
}

//set cb for sub service response cast
//if have response from sub service, need call the cb
func (r *Service) SetCBForResponseCast(cb func(connIds []uint32, messageId uint32, data []byte) bool) bool {
	return r.rpc.SetCBForResponseCast(cb)
}

////////////////////////////
//multi kind send data
////////////////////////////

//send data to assigned address of gate client
func (r *Service) SendClientReqByAddress(address string, req *gate.ByteMessage) bool {
	//basic check
	if address == "" || req == nil {
		return false
	}

	//get node face
	nodeFace := face.RunInterFace.GetNodeFace()
	if nodeFace == nil {
		return false
	}

	//get service by address
	subService := nodeFace.GetService(address)
	if subService == nil {
		return false
	}

	//cast data
	bRet := subService.SendClientReq(req)
	return bRet
}

//send data to assigned kind gate client
func (r *Service) SendClientReqByKind(kind string, req *gate.ByteMessage) bool {
	//basic check
	if kind == "" || req == nil {
		return false
	}

	//get node face
	nodeFace := face.RunInterFace.GetNodeFace()
	if nodeFace == nil {
		return false
	}

	//get node tag by kind
	nodeTag := nodeFace.PickNode(kind)
	if nodeTag == "" {
		return false
	}

	//get service by kind and tag
	subService := nodeFace.GetServiceByTag(kind, nodeTag)
	if subService == nil {
		return false
	}

	//cast data
	bRet := subService.SendClientReq(req)
	return bRet
}

//send data to all gate clients
func (r *Service) SendClientReqToAll(req *gate.ByteMessage) bool {
	//basic check
	if req == nil {
		return false
	}

	//get node face
	nodeFace := face.RunInterFace.GetNodeFace()
	if nodeFace == nil {
		return false
	}

	//get all sub service
	allSubService := nodeFace.GetAllService()
	if allSubService == nil || len(allSubService) <= 0 {
		return false
	}

	//send one by one
	for _, service := range allSubService {
		service.SendClientReq(req)
	}
	return true
}

/////////////////
//private func
/////////////////

//create rpc service
func (r *Service) createService() {
	var (
		tips string
		err error
	)

	//try listen tcp port
	listen, err := net.Listen("tcp", r.address)
	if err != nil {
		tips = "Create rpc service failed, error:" + err.Error()
		log.Println(tips)
		panic(tips)
	}

	//create rpc server with rpc stat support
	r.service = grpc.NewServer(
					grpc.StatsHandler(rpc.NewStat()),
				)

	//register call back
	gate.RegisterGateServiceServer(r.service, r.rpc)

	//begin rpc service
	go r.beginService(listen)
}

//begin rpc service
func (r *Service) beginService(listen net.Listener) {
	//service listen
	err := r.service.Serve(listen)
	if err != nil {
		tips := "Failed for rpc service, error:" + err.Error()
		log.Println(tips)
		panic(tips)
	}
}
