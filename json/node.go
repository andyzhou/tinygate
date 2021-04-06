package json

/*
 * json for node
 * - inter used for node notify by sub service
 * - this be auto send from client api
 */

//json info
type NodeJson struct {
	Kind string `json:"kind"`
	Tag string `json:"tag"` //used for unique of one kind
	BaseJson
}

/////////////////////////////
//construct for NodeJson
/////////////////////////////

//construct
func NewNodeJson() *NodeJson {
	this := &NodeJson{}
	return this
}

//encode json data
func (j *NodeJson) Encode() []byte {
	return j.BaseJson.Encode(j)
}

//decode json data
func (j *NodeJson) Decode(data []byte) bool {
	return j.BaseJson.Decode(data, j)
}
