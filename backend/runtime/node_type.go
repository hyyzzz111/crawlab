package runtime

//go:generate enumer -type NodeType -linecomment -tags a
const (
	Master NodeType = iota
	Worker
)

type NodeType int

var nodeType = Worker

func SetNodeType(newType NodeType) {
	nodeType = newType
}
func GetNodeType() NodeType {
	return nodeType
}
