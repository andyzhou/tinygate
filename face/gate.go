package face

import (
	"context"
	"fmt"
	"github.com/andyzhou/tinygate/define"
	"github.com/andyzhou/tinygate/json"
	pb "github.com/andyzhou/tinygate/proto"
	"google.golang.org/grpc"
	"io"
	"log"
	"sync"
	"time"
)

/*
 * gate face
 *
 * - inter face for client communicate with gate server
 * - support node detect, data receive and cast
 * - tcp service will be client side
 */

//gate info
type Gate struct {
	kind string //service kind
	tags []string
	address string //remote server address, host:port
	conn *grpc.ClientConn //rpc client connect
	client pb.GateServiceClient //service client
	stream pb.GateService_BindStreamClient //stream client
	ctx context.Context
	reqChan chan pb.ByteMessage
	closeChan chan bool
	needQuit bool
	sync.RWMutex
	//cb func
	cbForStreamReceived func(from string, in *pb.ByteMessage) bool //call back for received data
	cbForGateServerDown func(kind, addr string) bool //call back for gate server down
	cbForGateServerUp func(kind, addr string) bool //call back for gate server up
}

//construct
func NewGate(
			serviceKind,
			serverHost string,
			serverPort int,
			tags ... string,
		) *Gate {
	//self init
	this := &Gate{
		kind:serviceKind,
		tags:tags,
		address:fmt.Sprintf("%s:%d", serverHost, serverPort),
		ctx:context.Background(),
		reqChan:make(chan pb.ByteMessage, define.GateReqChanSize),
		closeChan:make(chan bool, 1),
	}

	//inter init
	this.interInit()

	//spawn main process
	go this.runMainProcess()

	return this
}

/////////
//api
////////

//quit
func (c *Gate) Quit() {
	//try catch panic
	defer func() {
		if err := recover(); err != nil {
			log.Println("Gate:Quit panic, err:", err)
		}
	}()

	//send close to chan
	c.Lock()
	defer c.Unlock()
	c.needQuit = true
	c.closeChan <- true
}

//cast data to server with stream mode
func (c *Gate) CastData(in *pb.ByteMessage) (bRet bool) {
	//basic check
	if in == nil || in.MessageId < 0 || in.Data == nil {
		return
	}

	//try catch panic
	defer func() {
		if err := recover(); err != nil {
			log.Println("Gate::CastData panic, err:", err)
			bRet = false
			return
		}
	}()

	//send request
	c.reqChan <- *in
	bRet = true
	return
}

//check connect is nil or not
func (c *Gate) ConnIsNil() bool {
	if c.conn == nil {
		return true
	}
	return false
}

//connect gate server
func (c *Gate) Connect(isReConn bool) bool {
	return c.connect(isReConn)
}

//get kind
func (c *Gate) GetKind() string {
	return c.kind
}

//get tag
func (c *Gate) GetTags() []string {
	return c.tags
}

//get connect state
func (c *Gate) GetConnStat()string {
	return c.conn.GetState().String()
}

//send general request to gate server
//this is sync request
func (c *Gate) SendGenReq(in *pb.GateReq) *pb.GateResp {
	if in == nil {
		return nil
	}
	resp, err := c.client.GenReq(context.Background(), in)
	if err != nil {
		return nil
	}
	return resp
}

//set cb for receive data for server with stream mode
func (c *Gate) SetCBForStreamReceived(
					cb func(from string, in *pb.ByteMessage) bool,
				) bool {
	if cb == nil || c.cbForStreamReceived != nil {
		return false
	}
	c.cbForStreamReceived = cb
	return true
}

//set cb for gate server down
func (c *Gate) SetCBForGateServerDown(
				cb func(string, string) bool,
			) bool {
	if cb == nil || c.cbForGateServerDown != nil {
		return false
	}
	c.cbForGateServerDown = cb
	return true
}

//set cb for gate server up
func (c *Gate) SetCBForGateServerUp(
				cb func(string, string) bool,
			) bool {
	if cb == nil || c.cbForGateServerUp != nil {
		return false
	}
	c.cbForGateServerUp = cb
	return true
}

///////////////
//private func
///////////////

//cast data to gate server pass stream mode
func (c *Gate) castData(in *pb.ByteMessage) bool {
	//basic check
	if in == nil || c.stream == nil {
		return false
	}

	//send data pass stream mode
	err := c.stream.Send(in)
	if err != nil {
		log.Println("Gate::castData failed, err:", err.Error())
		//try reconnect
		return false
	}
	return true
}

//receive stream data from gate server
//if set cb, will call the cb for received stream data
func (c *Gate) receiveGateStream() {
	var (
		in *pb.ByteMessage
		err error
	)

	//basic check
	if c.stream == nil || c.cbForStreamReceived == nil {
		return
	}

	//loop receive
	for {
		in, err = c.stream.Recv()
		if err == io.EOF {
			log.Println("Gate::receiveGateStream, gate data EOF")
			continue
		}
		if err != nil {
			log.Println("Gate::receiveGateStream, Receive gate data failed, " +
						"err:", err.Error())
			//gate server down, call the relate cb func to notify client side
			if c.cbForGateServerDown != nil {
				c.cbForGateServerDown(c.kind, c.address)
			}
			break
		}

		//call cb for cast gate data to current service node
		if c.cbForStreamReceived != nil {
			c.cbForStreamReceived(c.address, in)
		}
	}

	//lost connect, try reconnect
	if !c.needQuit {
		go c.connect(true)
	}
}

//notify current node to gate server
func (c *Gate) notifyServer() bool {
	//init node json
	nodeJson := json.NewNodeJson()
	nodeJson.Kind = c.kind

	//init byte message
	byteMessage := pb.ByteMessage{
		MessageId:define.MessageIdOfNodeUp,
		Data:nodeJson.Encode(),
	}

	//send to gate server
	if c.stream == nil {
		return false
	}

	err := c.stream.Send(&byteMessage)
	if err != nil {
		log.Println("Gate::notifyServer failed, err:", err.Error())
		return false
	}

	return true
}

//connect gate server
func (c *Gate) connect(isReConn bool) bool {
	var (
		stream pb.GateService_BindStreamClient
		err error
	)

	//release resource for reconnect
	if isReConn {
		//release old connect
		if c.conn != nil {
			c.Lock()
			c.conn.Close()
			c.conn = nil
			c.Unlock()
		}
	}

	//try connect gate server
	conn, err := grpc.Dial(
		c.address,
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Println("Gate::interInit, can't reconnect gate, err:", err.Error())
		return false
	}

	//reset client & conn
	client := pb.NewGateServiceClient(conn)
	if client == nil {
		log.Println("Gate::interInit, init stream failed")
		return false
	}

	//try create stream of both side
	tryTimes := 0
	for {
		stream, err = client.BindStream(c.ctx)
		if err == nil {
			break
		}
		if err != nil && tryTimes >= define.GateBindTryTimes {
			//too many errors, need break
			return false
		}
		if c.needQuit {
			return false
		}
		tryTimes++
		time.Sleep(time.Second)
	}

	//connect gate server succeed
	if c.cbForGateServerUp != nil {
		c.cbForGateServerUp(c.kind, c.address)
	}

	//sync gate property
	c.Lock()
	c.conn = conn
	c.stream = stream
	c.client = client
	c.Unlock()

	//notify gate server
	c.notifyServer()

	//spawn new process for receive stream data
	go c.receiveGateStream()

	return true
}

//run main process
func (c *Gate) runMainProcess() {
	var (
		req pb.ByteMessage
		needQuit, isOk bool
	)

	//defer
	defer func() {
		if err := recover(); err != nil {
			log.Println("Gate:runMainProcess panic, err:", err)
		}
		//close chan
		close(c.reqChan)
		close(c.closeChan)
	}()

	//loop
	for {
		if needQuit {
			break
		}
		select {
		case req, isOk = <- c.reqChan://cast data to gate server
			if isOk {
				c.castData(&req)
			}
		case <- c.closeChan:
			needQuit = true
		}
	}
}

//inter init
func (c *Gate) interInit() {
	//connect gate server
	go c.connect(false)
}