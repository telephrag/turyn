package main

import (
	"log"
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
		err := t.GatherFS(wd, t.Chain([]turyn.Middleware{t.Default}...))

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
		err := t.Gather("testdata/", t.Chain([]turyn.Middleware{t.Default}...))

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
		err := t.Gather("testdata/", t.Chain([]turyn.Middleware{t.Default}...))

		b.StopTimer()
		if err != nil {
			b.FailNow()
		}
		b.StartTimer()

		fout, err := os.OpenFile("test_atomic.tur", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
		if err != nil {
			log.Fatalln(err)
		}

		err = syscall.Fallocate(int(fout.Fd()), 0, 0, t.GetTotalSize())
		if err != nil {
			log.Fatalf("error allocating filespace for output: %s\n", err)
		}

		t.ProcessAtomicWait(fout, 0, runtime.NumCPU())

		b.StopTimer()
		t.Clear()
		b.StartTimer()
	}
}

func BenchmarkFullChan(b *testing.B) {
	t := turyn.New()

	for b.Loop() {
		err := t.Gather("testdata/", t.Chain([]turyn.Middleware{t.Default}...))

		b.StopTimer()
		if err != nil {
			b.FailNow()
		}
		b.StartTimer()

		fout, err := os.OpenFile("test_chan.tur", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
		if err != nil {
			log.Fatalln(err)
		}

		err = syscall.Fallocate(int(fout.Fd()), 0, 0, t.GetTotalSize())
		if err != nil {
			log.Fatalf("error allocating filespace for output: %s\n", err)
		}

		t.Process(fout, 0, runtime.NumCPU())

		b.StopTimer()
		t.Clear()
		b.StartTimer()
	}
}
