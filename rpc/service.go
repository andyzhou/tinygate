package rpc

import (
	"context"
	"errors"
	"github.com/andyzhou/gate/define"
	"github.com/andyzhou/gate/iface"
	pb "github.com/andyzhou/gate/proto"
	"io"
	"log"
)

/*
 * rpc request service implement
 * - stream service for server side
 * - stream data from rpc client side
 * - general request from rpc client side
 */

 //response from service
 type Response struct {
 	//kind string
 	remoteAddr string
 	byteMessage pb.ByteMessage
 }

 //service info
 type Service struct {
 	node iface.INode
 	clientStreamMap map[string]pb.GateService_BindStreamServer //remoteAddr -> stream interface
 	cbForStreamReq func(remoteAddr string, req *pb.ByteMessage) bool //cb for client stream request
 	cbForGenReq func(req *pb.GateReq) *pb.GateResp //cb for client gen request
	respChan chan Response //chan for send response
	closeChan chan bool
 	Base
 }

 //construct, step-1
func NewService() *Service {
	//self init
	this := &Service{
		clientStreamMap: make(map[string]pb.GateService_BindStreamServer),
		respChan:make(chan Response, define.ResponseChanSize),
		closeChan:make(chan bool, 1),
	}

	return this
}

//quit
func (r *Service) Quit() {
	//catch panic
	defer func() {
		if err := recover(); err != nil {
			log.Println("rpc Service:Quit panic, err:", err)
		}
	}()

	//send to close chan
	r.closeChan <- true
}
 
//set not face
func (r *Service) SetNodeFace(node iface.INode) bool {
	if node == nil {
		return false
	}
	r.node = node
	return true
}

//set cb for client general request
func (r *Service) SetCBForGenReq(cb func(req *pb.GateReq) *pb.GateResp) bool {
	if cb == nil {
		return false
	}
	r.cbForGenReq = cb
	return true
}

//set cb for client stream request
func (r *Service) SetCBForStreamReq(cb func(remoteAddr string, req *pb.ByteMessage) bool) bool {
	if cb == nil {
		return false
	}
	r.cbForStreamReq = cb
	return true
}

 //send stream data to remote client
func (r *Service) SendToClient(remoteAddr string, in *pb.ByteMessage) error {
	//basic check
	if remoteAddr == "" || in == nil {
		return errors.New("invalid parameter")
	}

	//get client stream
	stream, ok := r.clientStreamMap[remoteAddr]
	if !ok || stream == nil {
		return errors.New("can't get stream by address")
	}

	//send to client
	err := stream.SendMsg(in)
	return err
}

//implement interface of `GenReq`
//this is sync request
func (r *Service) GenReq(ctx context.Context, in *pb.GateReq) (*pb.GateResp, error) {
	if in == nil {
		return nil, errors.New("invalid parameter")
	}
	if r.cbForGenReq == nil {
		return nil, errors.New("invalid cb for gen request")
	}

	//call the cb func to process general requests
	resp := r.cbForGenReq(in)
	if resp == nil {
		return nil, errors.New("invalid response")
	}
	return resp, nil
}

 //implement interface of `BindStream`
 //receive stream data from rpc client side
func (r *Service) BindStream(stream pb.GateService_BindStreamServer) error {
	var (
		in *pb.ByteMessage
		err error
		tips string
		remoteAddr string
		messageId uint32
	)

	//get context
	ctx := stream.Context()

	//get tag by stream
	tag, ok := r.GetConnTagFromContext(ctx)
	if !ok {
		tips = "Can't get tag from node stream."
		log.Println(tips)
		return errors.New(tips)
	}

	//get remote addr
	remoteAddr = tag.RemoteAddr.String()

	//add remote stream into map
	r.clientStreamMap[remoteAddr] = stream

	//client node up
	r.node.ClientNodeUp(remoteAddr, &stream)

	//try receive stream data from node
	for {
		select {
		case <- ctx.Done():
			log.Println("Stream::BindStream, Receive down signal from client")
			return ctx.Err()
		default:
			//receive data from client
			in, err = stream.Recv()
			if err == io.EOF {
				log.Println("Stream::BindStream, Read done")
				return nil
			}
			if err != nil {
				log.Println("Stream::BindStream, " +
							"Read error:", err.Error())
				return err
			}

			//get message id
			messageId = in.MessageId

			//do relate opt by message id
			switch messageId {
			default:
				{
					//input stream data from rpc client node side
					if r.cbForStreamReq != nil {
						r.cbForStreamReq(remoteAddr, in)
					}
				}
			}
		}
	}

	return nil
}