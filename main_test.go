package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"turyn/turyn"
)

func BenchmarkGather(b *testing.B) {
	wd, _ := filepath.Abs("testdata/")
	t := turyn.New()

	for b.Loop() {
		err := t.GatherFS(wd, t.Chain([]turyn.Middleware{t.CollectPathSize}...))

		b.StopTimer()
		if err != nil {
			b.FailNow()
		}
		t.Clear()
		b.StartTimer()
	}
}

func BenchmarkGatherFilepath(b *testing.B) {
	//	wd, _ := filepath.Abs("testdata/")
	t := turyn.New()

	for b.Loop() {
		err := t.Gather("testdata/", t.Chain([]turyn.Middleware{t.CollectPathSize}...))

		b.StopTimer()
		if err != nil {
			b.FailNow()
		}
		t.Clear()
		b.StartTimer()
	}
}

func BenchmarkFullAtomic(b *testing.B) {
	t := turyn.New()

	for b.Loop() {
		err := t.Gather("testdata/",
			t.Chain([]turyn.Middleware{t.CheckIfDir, t.CollectPathSize}...))

		b.StopTimer()
		if err != nil {
			b.FailNow()
		}
		b.StartTimer()

		fout, err := os.OpenFile("test_atomic.tur", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
		if err != nil {
			b.Fatal(err)
		}

		err = syscall.Fallocate(int(fout.Fd()), 0, 0, t.GetTotalSize())
		if err != nil {
			b.Fatalf("error allocating filespace for output: %s\n", err)
		}

		t.ProcessAtomicWait(fout, 0, runtime.NumCPU())

		b.StopTimer()
		t.Clear()
		fout.Close()
		b.StartTimer()
	}
}

func BenchmarkFullChan(b *testing.B) {
	t := turyn.New()

	for b.Loop() {
		err := t.Gather("testdata/",
			t.Chain([]turyn.Middleware{t.CheckIfDir, t.CollectPathSize}...))

		b.StopTimer()
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()

		fout, err := os.OpenFile("test_chan.tur", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
		if err != nil {
			b.Fatal(err)
		}

		err = syscall.Fallocate(int(fout.Fd()), 0, 0, t.GetTotalSize())
		if err != nil {
			b.Fatalf("error allocating filespace for output: %s\n", err)
		}

		t.Process(fout, 0, runtime.NumCPU())

		b.StopTimer()
		t.Clear()
		fout.Close()
		b.StartTimer()
	}
}

func TestFull(t *testing.T) {
	tur := turyn.New()

	tur.Gather("testdata/input/", tur.Chain([]turyn.Middleware{
		tur.CheckIfDir, tur.CollectPathSize}...))

	fgot, _ := os.OpenFile("got.tur",
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)

	syscall.Fallocate(int(fgot.Fd()), 0, 0, tur.GetTotalSize())

	tur.Process(fgot, 0, runtime.NumCPU())

	fgot.Close()

	fgot, _ = os.OpenFile("got.tur", os.O_RDONLY, 0644)
	fexp, _ := os.OpenFile("testdata/expected.tur", os.O_RDONLY, 0644)

	fgotContent, _ := io.ReadAll(fgot)
	fexpContent, _ := io.ReadAll(fexp)

	fmt.Println(bytes.Compare(fgotContent, fexpContent))
}
