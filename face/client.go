package face

import (
	"fmt"
	"github.com/andyzhou/tinygate/define"
	"github.com/andyzhou/tinygate/iface"
	pb "github.com/andyzhou/tinygate/proto"
	"log"
	"sync"
	"time"
)

/*
 * gate client face
 *
 * - used by tcp service
 * - api for gate face
 * - support multi gate servers
 * - communicate with gate server pass general and stream mode
 */

//client info
type Client struct {
	gateMap map[string]iface.IGate //running gate server map, serverAddress -> Gate
	cbForStreamReceived func(from string, in *pb.ByteMessage) bool //call back for received data
	cbForGateServerDown func(kind string, addr string) bool //call back for gate server down
	cbForGateServerUp func(kind string, addr string) bool //call back for gate server up
	closeChan chan bool
	sync.Mutex `internal data locker`
}

//construct
//STEP-1
func NewClient() *Client {
	//self init
	this := &Client{
		gateMap:make(map[string]iface.IGate),
		closeChan:make(chan bool, 1),
	}

	//spawn main process
	go this.runMainProcess()

	return this
}

//quit
func (c *Client) Quit() {
	//catch panic
	defer func() {
		if err := recover(); err != nil {
			log.Println("Client:Quit panic, err:", err)
		}
	}()

	//clean gate map
	if c.gateMap != nil {
		for k, gate := range c.gateMap {
			gate.Quit()
			delete(c.gateMap, k)
		}
	}

	//send to close chan
	c.closeChan <- true
}

//set call back for received stream data from server side
//STEP-3
func (c *Client) SetCBForStreamReceived(
			cb func(from string, in *pb.ByteMessage) bool,
		) bool {
	if c.cbForStreamReceived != nil {
		return false
	}
	c.cbForStreamReceived = cb
	return true
}

//set call back for gate server down
//STEP-4
func (c *Client) SetCBForGateServerDown(
				cb func(serviceKind, addr string) bool,
			) bool {
	if c.cbForGateServerDown != nil {
		return false
	}
	c.cbForGateServerDown = cb
	return true
}

//set call back for gate server up
func (c *Client) SetCBForGateServerUp(cb func(kind, addr string) bool) bool {
	if c.cbForGateServerUp != nil {
		return false
	}
	c.cbForGateServerUp = cb
	return true
}

//set log option
//STEP-5, optional
func (c *Client) SetLog(dir, tag string) bool {
	if dir == "" || tag == "" {
		return false
	}
	//set log service
	//c.logService = tc.NewLogService(dir, tag)
	return true
}

//pick one rand gate server by service kind
func (c *Client) PickOneGateServer(serviceKind string) iface.IGate {
	//basic check
	if serviceKind == "" || c.gateMap == nil {
		return nil
	}

	//begin loop gate server map and pick one
	for _, gs := range c.gateMap {
		if gs.GetKind() == serviceKind {
			return gs
		}
	}
	return nil
}

//add gate server
//STEP-4
func (c *Client) AddGateServer(
					serviceKind, host string,
					port int,
					tags ... string,
				) bool {
	//basic check
	if serviceKind == "" || host == "" || port <= 0 {
		return false
	}

	//format address
	address := fmt.Sprintf("%s:%d", host, port)

	//check gate has exists or not
	old := c.getGateByAddr(address)
	if old != nil {
		return true
	}

	//init gate
	gate := NewGate(serviceKind, host, port, tags...)

	//set callback function
	gate.SetCBForStreamReceived(c.cbForStreamReceived)
	gate.SetCBForGateServerDown(c.cbForGateServerDown)
	gate.SetCBForGateServerUp(c.cbForGateServerUp)

	//sync into map
	c.Lock()
	defer c.Unlock()
	c.gateMap[address] = gate

	return true
}

//send general request to remote gate server
func (c *Client) SendGenReq(in *pb.GateReq) *pb.GateResp {
	var (
		gate iface.IGate
	)

	//basic check
	if in == nil {
		return nil
	}
	if in.Address != "" {
		//get gate by address
		gate = c.getGateByAddr(in.Address)
	}else{
		//pick gate by service kind
		gate = c.getGateByKind(in.Service)
	}
	if gate == nil {
		return nil
	}
	//send general request
	return gate.SendGenReq(in)
}

//cast data to gate server
func (c *Client) CastData(
			address string,
			in *pb.ByteMessage,
		) bool {

	//get remote gate by address
	gate := c.getGateByAddr(address)
	if gate == nil {
		return false
	}

	//cast response to tcp client
	bRet := gate.CastData(in)

	return bRet
}

//cast data to one kind gates
func (c *Client) CastDataByKind(kind string, in *pb.ByteMessage) bool {
	if kind == "" || in == nil {
		return false
	}

	if c.gateMap == nil || len(c.gateMap) <= 0 {
		return false
	}

	//loop gate and cast
	for _, gate := range c.gateMap {
		if gate.GetKind() != kind {
			continue
		}
		gate.CastData(in)
	}

	return true
}

//cast data to all gate
func (c *Client) CastDataToAll(in *pb.ByteMessage) bool {
	if in == nil || c.gateMap == nil {
		return false
	}
	//loop gate and cast
	for _, gate := range c.gateMap {
		gate.CastData(in)
	}
	return true
}

////////////////
//private func
///////////////

//run main process
func (c *Client) runMainProcess() {
	var (
		ticker = time.NewTicker(time.Second * define.GateStatCheckRate)
	)

	//defer
	defer func() {
		if err := recover(); err != nil {
			log.Println("Client:runMainProcess panic, err:", err)
		}

		//clean up
		ticker.Stop()
		close(c.closeChan)
	}()

	//loop
	for {
		select {
		case <- ticker.C:
			{
				//check gate status
				c.checkGateStatus()
			}
		case <- c.closeChan:
			return
		}
	}
}

//get gate by address
func (c *Client) getGateByAddr(address string) iface.IGate {
	if address == "" {
		return nil
	}
	v, ok := c.gateMap[address]
	if !ok {
		return nil
	}
	return v
}

//pick rand gate by kind
func (c *Client) getGateByKind(kind string) iface.IGate {
	var (
		address string
	)

	//basic check
	if kind == "" || c.gateMap == nil {
		return nil
	}

	//pick address by kind
	for addr, v := range c.gateMap {
		if v.GetKind() == kind {
			address = addr
			break
		}
	}

	if address == "" {
		return nil
	}

	//get gate by address
	return c.getGateByAddr(address)
}

//check gate connect status process
//this used for check downed gate server
func (c *Client) checkGateStatus() bool {
	//basic check
	if c.gateMap == nil || len(c.gateMap) <= 0 {
		return false
	}

	//loop check
	for _, gate := range c.gateMap {
		if gate.ConnIsNil() {
			gate.Connect(true)
			continue
		}
		state := gate.GetConnStat()
		if state == "TRANSIENT_FAILURE" || state == "SHUTDOWN" {
			//gate down, try connect again
			gate.Connect(true)
		}
	}
	return true
}