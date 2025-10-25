package turyn

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

type Turyn struct {
	files []string // absolute paths to files
	sizes []int64  // size of files for future preallocation
}

func New() *Turyn {
	return &Turyn{
		files: []string{},
		sizes: []int64{},
	}
}

// left to be for benchmarking later
func (t *Turyn) GatherFS(workDir string, wdf fs.WalkDirFunc) error {
	return fs.WalkDir(os.DirFS(workDir), ".", wdf)
}

func (t *Turyn) Gather(workDir string, wdf fs.WalkDirFunc) error {
	return filepath.WalkDir(workDir, wdf)
}

func (t *Turyn) Process(fout *os.File, baseOffset int64, ncpu int) {

	var writersRunning atomic.Int32
	var writeOffset int64
	for i := range t.files {
		// TODO: benchmark this compared to idiomatic channel implementation
		for writersRunning.Load() >= int32(ncpu) {
			time.Sleep(time.Millisecond)
		}

		writersRunning.Add(1)
		go func(idx int, off int64) {
			defer writersRunning.Add(-1)

			r, err := os.OpenFile(t.files[idx], os.O_RDONLY, 0644)
			if err != nil {
				log.Fatalf("error opening %s for reading: %s\n",
					t.files[idx],
					err,
				)
			}
			defer r.Close()

			w := io.NewOffsetWriter(fout, off)

			fmt.Fprintf(w, "||| %s\n", t.files[idx])
			_, err = io.Copy(w, r)
			if err != nil {
				log.Fatalf("error copying contents of %s into output: %s\n",
					t.files[idx],
					err,
				)
			}
			fmt.Fprint(w, "\n\n")

			// fmt.Printf("%10d %s\n", off, t.files[idx]) // TODO: add -v flag

		}(i, writeOffset)

		writeOffset += (t.sizes[i] + 7 + int64(len(t.files[i])))
	}

	for writersRunning.Load() > 0 {
		time.Sleep(time.Millisecond)
	}
}

// Shamelessly stolen from chi
type Middleware func(fs.WalkDirFunc) fs.WalkDirFunc

func (t *Turyn) Chain(middleware ...Middleware) fs.WalkDirFunc {
	wdf := middleware[len(middleware)-1](t.dummy)
	for i := len(middleware) - 2; i >= 0; i-- {
		wdf = middleware[i](wdf)
	}

	return wdf
}

func (t *Turyn) Default(next fs.WalkDirFunc) fs.WalkDirFunc {
	return fs.WalkDirFunc(func(path string, d fs.DirEntry, err error) error {
		if d == nil {
			panic("DirEntry is nil")
		}

		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return Err(path, err)
			}

			t.files = append(t.files, path)
			t.sizes = append(t.sizes, info.Size())

			//	fmt.Printf("%10d %s\n", info.Size(), path) // debug
		}

		return next(path, d, err)
	})
}

func (t *Turyn) dummy(path string, d fs.DirEntry, err error) error {
	return nil
}

func (t *Turyn) GetTotalSize() int64 {
	var totalSize int64
	for i := range t.sizes {
		// ||| + " " + filename + \n + filesize + \n\n (after contents)
		totalSize += (7 + int64(len(t.files[i])) + t.sizes[i])
	}
	return totalSize
}

// testing util
func (t *Turyn) Clear() {
	t.files = make([]string, 0)
	t.sizes = make([]int64, 0)
}
