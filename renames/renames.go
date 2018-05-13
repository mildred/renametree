package renames

import (
	"github.com/mildred/renametree/dir"
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

func (r *Renames) Rename() error {
	return nil
}
