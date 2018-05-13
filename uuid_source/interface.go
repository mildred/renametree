package uuidsource

type Interface interface {
	GetUuidByPath(path string) string
}
