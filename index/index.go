package index

type Index struct {
	Files []*File
	Paths []*Path
	Dirty bool `json:"-"`
}

type File struct {
	Inum uint64
	Uuid string
}

type Path struct {
	Uuid string
	Time uint64
	Path string
}

func (idx *Index) GetFileByInum(inum uint64) *File {
	for _, file := range idx.Files {
		if file.Inum == inum {
			return file
		}
	}
	return nil
}

func (idx *Index) GetLastPathByUuid(uuid string) *Path {
	var lastPath *Path
	for _, path := range idx.Paths {
		if path.Uuid == uuid {
			if lastPath == nil || lastPath.Time < path.Time {
				lastPath = path
			}
		}
	}
	return lastPath
}

func (idx *Index) AddPathToUuid(uuid string, path string, time uint64) {
	idx.Dirty = true
	idx.Paths = append(idx.Paths, &Path{
		Uuid: uuid,
		Path: path,
		Time: time,
	})
}

func (idx *Index) AddFile(inum uint64, uuid string) *File {
	idx.Dirty = true
	f := &File{
		Inum: inum,
		Uuid: uuid,
	}
	idx.Files = append(idx.Files, f)
	return f
}
