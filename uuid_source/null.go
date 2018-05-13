package uuidsource

type NullSource struct{}

func (n *NullSource) GetUuidByPath(path string) string {
	return ""
}

var Null *NullSource = new(NullSource)
