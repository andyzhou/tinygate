package json

/*
 * json for bind or unbind
 * - bind player and tcp connect id
 * - bind gate and sub service tag
 * - this notify from rpc client
 */

 //json info
 type BindJson struct {
 	Opt int `json:"opt"` //1:bind 2:unbind
 	ConnId uint32 `json:"connId"`
 	PlayerId int64 `json:"playerId"`
 	Nodes map[string]string `json:"nodes"` //nodeKind -> nodeTag
 	BaseJson
 }

/////////////////////////////
//construct for BindJson
/////////////////////////////

//construct
func NewBindJson() *BindJson {
	this := &BindJson{
		Nodes:make(map[string]string),
	}
	return this
}

//encode BindJson data
func (j *BindJson) Encode() []byte {
	return j.BaseJson.Encode(j)
}

//decode json data
func (j *BindJson) Decode(data []byte) bool {
	return j.BaseJson.Decode(data, j)
}
