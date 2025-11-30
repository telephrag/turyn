package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"turyn/turyn"
)

func init() { log.SetFlags(0) }

func main() {
	wd := flag.String("wd", ".", "Working directory with files to join.")
	outf := flag.String("o", "out.tur", "File to output to.")
	ow := flag.Bool("ow", false, `Weather to overwrite output file 
		if it already exists. Program fails if false and file does exist.`)
	flag.Parse()

	if _, err := os.Stat(*wd); err != nil { // failing gracefully on bad path
		log.Fatalf("Failed to stat working directory: %s\n", *wd)
	}

	ofp := filepath.Dir(*outf)
	if _, err := os.Stat(ofp); err != nil {
		log.Fatalf("Can't stat path %s to output file: %s\n", ofp, err)
	}

	ofi, err := os.Stat(*outf)
	if err == nil {
		if ofi.IsDir() {
			log.Fatalf("Can't write output: %s is a directory\n", *outf)
		}

		if !*ow {
			log.Fatalf("%s exists but -ow flag is not set.\n", *outf)
		}
	}

	t := turyn.New()

	chain := []turyn.Middleware{t.CheckIfDir, t.CollectPathSize}
	if err := t.Gather(*wd, t.Chain(chain...)); err != nil {
		log.Fatalln(err)
	}

	if *ow && ofi != nil { // inefficient, but it was your choice to use -ow
		t.ExcludeOutputFile(*outf)
	}

	fout, err := os.OpenFile(*outf, os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		log.Fatalln(err)
	}

	err = syscall.Fallocate(int(fout.Fd()), 0, 0, t.GetTotalSize())
	if err != nil {
		log.Fatalf("error allocating filespace for output: %s\n", err)
	}

	t.Process(fout, 0, runtime.NumCPU())
}
