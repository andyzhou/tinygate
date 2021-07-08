This is a gate service library, bridge the tcp/websocket and sub other service.

# feature
 - rpc gate bridge interface
 - used for tcp4/6, websocket, http, etc.
 - use rpc stream mode for performance
 - support general sync request
 
# api

 - service.go for gate server side
 - client.go for gate client side 
 
# how gen proto

cd proto
protoc --go_out=plugins=grpc:. *.proto
 
# how to use？

- see code under the `example` dir
 
 # tips
 - support gen and stream mode
 - if use for tcp service, tcp server will be client side
 - sub service will be server side, receive request from tcp server,
   and send resp data to tcp server.