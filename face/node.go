package face

import (
	"github.com/andyzhou/gate/iface"
	"github.com/andyzhou/gate/json"
	pb "github.com/andyzhou/gate/proto"
	"sync"
)

/*
 * node face, implement of INode
 */

 //kind nodes
 type KindNodes struct {
 	nodes map[string]string `tag -> remoteAddr`
 }

 //face info
 type Node struct {
 	cbForNodeDown func(serviceKind, remoteAddr string) bool
 	serviceMap map[string]iface.IService `remoteAddr -> IService`
 	kindMap map[string]*KindNodes `active node map, kind -> KindNodes`
 	sync.RWMutex
 }

 //construct
func NewNode() *Node {
	//self init
	this := &Node{
		serviceMap:make(map[string]iface.IService),
		kindMap:make(map[string]*KindNodes),
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

//pick one active node
func (f *Node) PickNode(
					kind string,
				) string {
	var (
		tag string
	)
	//get kind nodes
	kindNodes := f.getAllByKind(kind)
	if kindNodes == nil || kindNodes.nodes == nil {
		return tag
	}
	//pick random one
	for tmpTag, _ := range kindNodes.nodes {
		return tmpTag
	}
	return tag
}

//get service by node tag
func (f *Node) GetServiceByTag(
					kind, tag string,
				) iface.IService {
	//basic check
	if kind == "" || tag == "" {
		return nil
	}

	//get kind nodes
	kindNodes, ok := f.kindMap[kind]
	if !ok {
		return nil
	}

	//get remote address by tag
	remoteAddr, ok := kindNodes.nodes[tag]
	if remoteAddr == "" {
		return nil
	}

	//get service by remote address
	return f.GetService(remoteAddr)
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

//get kind by remote address
func (f *Node) GetKind(
					address string,
				) string {
	var (
		kind string
	)
	if address == "" {
		return kind
	}
	v, ok := f.serviceMap[address]
	if !ok {
		return kind
	}
	return v.GetKind()
}

//rpc client node down
//means sub server node down
func (f *Node) NodeDown(
					remoteAddress string,
				) bool {
	if remoteAddress == "" {
		return false
	}

	//get kind of remote address
	service, ok := f.serviceMap[remoteAddress]
	if !ok {
		return false
	}

	//notify outside
	if f.cbForNodeDown != nil {
		f.cbForNodeDown(service.GetKind(), remoteAddress)
	}

	//get all nodes of one kind
	kindNodes := f.getAllByKind(service.GetKind())

	//get service tag
	tag := service.GetTag()

	//remove with locker
	f.Lock()
	defer f.Unlock()
	delete(f.serviceMap, remoteAddress)
	if kindNodes != nil {
		delete(kindNodes.nodes, tag)
	}
	return true
}

//rpc client node up to service
func (f *Node) NodeUp(
					remoteAddress string,
					jsonObj *json.NodeJson,
					stream *pb.GateService_BindStreamServer,
				) bool {
	var (
		isNew bool
	)

	//basic check
	if remoteAddress == "" || jsonObj == nil || stream == nil {
		return false
	}

	//get key data
	kind := jsonObj.Kind
	tag := jsonObj.Tag

	//get all nodes of kind
	kindNodes := f.getAllByKind(kind)
	if kindNodes == nil {
		kindNodes = &KindNodes{
			nodes:make(map[string]string),
		}
		isNew = true
	}

	//sync current node info
	service := NewService(
					kind,
					tag,
					remoteAddress,
					stream,
				)

	//add into map with locker
	f.Lock()
	defer f.Unlock()

	f.serviceMap[remoteAddress] = service
	kindNodes.nodes[tag] = remoteAddress
	if isNew {
		//sync kind nodes
		f.kindMap[kind] = kindNodes
	}

	return true
}


//set cb for sub service node down
func (f *Node) SetCBForNodeDown(cb func(serviceKind, remoteAddr string) bool) bool {
	if cb == nil {
		return false
	}
	f.cbForNodeDown = cb
	return true
}

///////////////
//private func
///////////////

//get all nodes by kind
func (f *Node) getAllByKind(kind string) *KindNodes {
	//basic check
	if kind == "" || f.kindMap == nil {
		return nil
	}
	v, ok := f.kindMap[kind]
	if !ok {
		return nil
	}
	return v
}