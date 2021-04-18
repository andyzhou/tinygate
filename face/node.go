package face

import (
	"github.com/andyzhou/gate/iface"
	"github.com/andyzhou/gate/json"
	pb "github.com/andyzhou/gate/proto"
	"sync"
)

/*
 * node face, implement of INode
 * - used for gate server side
 * - manage client nodes
 * - use IService as inter bridge between client and service
 */

 //face info
 type Node struct {
 	cbForClientNodeDown func(remoteAddr string) bool
 	serviceMap map[string]iface.IService //client service map, remoteAddr -> IService
 	clientNodes map[string]int64 //active client node map, remoteAddr -> activeTime`
 	sync.RWMutex
 }

 //construct
func NewNode() *Node {
	//self init
	this := &Node{
		serviceMap:make(map[string]iface.IService),
		clientNodes:make(map[string]int64),
	}

	return this
}

//////////////////////
//implement of INode
//////////////////////

//quit
func (f *Node) Quit() {
	if f.serviceMap != nil {
		for _, service := range f.serviceMap {
			service.Quit()
		}
	}
}

//get remote node service
func (f *Node) GetService(
					address string,
				) iface.IService {
	if address == "" {
		return nil
	}
	service, ok := f.serviceMap[address]
	if !ok {
		return nil
	}
	return service
}

//get all service
func (f *Node) GetAllService() map[string]iface.IService {
	return f.serviceMap
}

//rpc client node down
//means remote client node down
func (f *Node) ClientNodeDown(
					remoteAddress string,
				) bool {
	if remoteAddress == "" {
		return false
	}

	//notify outside
	if f.cbForClientNodeDown != nil {
		f.cbForClientNodeDown(remoteAddress)
	}

	//remove with locker
	f.Lock()
	defer f.Unlock()
	delete(f.clientNodes, remoteAddress)
	return true
}

//rpc client node up
//means remote client node up
func (f *Node) ClientNodeUp(
					remoteAddress string,
					jsonObj *json.NodeJson,
					stream *pb.GateService_BindStreamServer,
				) bool {
	//basic check
	if remoteAddress == "" || stream == nil {
		return false
	}

	//check old
	_, ok := f.serviceMap[remoteAddress]
	if ok {
		return true
	}

	//init new service for current node
	service := NewService(
					remoteAddress,
					stream,
				)

	//add into map with locker
	f.Lock()
	defer f.Unlock()
	f.serviceMap[remoteAddress] = service
	return true
}


//set cb for client node down
func (f *Node) SetCBForClientNodeDown(cb func(remoteAddr string) bool) bool {
	if cb == nil {
		return false
	}
	f.cbForClientNodeDown = cb
	return true
}