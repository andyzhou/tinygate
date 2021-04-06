package face

import (
	"fmt"
	"github.com/andyzhou/gate/define"
	"github.com/andyzhou/gate/iface"
	"github.com/andyzhou/gate/json"
	pb "github.com/andyzhou/gate/proto"
	"log"
	"sync"
	"time"
)

/*
 * gate client face
 *
 * - used by sub service
 * - api for gate face
 * - support multi gate server
 * - communicate byte data with gate server pass stream mode
 */

//inter macro define
const (
	GateStatCheckRate = 5 //xx seconds
)

//client info
type Client struct {
	kind string
	gateMap map[string]iface.IGate                                `running gate map, serverAddress -> Gate`
	cbForReceive func(from string, in *pb.ByteMessage) bool `call back for received data`
	cbForGateDown func(kind string, addr string) bool       `call back for gate server down`
	closeChan chan bool
	sync.Mutex `internal data locker`
}

//construct
//STEP-1
func NewClient(kind string) *Client {
	//self init
	this := &Client{
		kind:kind,
		gateMap:make(map[string]iface.IGate),
		closeChan:make(chan bool, 1),
	}

	//spawn main process
	go this.runMainProcess()

	return this
}

///////
//api
//////

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


//set call back for received stream data
//STEP-2
func (c *Client) SetCBForStream(
			cb func(from string, in *pb.ByteMessage) bool,
		) bool {
	if c.cbForReceive != nil {
		return false
	}
	c.cbForReceive = cb
	return true
}

//set call back for gate server down
//STEP-3
func (c *Client) SetCBForGateDown(
				cb func(kind, addr string) bool,
			) bool {
	if c.cbForGateDown != nil {
		return false
	}
	c.cbForGateDown = cb
	return true
}


//add gate server
//STEP-4
func (c *Client) AddGateServer(tag, host string, port int) bool {
	//basic check
	if tag == "" || host == "" || port <= 0 {
		return false
	}

	//format address
	address := fmt.Sprintf("%s:%d", host, port)

	//check gate has exists or not
	_, ok := c.gateMap[address]
	if ok {
		return true
	}

	//init gate
	gate := NewGate(c.kind, tag, host, port)

	//set callback function
	gate.SetCBForStreamReceive(c.cbForReceive)
	gate.SetCBForGateDown(c.cbForGateDown)

	//sync into map
	c.Lock()
	defer c.Unlock()
	c.gateMap[address] = gate

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


//bind batch node and tag for single client
func (c *Client) BindNodeTags(
			fromAddr string,
			bindJson *json.BindJson,
		) bool {
	//basic check
	if fromAddr == "" || bindJson == nil {
		return false
	}

	//set opt
	bindJson.Opt = define.NodeOptBind

	//init byte message
	in := &pb.ByteMessage{
		MessageId:define.MessageIdOfBindOrUnbind,
		ConnId:bindJson.ConnId,
		PlayerId:bindJson.PlayerId,
		Data:bindJson.Encode(),
	}

	//send to remote gate server
	bRet := c.CastData(fromAddr, in)

	return bRet
}

//cast data to one gate
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

func (c *Client) CastDataByTag(
			tag string,
			in *pb.ByteMessage,
		) bool {
	//get remote gate by address
	gate := c.getGateByAddr(tag)
	if gate == nil {
		return false
	}

	//cast response to tcp client
	bRet := gate.CastData(in)

	return bRet
}

//cast data to one kind gates
func (c *Client) CastDataToOneKind(kind string, in *pb.ByteMessage) bool {
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
		ticker = time.NewTicker(time.Second * GateStatCheckRate)
		needQuit bool
	)

	//defer
	defer func() {
		if err := recover(); err != nil {
			log.Println("Client:runMainProcess panic, err:", err)
		}

		ticker.Stop()
		//close chan
		close(c.closeChan)
	}()

	//loop
	for {
		if needQuit == true {
			break
		}
		select {
		case <- ticker.C:
			{
				//check gate status
				c.checkGateStatus()
			}
		case <- c.closeChan:
			needQuit = true
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

//get gate by tag
func (c *Client) getGateByTag(tag string) iface.IGate {
	if tag == "" {
		return nil
	}
	for _, gate := range c.gateMap {
		if gate.GetTag() == tag {
			return gate
		}
	}
	return nil
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