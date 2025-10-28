package turyn

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
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

type Middleware func(fs.WalkDirFunc) fs.WalkDirFunc

func (t *Turyn) Chain(middleware ...Middleware) fs.WalkDirFunc {
	wdf := middleware[len(middleware)-1](t.dummy)
	for i := len(middleware) - 2; i >= 0; i-- {
		wdf = middleware[i](wdf)
	}
	return wdf
}

func (t *Turyn) Gather(workDir string, wdf fs.WalkDirFunc) error {
	return filepath.WalkDir(workDir, wdf)
}

// Shamelessly stolen from chi
func (t *Turyn) CheckIfDir(next fs.WalkDirFunc) fs.WalkDirFunc {
	return fs.WalkDirFunc(func(path string, d fs.DirEntry, err error) error {
		if d == nil {
			panic("DirEntry is nil")
		}

		if !d.IsDir() {
			// fmt.Println(path) // debug
			return next(path, d, err)
		}

		return nil
	})
}

func (t *Turyn) CollectPathSize(next fs.WalkDirFunc) fs.WalkDirFunc {
	return fs.WalkDirFunc(func(path string, d fs.DirEntry, err error) error {
		info, err := d.Info()
		if err != nil {
			return Err(path, err)
		}

		t.files = append(t.files, path)
		t.sizes = append(t.sizes, info.Size())

		//	fmt.Printf("%10d %s\n", info.Size(), path) // debug

		return next(path, d, err)
	})
}

func (t *Turyn) dummy(path string, d fs.DirEntry, err error) error {
	return nil
}

func (t *Turyn) Process(fout *os.File, baseOffset int64, ncpu int) {
	var writeOffset int64
	var wg sync.WaitGroup
	writers := make(chan struct{}, ncpu)
	for i := range t.files {
		writers <- struct{}{}
		wg.Add(1)
		go func(idx int, off int64) {
			t.WriteFile(idx, off, fout)
			<-writers
			wg.Done()
		}(i, writeOffset)

		writeOffset += (t.sizes[i] + 7 + int64(len(t.files[i])))
	}
	wg.Wait()
}

func (t *Turyn) WriteFile(i int, off int64, fout *os.File) error {
	r, err := os.OpenFile(t.files[i], os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening %s for reading: %w\n",
			t.files[i],
			err,
		)
	}
	defer r.Close()

	// TODO: is creating these thread safe?
	w := io.NewOffsetWriter(fout, off)

	fmt.Fprint(w, "\n")

	fmt.Fprintf(w, "||| %s\n", t.files[i])
	_, err = io.Copy(w, r)
	if err != nil {
		return fmt.Errorf("error copying contents of %s into output: %w\n",
			t.files[i],
			err,
		)
	}
	fmt.Fprint(w, "\n")

	// fmt.Printf("%10d %s\n", off, t.files[idx]) // TODO: add -v flag

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
