package gate

import (
	"github.com/andyzhou/gate/face"
	"github.com/andyzhou/gate/iface"
	"github.com/andyzhou/gate/json"
	pb "github.com/andyzhou/gate/proto"
)

/*
 * gate client
 *
 * - used at tcp server side
 * - communicate with gate server
 * - send request to gate server/sub service
 * - receive response from gate server/sub service
 */

//client info
type Client struct {
	client iface.IClient
}

//construct
//STEP-1
func NewClient() *Client {
	//self init
	this := &Client{
		client:face.NewClient(),
	}
	return this
}

//quit
func (c *Client) Quit() {
	c.client.Quit()
}

//set call back for received stream data from gate server
//STEP-2
func (c *Client) SetCBForStreamReceived(
			cb func(from string, in *pb.ByteMessage) bool,
		) bool {
	return c.client.SetCBForStreamReceived(cb)
}

//set call back for gate server down
//STEP-3
func (c *Client) SetCBForGateDown(
			cb func(kind, addr string) bool,
		) bool {
	return c.client.SetCBForGateDown(cb)
}

//add gate server
//support multi gates
//STEP-4
func (c *Client) AddGateServer(kind, host string, port int) bool {
	return c.client.AddGateServer(kind, host, port)
}

//set log option
//STEP-5, optional
func (c *Client) SetLog(dir, tag string) bool {
	return c.client.SetLog(dir, tag)
}

//bind batch node and tag for single client
func (c *Client) BindNodeTags(
			fromAddr string,
			bindJson *json.BindJson,
		) bool {
	return c.client.BindNodeTags(fromAddr, bindJson)
}

//send gen request
func (c *Client) SendGenReq(in *pb.GateReq) *pb.GateResp {
	return c.client.SendGenReq(in)
}

//cast stream data to one gate
func (c *Client) CastData(
			address string,
			in *pb.ByteMessage,
		) bool {
	return c.client.CastData(address, in)
}

func (c *Client) CastDataByTag(
			tag string,
			in *pb.ByteMessage,
		) bool {
	return c.client.CastDataByTag(tag, in)
}

//cast data to one kind gates
func (c *Client) CastDataToOneKind(kind string, in *pb.ByteMessage) bool {
	return c.client.CastDataToOneKind(kind, in)
}

//cast data to all gate
func (c *Client) CastDataToAll(in *pb.ByteMessage) bool {
	return c.client.CastDataToAll(in)
}