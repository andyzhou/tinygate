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
 * gate client demo, for tcp server side.
 */

const (
	//gate setting
	gateServer = "localhost"
	gatePort = 7100

	//others
	gateServerKind = "chat"
)

//cb for received stream data from gate server
func cbForReceivedStreamData(from string, in *pb.ByteMessage) bool {
	fmt.Println("cbForReceivedStreamData, from:", from, ", in:", string(in.Data))
	return true
}

//cb for gate server down
func cbForGateServerDown(kind, addr string) bool {
	fmt.Println("cbForGateServerDown, kind:", kind, ", addr:", addr)
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
	c := gate.NewClient()

	//set relate cb
	c.SetCBForGateServerDown(cbForGateServerDown)
	c.SetCBForStreamReceived(cbForReceivedStreamData)

	//set log
	c.SetLog("log", "client")

	//add gate server
	bRet := c.AddGateServer(gateServerKind, gateServer, gatePort)
	if !bRet {
		fmt.Println("add gate server failed")
		return
	}

	//wg add
	wg.Add(1)
	fmt.Println("start client..")

	go sendGenReqToGate(c)
	go sendStreamDataToGate(c)

	wg.Wait()
	fmt.Println("stop client..")
}

//send general request to gate
func sendGenReqToGate(c *gate.Client)  {
	var (
		in = pb.GateReq{}
		ticker = time.NewTicker(time.Second * 2)
		messageData string
	)

	//defer
	defer func() {
		ticker.Stop()
	}()

	//loop
	messageIdStart := uint32(30)
	messageIdEnd := 40
	for {
		select {
		case <- ticker.C:
			{
				//set pb data
				in.Service = gateServerKind
				in.MessageId = uint32(rand.Intn(messageIdEnd))
				if in.MessageId < messageIdStart {
					in.MessageId = messageIdStart
				}
				messageData = fmt.Sprintf("%d-%d", in.MessageId, time.Now().Unix())
				in.Data = []byte(messageData)

				//send general request to gate server
				resp := c.SendGenReq(&in)
				if resp != nil {
					fmt.Println("client resp, resp:", resp)
				}
			}
		}
	}
}

//send stream data to gate
func sendStreamDataToGate(c *gate.Client) {
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