package define

//config file
const (
	ConfMain = "gate.conf"
	ConfNode = "node.conf"
)

//property name of tcp connection
const (
	TcpPropertyPlayerId = "playerId"
)

//bind opt
const (
	NodeOptBind = 1
	NodeOptUnbind = 2
)

//node rule kind
const (
	NodeRuleOfHash = iota
	NodeRuleOfPersistent
)

//client relate
const (
	ClientReqChanSize = 2048
	ClientCastChanSize = 2048
	LazyCastChanSize = 4096
	ClientActiveCheckRate = 30
)