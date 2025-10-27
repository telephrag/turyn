package main

import (
	"flag"
	"log"
	"os"
	"runtime"
	"syscall"
	"turyn/turyn"
)

func init() { log.SetFlags(0) }

func main() {
	wd := flag.String("d", ".", "Working directory with files to join.")
	flag.Parse()

	_, err := os.Stat(*wd)
	if err != nil {
		log.Fatalln("Failed to stat rootpath.")
	}

	t := turyn.New()

	chain := []turyn.Middleware{t.CheckIfDir, t.CollectPathSize}

	if err := t.Gather(*wd, t.Chain(chain...)); err != nil {
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

	t.ProcessAtomicWait(fout, 0, runtime.NumCPU())
}
