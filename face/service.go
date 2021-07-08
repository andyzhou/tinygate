package face

import (
	"github.com/andyzhou/gate/define"
	pb "github.com/andyzhou/gate/proto"
	"log"
)

/*
 * service face, implement of IService
 * - used for remote sub rpc service node
 * - dynamic service face
 * - async process in channel
 * - one client node one service instance
 */

 //face info
 type Service struct {
	 remoteAddr string //client node remote address
	 stream *pb.GateService_BindStreamServer //stream server from client node
	 clientRespChan chan pb.ByteMessage //chan for send client response
	 closeChan chan bool
 }
 
 //construct
func NewService(
				remoteAddr string, //remote client address
				stream *pb.GateService_BindStreamServer,
			) *Service {
	//self init
	this := &Service{
		remoteAddr:remoteAddr,
		stream:stream,
		clientRespChan:make(chan pb.ByteMessage, define.ResponseChanSize),
		closeChan:make(chan bool, 1),
	}

	//spawn main process
	go this.runMainProcess()
	return this
}

//////////////////////
//implement of IService
//////////////////////

//quit
func (f *Service) Quit() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Service:Quit panic, err:", err)
		}
	}()

	//send to chan
	f.closeChan <- true
}

 //send resp to client node
 //used for stream mode
func (f *Service) SendClientResp(resp *pb.ByteMessage) (bRet bool) {
	//basic check
	if resp == nil || resp.MessageId < 0 {
		bRet = false
		return
	}

	//try catch panic
	defer func() {
		if err := recover(); err != nil {
			log.Println("Service::SendClientResp panic, err:", err)
			bRet = false
			return
		}
	}()

	//send to chan
	f.clientRespChan <- *resp
	bRet = true
	return
}

//get remote client address
func (f *Service) GetRemoteAddr() string {
	return f.remoteAddr
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
		resp pb.ByteMessage //response for client
		needQuit, isOk bool
		err error
	)

	//defer close chan
	defer func() {
		if err := recover(); err != nil {
			log.Println("Service::runMainProcess panic, err:", err)
		}
		close(f.clientRespChan)
		close(f.closeChan)
	}()

	//loop
	for {
		if needQuit {
			break
		}
		select {
		case resp, isOk = <- f.clientRespChan:
			if isOk && f.stream != nil {
				//cast to client node pass stream mode
				err = (*f.stream).Send(&resp)
				if err != nil {
					log.Println("Service::runMainProcess send failed, err:",
								err.Error())
				}
			}
		case <- f.closeChan:
			needQuit = true
		}
	}
}