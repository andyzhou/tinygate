package tinygate

import (
	"errors"
	"fmt"
	"github.com/andyzhou/tinygate/face"
	"github.com/andyzhou/tinygate/iface"
	pb "github.com/andyzhou/tinygate/proto"
	"github.com/andyzhou/tinygate/rpc"
	"google.golang.org/grpc"
	"log"
	"net"
)

/*
 * gate service
 *
 * - service api for sub service side
 * - base on g-rpc
 * - receive request from client side
 * - communicate pass rpc protocol
 */

//service info
type Service struct {
	address string //rpc service address
	node iface.INode //client node manage instance
	rpc *rpc.Service //rpc service instance
	service *grpc.Server //g-rpc server
}

//construct
func NewService(rpcPort int) *Service {
	//self init
	address := fmt.Sprintf(":%d", rpcPort)
	this := &Service{
		address:address,
		node: face.NewNode(),
		rpc:rpc.NewService(),
	}
	//set node face for rpc service
	this.rpc.SetNodeFace(this.node)
	return this
}

//stop
func (r *Service) Stop() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Service:Stop panic, err:", err)
		}
	}()
	//do some cleanup
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

///////////////////
//api
///////////////////

//send stream data to gate client by remote address
func (r *Service) SendStreamDataResp(
					resp *pb.ByteMessage,
					address ...string,
				) error {
	var (
		subService iface.IService
	)

	//basic check
	if address == nil || resp == nil {
		return errors.New("invalid parameter")
	}
	if r.node == nil {
		return errors.New("nod is nil")
	}

	//send one by one
	for _, oneAddr := range address {
		//get target client by address
		subService = r.node.GetService(oneAddr)
		if subService == nil {
			continue
		}

		//cast resp stream data to client node
		subService.SendClientResp(resp)
	}
	return nil
}

//send stream data to all gate clients
func (r *Service) SendStreamDataRespToAll(
					resp *pb.ByteMessage,
				) error {
	//basic check
	if resp == nil || r.node == nil {
		return errors.New("invalid parameter")
	}

	//get all sub service
	allSubService := r.node.GetAllService()
	if allSubService == nil || len(allSubService) <= 0 {
		return errors.New("no any sub service")
	}

	//send one by one
	for _, service := range allSubService {
		service.SendClientResp(resp)
	}
	return nil
}

///////////////////
//relate cb setup
///////////////////

//set cb for client node down
func (r *Service) SetCBForClientNodeDown(cb func(remoteAddr string) bool) bool {
	if r.node == nil {
		return false
	}
	return r.node.SetCBForClientNodeDown(cb)
}

//set cb of stream request from gate client
func (r *Service) SetCBForStreamReq(cb func(remoteAddr string, in *pb.ByteMessage) bool) error {
	return r.rpc.SetCBForStreamReq(cb)
}

//set cb of response for general request from gate client
func (r *Service) SetCBForGenReq(cb func(req *pb.GateReq) *pb.GateResp) error {
	return r.rpc.SetCBForGenReq(cb)
}

/////////////////
//private func
/////////////////

//create rpc service
func (r *Service) createService() {
	//try listen tcp port
	listen, err := net.Listen("tcp", r.address)
	if err != nil {
		tips := "Create rpc service failed, error:" + err.Error()
		log.Println(tips)
		panic(tips)
	}

	//init rpc stat
	rpcStat := rpc.NewStat(r.node)

	//create rpc server with rpc stat support
	r.service = grpc.NewServer(
					grpc.StatsHandler(rpcStat),
				)

	//register call back
	pb.RegisterGateServiceServer(r.service, r.rpc)

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
