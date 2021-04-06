package main

import (
	"fmt"
	"github.com/andyzhou/gate"
	pb "github.com/andyzhou/gate/proto"
	"math/rand"
	"sync"
	"time"
)

/*
 * gate client demo
 */

const (
	gateServer = "localhost"
	gatePort = 7100

	clientKind = "game"
	clientTag = "tag"
)

func cbForReceivedStreamData(from string, in *pb.ByteMessage) bool {
	fmt.Println("cbForReceivedStreamData, from:", from, ", in:", string(in.Data))
	return true
}

func cbForGateDown(kind, addr string) bool {
	fmt.Println("cbForGateDown, kind:", kind, ", addr:", addr)
	return true
}

func main()  {
	//init wg
	wg := new(sync.WaitGroup)

	//try catch panic
	defer func(wg *sync.WaitGroup) {
		if err := recover(); err != nil {
			fmt.Println("panic, err:", err)
			wg.Done()
		}
	}(wg)

	//init client
	c := gate.NewClient(clientKind)

	//set relate cb
	c.SetCBForGateDown(cbForGateDown)
	c.SetCBForStream(cbForReceivedStreamData)

	//set log
	c.SetLog("log", "client")

	//get gate server
	bRet := c.AddGateServer(clientTag, gateServer, gatePort)
	if !bRet {
		fmt.Println("add gate server failed")
		return
	}

	//wg add
	wg.Add(1)
	fmt.Println("start client..")

	go sendDataToGate(c)

	wg.Wait()
	fmt.Println("stop client..")
}

//send data to gate
func sendDataToGate(c *gate.Client) {
	var (
		in = pb.ByteMessage{}
		ticker = time.NewTicker(time.Second * 2)
		messageData string
	)

	//defer
	defer func() {
		ticker.Stop()
	}()

	//loop
	messageIdStart := uint32(10)
	messageIdEnd := 30
	for {
		select {
		case <- ticker.C:
			{
				//set pb data
				in.MessageId = uint32(rand.Intn(messageIdEnd))
				if in.MessageId < messageIdStart {
					in.MessageId = messageIdStart
				}
				messageData = fmt.Sprintf("%d-%d", in.MessageId, time.Now().Unix())
				in.Data = []byte(messageData)

				//cast to all gate server
				c.CastDataToAll(&in)
			}
		}
	}
}