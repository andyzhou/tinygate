package define

/*
 * tcp data message id
 */

 //inter message id, 0 ~ 20
 const (
 	MessageIdOfNodeUp = iota //node up
 	MessageIdOfBindOrUnbind //player node bind or unbind
	 MessageIdOfHeartBeat
 	MessageIdOfClientClosed //tcp client disconnect
 )