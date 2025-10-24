package turyn

import (
	"fmt"
	"io/fs"
	"os"
)

type Turyn struct {
	workDir string
	files   []string // absolute paths to files
	sizes   []int64  // size of files for future preallocation
}

func New(workDir string) *Turyn {
	return &Turyn{
		workDir: workDir,
		files:   []string{},
		sizes:   []int64{},
	}
}

func (t *Turyn) Walk(workDir string, wdf fs.WalkDirFunc) error {
	return fs.WalkDir(os.DirFS(workDir), ".", wdf)
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

			fmt.Printf("%10d %s\n", info.Size(), path) // debug
		}

		return next(path, d, err)
	})
}

func (t *Turyn) dummy(path string, d fs.DirEntry, err error) error {
	return nil
}

func (t *Turyn) GetFileCount() int {
	return len(t.files)
}

func (t *Turyn) Prepare() ([]string, []int64) {
	return t.files, t.sizes
}
