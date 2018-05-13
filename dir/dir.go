package dir

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"syscall"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/mildred/renametree/index"
	"github.com/mildred/renametree/uuid_source"
)

type Options struct {
	// if true, associate files that have new inodes to old files in the same
	// location. This can be used to continue detect renames with files that have
	// been replaced entirely up to the inode.
	AssociateChangedInodes bool

	// Always generate UUID if a file doesn't have one. If no uuid is generated,
	// the file is not tracked
	AlwaysGenerateUuid bool
}

var DefaultOptions Options = Options{
	AssociateChangedInodes: false,
	AlwaysGenerateUuid:     true,
}

type Dir struct {
	startTime time.Time
	path      string
	inum      uint64
	indexpath string
	Index     *index.Index
	Options   Options
	Log       *log.Logger
}

func (d *Dir) Path() string {
	return d.path
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
		Options:   DefaultOptions,
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
	if err != nil && os.IsNotExist(err) {
		d.Index = new(index.Index)
		return nil
	} else if err != nil {
		return fmt.Errorf("cannot open %s, %s", d.indexpath, err)
	}
	defer f.Close()
	d.Index = new(index.Index)
	err = json.NewDecoder(f).Decode(d.Index)
	if err != nil {
		return fmt.Errorf("cannot parse %s, %s", d.indexpath, err)
	}
	return nil
}

func (d *Dir) saveindex() error {
	//d.Log.Printf("save index %s", d.indexpath)
	//json.NewEncoder(os.Stderr).Encode(d.Index)
	f, err := os.Create(d.indexpath + ".saving")
	if err != nil {
		return err
	}
	func() {
		defer f.Close()
		enc := json.NewEncoder(f)
		enc.SetIndent("", "\t")
		err = enc.Encode(d.Index)
	}()
	if err != nil {
		return err
	}
	defer func() { d.Index.Dirty = false }()
	return os.Rename(d.indexpath+".saving", d.indexpath)
}

func (d *Dir) Update(src uuidsource.Interface) (err error) {
	if src == nil {
		src = uuidsource.Null
	}
	defer func() {
		if d.Index.Dirty {
			e := d.saveindex()
			if e != nil {
				err = multierror.Append(err, e).ErrorOrNil()
			}
		}
	}()
	return d.updateDirContents(src, "")
}

func (d *Dir) updateDirContents(src uuidsource.Interface, prefix string) error {
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
			err := d.updateFile(src, path.Join(prefix, name), ent)
			if err != nil {
				errs = multierror.Append(errs, err)
			}
			if ent.IsDir() {
				err := d.updateDirContents(src, path.Join(prefix, name))
				if err != nil {
					errs = multierror.Append(errs, err)
				}
			}
		}
	}
	return errs.ErrorOrNil()
}

func (d *Dir) updateFile(src uuidsource.Interface, prefix string, st os.FileInfo) error {
	inum := getinum(st)
	file := d.Index.GetFileByInum(inum)
	if uuid := src.GetUuidByPath(prefix); uuid != "" {
		file = d.Index.AddFile(inum, uuid)
	}
	if d.Options.AssociateChangedInodes && file == nil {
		path := d.Index.GetPathByLastPath(prefix)
		if path != nil {
			file = d.Index.AddFile(inum, file.Uuid)
		}
	}
	if d.Options.AlwaysGenerateUuid && file == nil {
		uuid, err := d.genUuid(prefix, st)
		if err != nil {
			return err
		}
		file = d.Index.AddFile(inum, uuid)
	}
	if file != nil {
		path := d.Index.GetLastPathByUuid(file.Uuid)
		if path == nil || path.Path != prefix {
			d.Index.AddPathToUuid(file.Uuid, prefix, uint64(d.startTime.Unix()))
		}
	}
	return nil
}

func (d *Dir) genUuid(prefix string, st os.FileInfo) (string, error) {
	hash := sha1.New()
	pathname := path.Join(d.path, prefix)
	fmt.Fprintf(hash, "%d\000%s\000", d.startTime.Unix(), prefix)
	if !st.IsDir() {
		f, err := os.Open(pathname)
		if err != nil {
			return "", fmt.Errorf("cannot open %s, %v", pathname, err)
		}
		defer f.Close()
		_, err = io.Copy(hash, f)
		if err != nil {
			return "", fmt.Errorf("read %s, %s", pathname, err)
		}
	}
	return hex.EncodeToString(hash.Sum([]byte{})), nil
}
