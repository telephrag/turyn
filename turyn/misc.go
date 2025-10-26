package turyn

import (
	"io/fs"
	"log"
	"os"
	"sync/atomic"
	"time"
)

// testing util
func (t *Turyn) Clear() {
	t.files = make([]string, 0)
	t.sizes = make([]int64, 0)
}

// inferior to chan+waitgroup, left to be for now
func (t *Turyn) ProcessAtomicWait(fout *os.File, baseOffset int64, ncpu int) {
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
			if err := t.WriteFile(idx, off, fout); err != nil {
				log.Fatal(err)
			}
		}(i, writeOffset)

		writeOffset += (t.sizes[i] + 7 + int64(len(t.files[i])))
	}

	for writersRunning.Load() > 0 {
		time.Sleep(time.Millisecond)
	}
}

// left to be for benchmarking later
func (t *Turyn) GatherFS(workDir string, wdf fs.WalkDirFunc) error {
	return fs.WalkDir(os.DirFS(workDir), ".", wdf)
}
