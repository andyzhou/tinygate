package rpc

import (
	"context"
	"errors"
	"github.com/andyzhou/tinygate/define"
	"github.com/andyzhou/tinygate/iface"
	pb "github.com/andyzhou/tinygate/proto"
	"io"
	"log"
	"sync"
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
	closeChan chan struct{}
 	Base
 	sync.RWMutex
 }

 //construct, step-1
func NewService() *Service {
	//self init
	this := &Service{
		clientStreamMap: make(map[string]pb.GateService_BindStreamServer),
		respChan:make(chan Response, define.ResponseChanSize),
		closeChan:make(chan struct{}, 1),
	}
	return this
}

//quit
func (r *Service) Quit() {
	//catch panic
	defer func() {
		if err := recover(); err != nil {
			log.Printf("rpc Service:Quit panic, err:%v", err)
		}
	}()

	//send to close chan
	close(r.closeChan)
}
 
//set node face
func (r *Service) SetNodeFace(node iface.INode) error {
	//check
	if node == nil {
		return errors.New("nod is nil")
	}
	//sync with locker
	r.Lock()
	defer r.Unlock()
	r.node = node
	return nil
}

//set cb for client general request
func (r *Service) SetCBForGenReq(cb func(req *pb.GateReq) *pb.GateResp) error {
	if cb == nil {
		return errors.New("invalid parameter")
	}
	r.Lock()
	defer r.Unlock()
	r.cbForGenReq = cb
	return nil
}

//set cb for client stream request
func (r *Service) SetCBForStreamReq(cb func(remoteAddr string, req *pb.ByteMessage) bool) error {
	if cb == nil {
		return errors.New("invalid parameter")
	}
	r.Lock()
	defer r.Unlock()
	r.cbForStreamReq = cb
	return nil
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
	r.Lock()
	r.clientStreamMap[remoteAddr] = stream
	r.Unlock()

	//client node up
	r.node.ClientNodeUp(remoteAddr, &stream)

	//defer
	defer func() {
		if err := recover(); err != nil {
			log.Println("Stream::BindStream panic, err:", err)
		}
		//clean up
		r.Lock()
		delete(r.clientStreamMap, remoteAddr)
		r.Unlock()
	}()

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
				log.Printf("Stream::BindStream, Read error:%v", err.Error())
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