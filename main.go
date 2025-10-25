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
	_wd := flag.String("d", ".", "Working directory with files to join.")
	flag.Parse()

	_, err := os.Stat(*_wd)
	if err != nil {
		log.Fatalln("Failed to stat rootpath.")
	}

	wd, err := filepath.Abs(*_wd)
	if err != nil {
		log.Fatalln(err)
	}

	t := turyn.New()

	chain := []turyn.Middleware{t.Default}

	if err := t.Gather(wd, t.Chain(chain...)); err != nil {
		log.Fatalln(err)
	}

	fout, err := os.OpenFile("out.tur", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
	if err != nil {
		log.Fatalln(err)
	}

	// TODO: use truncate in case we are overwriting existing file?
	err = syscall.Fallocate(int(fout.Fd()), 0, 0, t.GetTotalSize())
	if err != nil {
		log.Fatalf("error allocating filespace for output: %s\n", err)
	}

	t.Process(fout, 0, runtime.NumCPU())
}
