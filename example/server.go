package main

import (
	"fmt"
	"github.com/andyzhou/gate"
	pb "github.com/andyzhou/gate/proto"
	"sync"
	"time"
)

/*
 * gate server demo for sub service.
 */

const (
	//rpc service port
	rpcPort = 7100
	subServiceKind = "game"
)

//cb stream request from gate server
func cbForStreamReq(connIds []uint32, msgId uint32, data []byte) bool {
	fmt.Println("cbForResponseCast, connIds:", connIds,
				", msgId:", msgId, ", data:", string(data))
	return true
}

//cb response for the request from gate client side
//this for the sync request
func cbForGenResp(in *pb.GateReq) *pb.GateResp {
	return nil
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
	//cb for stream data request
	s.SetCBForStreamReq(cbForStreamReq)

	//cb for general request
	s.SetCBForGenReq(cbForGenResp)

	//wg add
	wg.Add(1)

	//start
	fmt.Println("start service..")
	s.Start()

	//send data to client
	go sendDataToGateClient(s)

	wg.Wait()
	fmt.Println("stop service..")
}

//send data to gate client
func sendDataToGateClient(s *gate.Service) {
	var (
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
				in.MessageId = 20
				in.Data = []byte("server side message..")
				s.SendClientReqToAll(&in)
			}
		}
	}
}