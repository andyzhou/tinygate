package rpc

import (
	"context"
	"github.com/andyzhou/tinygate/iface"
	"google.golang.org/grpc/stats"
	"log"
)

/*
 * rpc stat handler
 * Need apply `TagConn`, `TagRPC`,
 * `HandleConn`, `HandleRPC` methods.
 */


type Stat struct {
	node iface.INode
	base *Base
}

//construct
func NewStat(node iface.INode) *Stat {
	this := &Stat{
		node: node,
		base:new(Base),
	}
	return this
}

func (h *Stat) TagConn(
					ctx context.Context,
					info *stats.ConnTagInfo,
				) context.Context {
	return context.WithValue(ctx, ConnCtxKey{}, info)
}

func (h *Stat) TagRPC(
					ctx context.Context,
					info *stats.RPCTagInfo,
				) context.Context {
	return ctx
}

//handle client node conn
func (h *Stat) HandleConn(ctx context.Context, s stats.ConnStats) {
	//get connect tag from context
	tag, ok := h.base.GetConnTagFromContext(ctx)
	if !ok {
		log.Fatal("Stat::HandleConn, can not get conn tag")
		return
	}
	//do relate opt by connect stat type
	switch s.(type) {
	case *stats.ConnBegin:
		//client node connect
		log.Printf("Stat::HandleConn, client node up, tag:%v", tag)
		//nodeFace.NodeUp(tag, tag.RemoteAddr.String())
	case *stats.ConnEnd:
		//client node down
		log.Printf("Stat::HandleConn, client node down, tag:%v", tag)
		if h.node != nil {
			h.node.ClientNodeDown(tag.RemoteAddr.String())
		}
	default:
		log.Printf("illegal ConnStats type\n")
	}
}

func (h *Stat) HandleRPC(ctx context.Context, s stats.RPCStats) {
	//fmt.Println("HandleRPC, IsClient:", s.IsClient())
}

