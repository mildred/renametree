package dir

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"syscall"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/mildred/renametree/index"
)

type Dir struct {
	startTime time.Time
	path      string
	inum      uint64
	indexpath string
	index     *index.Index
}

func getinum(st0 os.FileInfo) uint64 {
	st := st0.Sys().(*syscall.Stat_t)
	if st == nil {
		panic("getinum() failed while processing stat_t results")
	}
	return st.Ino
}

func Open(t time.Time, path string) (*Dir, error) {
	d := &Dir{
		startTime: t,
		path:      path,
	}
	return d, d.openIndex()
}

func (d *Dir) openIndex() error {
	st, err := os.Stat(d.path)
	if err != nil {
		return err
	}
	d.inum = getinum(st)
	d.indexpath = path.Join(d.path, fmt.Sprintf(".renametree-%d.v0.idx", d.inum))
	return d.readindex()
}

func (d *Dir) readindex() error {
	f, err := os.Open(d.indexpath)
	if err == nil {
		return err
	}
	defer f.Close()
	d.index = new(index.Index)
	return json.NewDecoder(f).Decode(d.index)
}

func (d *Dir) saveindex() error {
	f, err := os.Create(d.indexpath + ".saving")
	if err == nil {
		return err
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(d.index)
	if err != nil {
		return err
	}
	defer func() { d.index.Dirty = false }()
	return os.Rename(d.indexpath+".saving", d.indexpath)
}

func (d *Dir) Update() (err error) {
	defer func() {
		if d.index.Dirty {
			e := d.saveindex()
			if e != nil {
				err = multierror.Append(err, e).ErrorOrNil()
			}
		}
	}()
	return d.updateDirContents("")
}

func (d *Dir) updateDirContents(prefix string) error {
	pathname := path.Join(d.path, prefix)
	fd, err := os.Open(pathname)
	if err != nil {
		return err
	}
	names, err := fd.Readdirnames(-1)
	if err != nil {
		return err
	}
	var errs *multierror.Error
	for _, name := range names {
		ent, err := os.Lstat(path.Join(pathname, name))
		if err != nil {
			errs = multierror.Append(errs, err)
		} else {
			err := d.updateFile(path.Join(prefix, name), ent)
			if err != nil {
				errs = multierror.Append(errs, err)
			}
			if ent.IsDir() {
				err := d.updateDirContents(path.Join(prefix, name))
				if err != nil {
					errs = multierror.Append(errs, err)
				}
			}
		}
	}
	return errs.ErrorOrNil()
}

func (d *Dir) updateFile(prefix string, st os.FileInfo) error {
	inum := getinum(st)
	file := d.index.GetFileByInum(inum)
	if file == nil {
		uuid, err := d.genUuid(prefix, st)
		if err != nil {
			return err
		}
		file = d.index.AddFile(inum, uuid)
	}
	path := d.index.GetLastPathByUuid(file.Uuid)
	if path == nil || path.Path != prefix {
		d.index.AddPathToUuid(file.Uuid, prefix, uint64(d.startTime.Unix()))
	}
	return nil
}

func (d *Dir) genUuid(prefix string, st os.FileInfo) (string, error) {
	hash := sha512.New()
	fmt.Fprintf(hash, "%d\000%s\000", getinum(st), prefix)
	f, err := os.Open(st.Name())
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = io.Copy(hash, f)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum([]byte{})), nil
}
