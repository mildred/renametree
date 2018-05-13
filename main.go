package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/mildred/renametree/dir"
	"github.com/mildred/renametree/renames"
)

func main() {
	l := log.New(os.Stdout, "", log.LstdFlags)
	var dryRun bool = false

	flag.BoolVar(&dryRun, "n", dryRun, "Dry run")

	flag.Parse()
	dirs := flag.Args()
	t := time.Now()

	dirA, err := dir.Open(t, dirs[0])
	if err != nil {
		log.Fatalf("Cannot open %s: %s", dirs[0], err)
	}
	dirA.Log = l

	dirB, err := dir.Open(t, dirs[1])
	if err != nil {
		log.Fatalf("Cannot open %s: %s", dirs[1], err)
	}
	dirB.Log = l

	errA := dirA.Update(dirB.Index)
	errB := dirB.Update(dirA.Index)

	err = multierror.Append(errA, errB).ErrorOrNil()
	if err != nil {
		log.Fatalf("Updating index files: %s", err)
	}

	renames := renames.New(dirA, dirB)
	renames.Log = l
	renames.Options.DryRun = dryRun
	err = renames.Rename()
	if err != nil {
		log.Fatalf("Detecting renames: %s", err)
	}

	err = renames.FindConflicts()
	if err != nil {
		log.Fatal(err)
	}
}
