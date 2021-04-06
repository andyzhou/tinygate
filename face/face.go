package face

import (
	"github.com/andyzhou/gate/iface"
)

/*
 * inter face base
 */

//inter face base info
type InterFace struct {
	node iface.INode
}

//declare global variable
var RunInterFace *InterFace

//init
func init()  {
	RunInterFace = NewInterFace()
}

//construct
func NewInterFace() *InterFace {
	this := &InterFace{
		node:NewNode(),
	}
	return this
}

//quit
func (f *InterFace) Quit() {
	f.node.Quit()
}

//get relate face
func (f *InterFace) GetNodeFace() iface.INode {
	return f.node
}
