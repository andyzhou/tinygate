package define

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
	LazyCastChanSize = 4096
)