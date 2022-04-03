package main

import (
	"fmt"
	"github.com/andyzhou/gate"
	pb "github.com/andyzhou/gate/proto"
	"log"
	"math/rand"
	"sync"
	"time"
)

/*
 * gate server demo for sub service side.
 */

const (
	//rpc service port
	rpcPort = 7100
	maxMessageId = 100
)

//cb stream request from gate server
func cbForStreamReq(remoteAddr string, req *pb.ByteMessage) bool {
	log.Println("cbForStreamReq, remoteAddr:", remoteAddr)
	return true
}

//cb for the request from gate client side
//this for the sync request
func cbForGenReq(in *pb.GateReq) *pb.GateResp {
	log.Println("cbForGenReq, in messageId:", in.MessageId)

	//init resp
	resp := &pb.GateResp{
		ErrorCode: 1,
		ErrorMessage: "test",
	}
	return resp
}

func main() {
	wg := new(sync.WaitGroup)

	//try catch panic
	defer func(wg *sync.WaitGroup) {
		if err := recover(); err != nil {
			fmt.Println("panic, err:", err)
			wg.Done()
		}
	}(wg)

	//init
	s := gate.NewService(rpcPort)
	
	//set cb
	//cb for stream data request process
	s.SetCBForStreamReq(cbForStreamReq)

	//cb for general request process
	s.SetCBForGenReq(cbForGenReq)

	//wg add
	wg.Add(1)

	//start
	log.Printf("start service,localhost:%d\n", rpcPort)
	s.Start()

	//send data to client
	go sendDataToGateClient(s)

	wg.Wait()
	log.Println("stop service..")
}

//send data to gate client
func sendDataToGateClient(s *gate.Service) {
	var (
		messageId int
		msgPara = "server side message %v"
		msg string
		in = pb.ByteMessage{}
		ticker = time.NewTicker(time.Second * 3)
	)

	//defer
	defer func() {
		ticker.Stop()
	}()

	//loop
	for {
		select {
		case <- ticker.C:
			{
				messageId = rand.Intn(maxMessageId)
				msg = fmt.Sprintf(msgPara, time.Now().Unix())
				in.MessageId = uint32(messageId)
				in.Data = []byte(msg)
				s.SendStreamDataRespToAll(&in)
			}
		}
	}
}