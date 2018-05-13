package renames

import (
	"fmt"
	"sort"

	"github.com/hashicorp/go-multierror"
	"github.com/mildred/renametree/dir"
	"github.com/mildred/renametree/index"
)

type Renames struct {
	dirA *dir.Dir
	dirB *dir.Dir
}

func New(dirA, dirB *dir.Dir) *Renames {
	return &Renames{
		dirA: dirA,
		dirB: dirB,
	}
}

type intersecting struct {
	A PathHistory
	B PathHistory
}

type commonHistory struct {
	A PathHistory
	B PathHistory
}

func newCommonHistory(a, b PathHistory) *commonHistory {
	ab := new(commonHistory)
	found := false
	for i := len(a) - 1; i >= 0 && !found; i-- {
		for j := len(b) - 1; j >= 0 && !found; j-- {
			if a[i].Time == b[j].Time && a[i].Path == b[j].Path {
				if i+1 < len(a) {
					ab.A = a[i+1:]
				}
				if j+1 < len(b) {
					ab.B = b[j+1:]
				}
				found = true
			}
		}
	}
	if !found {
		ab.A = a
		ab.B = b
	}
	for len(ab.A) > 0 && len(ab.B) > 0 && ab.A[0].Path == ab.B[0].Path {
		ab.A = ab.A[1:]
		ab.B = ab.B[1:]
	}
	return ab
}

func (ab *commonHistory) AisOlder() bool {
	return len(ab.A) == 0 && len(ab.B) > 0
}

func (ab *commonHistory) BisOlder() bool {
	return (&commonHistory{A: ab.B, B: ab.A}).AisOlder()
}

type PathHistory []*index.Path

func (s PathHistory) Len() int      { return len(s) }
func (s PathHistory) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (p PathHistory) LastPath() string {
	if len(p) == 0 {
		panic("LastPath() on an empty path history")
	}
	return p[len(p)-1].Path
}

type ByTime struct{ PathHistory }

func (s ByTime) Less(i, j int) bool { return s.PathHistory[i].Time < s.PathHistory[j].Time }

func (r *Renames) findIntersectingUuid() map[string]intersecting {
	set := map[string]intersecting{}
	res := map[string]intersecting{}
	for _, p := range r.dirA.Index.Paths {
		ab := set[p.Uuid]
		ab.A = append(ab.A, p)
		set[p.Uuid] = ab
	}
	for _, p := range r.dirB.Index.Paths {
		ab := set[p.Uuid]
		ab.B = append(ab.B, p)
		set[p.Uuid] = ab
	}
	for uuid, ab := range set {
		if len(ab.A) > 0 && len(ab.B) > 0 {
			sort.Sort(ByTime{ab.A})
			sort.Sort(ByTime{ab.B})
			res[uuid] = ab
		}
	}
	return res
}

func (r *Renames) Rename() (err error) {
	for uuid, ab := range r.findIntersectingUuid() {
		if ab.A.LastPath() == ab.B.LastPath() {
			continue
		} else {
			hist := newCommonHistory(ab.A, ab.B)
			if hist.AisOlder() {
				fmt.Printf("rename in %s %s to %s\n", r.dirA.Path(), ab.A.LastPath(), ab.B.LastPath())
			} else if hist.BisOlder() {
				fmt.Printf("rename in %s %s to %s\n", r.dirB.Path(), ab.B.LastPath(), ab.A.LastPath())
			} else {
				err = multierror.Append(err, fmt.Errorf("Conflict for file %s\n\t%s\n\t%s", uuid, ab.A.LastPath(), ab.B.LastPath()))
			}
		}
	}
	return
}

func (r *Renames) FindConflicts() (err error) {
	return nil
}
