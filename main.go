package main

import (
	"flag"
	"log"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/mildred/renametree/dir"
	"github.com/mildred/renametree/renames"
)

func main() {
	flag.Parse()
	dirs := flag.Args()
	t := time.Now()

	dirA, err := dir.Open(t, dirs[0])
	if err != nil {
		log.Fatal(err)
	}

	dirB, err := dir.Open(t, dirs[1])
	if err != nil {
		log.Fatal(err)
	}

	errA := dirA.Update()
	errB := dirB.Update()

	err = multierror.Append(errA, errB).ErrorOrNil()
	if err != nil {
		log.Fatal(err)
	}

	renames := renames.New(dirA, dirB)
	err = renames.Rename()
	if err != nil {
		log.Fatal(err)
	}
}
