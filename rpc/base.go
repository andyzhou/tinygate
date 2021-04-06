package rpc

import (
	"context"
	"google.golang.org/grpc/stats"
)

/*
 * Basic function face
 */

//connect ctx key info
type ConnCtxKey struct{}

//basic face info
type Base struct {}

//get connect tag from context
func (b *Base) GetConnTagFromContext(
					ctx context.Context,
				) (*stats.ConnTagInfo, bool) {
	tag, ok := ctx.Value(ConnCtxKey{}).(*stats.ConnTagInfo)
	return tag, ok
}

