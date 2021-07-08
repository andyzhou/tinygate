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

//others
const (
	GateReqChanSize = 1024 * 5
	GateBindTryTimes = 5
	GateStatCheckRate = 5 //xx seconds
	ResponseChanSize = 1024 * 5
)