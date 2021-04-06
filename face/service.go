package face

import (
	"github.com/andyzhou/gate/define"
	pb "github.com/andyzhou/gate/proto"
	"log"
)

/*
 * service face, implement of IService
 * - remote sub rpc service node
 * - dynamic service face
 * - async process in channel
 */

 //face info
 type Service struct {
	 kind string
	 tag string
	 remoteAddr string
	 stream *pb.GateService_BindStreamServer
	 clientReqChan chan pb.ByteMessage //client request
	 closeChan chan bool
 }
 
 //construct
func NewService(
				kind, tag, remoteAddr string,
				stream *pb.GateService_BindStreamServer,
			) *Service {
	//self init
	this := &Service{
		kind:kind,
		tag:tag,
		remoteAddr:remoteAddr,
		stream:stream,
		clientReqChan:make(chan pb.ByteMessage, define.ClientReqChanSize),
		closeChan:make(chan bool, 1),
	}
	return this
}

//////////////////////
//implement of IService
//////////////////////

//quit
func (f *Service) Quit() {
	f.closeChan <- true
}

//tcp client request on current service
func (f *Service) ClientReq(req *pb.ByteMessage) (bRet bool) {
	//basic check
	if req == nil || req.MessageId < 0 {
		bRet = false
		return
	}

	//try catch panic
	defer func() {
		if err := recover(); err != nil {
			log.Println("Service::ClientReq panic, err:", err)
			bRet = false
			return
		}
	}()

	//send to chan
	f.clientReqChan <- *req
	bRet = true

	return
}

//get
func (f *Service) GetRemoteAddr() string {
	return f.remoteAddr
}

func (f *Service) GetKind() string {
	return f.kind
}

func (f *Service) GetTag() string {
	return f.tag
}

func (f *Service) GetStream() *pb.GateService_BindStreamServer {
	return f.stream
}


////////////////
//private func
////////////////

//run main process
func (f *Service) runMainProcess() {
	var (
		clientReq pb.ByteMessage
		needQuit, isOk bool
		err error
	)

	//defer close chan
	defer func() {
		close(f.clientReqChan)
		close(f.closeChan)
	}()

	//loop
	for {
		if needQuit {
			break
		}
		select {
		case clientReq, isOk = <- f.clientReqChan:
			if isOk && f.stream != nil {
				//cast to target sub service pass stream mode
				err = (*f.stream).Send(&clientReq)
				if err != nil {
					log.Println("Service::runMainProcess send failed, err:", err.Error())
				}
			}
		case <- f.closeChan:
			needQuit = true
		}
	}
}