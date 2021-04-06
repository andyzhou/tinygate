package define

/*
 * tcp data message id
 */

 //inter message id, 0 ~ 20
 const (
 	MessageIdOfNodeUp = iota //node up
 	MessageIdOfBindOrUnbind //player node bind or unbind
 	MessageIdOfClientClosed //tcp client disconnect
 )

 ////special message tag
 //const (
 //	MessageTagOfHeartBeat = "heartBeat"
 //	MessageTagOfLogin = "login"
 //)
