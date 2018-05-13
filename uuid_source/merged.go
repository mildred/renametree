package uuidsource

type Merged []Interface

func (m *Merged) GetUuidByPath(path string) string {
	for _, src := range *m {
		uuid := src.GetUuidByPath(path)
		if uuid != "" {
			return uuid
		}
	}
	return ""
}
