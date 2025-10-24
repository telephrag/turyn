package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"
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

	t := turyn.New(wd)

	chain := []turyn.Middleware{t.Default}

	if err := t.Walk(wd, t.Chain(chain...)); err != nil {
		log.Fatalln(err)
	}
	if fc := t.GetFileCount(); fc == 0 {
		log.Fatalln("Failed to walk filetree, aborting...")
	}

	fout, err := os.OpenFile("out.tur", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalln(err)
	}

	files, sizes := t.Prepare()
	totalSize := getTotalSize(files, sizes)

	fmt.Printf("Total: %d\n", totalSize)

	// TODO: use truncate in case we are overwriting existing file?
	err = syscall.Fallocate(int(fout.Fd()), 0, 0, totalSize)

	ncpu := runtime.NumCPU()
	// ncpu = 1
	var writersRunning atomic.Int32
	var writeOffset int64
	os.Chdir(wd)
	for i := range files {
		for writersRunning.Load() >= int32(ncpu) {
		}

		writersRunning.Add(1)
		go func(idx int, off int64) {
			defer writersRunning.Add(-1)

			r, err := os.OpenFile(files[idx], os.O_RDONLY, 0644)
			if err != nil {
				log.Fatalln(err) // TODO: let other goroutines to finish writing
			}
			defer r.Close()

			w := io.NewOffsetWriter(fout, off)

			// not exposing user's filetree
			fmt.Fprintf(w, "||| %s\n", files[idx])
			_, err = io.Copy(w, r)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Fprint(w, "\n\n")

			fmt.Printf("%10d %s\n", off, files[idx])

		}(i, writeOffset)

		writeOffset += (sizes[i] + 7 + int64(len(files[i])))
	}

	for writersRunning.Load() > 0 {
		time.Sleep(time.Millisecond)
	}

}

func getTotalSize(files []string, sizes []int64) int64 {
	var totalSize int64 = 0
	for i := range sizes {
		// ||| + " " + filename + \n + filesize + \n\n (after contents)
		totalSize += (7 + int64(len(files[i])) + sizes[i])
	}
	return totalSize
}
