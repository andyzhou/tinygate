package main

import (
	"fmt"
	"github.com/andyzhou/gate"
	pb "github.com/andyzhou/gate/proto"
	"sync"
	"time"
)

/*
 * gate server demo
 */

const (
	rpcPort = 7100
	subServiceKind = "game"
)

func cbForResponseCast(connIds []uint32, msgId uint32, data []byte) bool {
	fmt.Println("cbForResponseCast, connIds:", connIds,
				", msgId:", msgId, ", data:", string(data))
	return true
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
	s.SetCBForResponseCast(cbForResponseCast)

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
				s.SendClientReqByKind(subServiceKind, &in)
			}
		}
	}
}