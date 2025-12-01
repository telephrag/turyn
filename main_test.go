package main

import (
	"bytes"
	"io"
	"os"
	"runtime"
	"syscall"
	"testing"

	"github.com/telephrag/turyn/turyn"
)

func BenchmarkGatherFS(b *testing.B) {
	t := turyn.New()

	for b.Loop() {
		err := t.GatherFS("testdata/input",
			t.Chain([]turyn.Middleware{t.CollectPathSize}...))

		b.StopTimer()
		if err != nil {
			b.FailNow()
		}
		t.Clear()
		b.StartTimer()
	}
}

func BenchmarkGatherFilepath(b *testing.B) {
	t := turyn.New()

	for b.Loop() {
		err := t.Gather("testdata/input",
			t.Chain([]turyn.Middleware{t.CollectPathSize}...))

		b.StopTimer()
		if err != nil {
			b.FailNow()
		}
		t.Clear()
		b.StartTimer()
	}
}

// Why you don't use atomics as synchronisation primitives.
// See turyn.ProcessAtomicWait() for details.
func BenchmarkFullAtomic(b *testing.B) {
	t := turyn.New()

	for b.Loop() {
		err := t.Gather("testdata/input",
			t.Chain([]turyn.Middleware{t.CheckIfDir, t.CollectPathSize}...))

		b.StopTimer()
		if err != nil {
			b.FailNow()
		}
		b.StartTimer()

		fout, err := os.OpenFile("testdata/test_atomic.tur", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
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

// Actual benchmark of full program execution
func BenchmarkFullChan(b *testing.B) {
	t := turyn.New()

	for b.Loop() {
		err := t.Gather("testdata/input",
			t.Chain([]turyn.Middleware{t.CheckIfDir, t.CollectPathSize}...))

		b.StopTimer()
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()

		fout, err := os.OpenFile("testdata/test_chan.tur", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
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

	err := tur.Gather("testdata/input/", tur.Chain([]turyn.Middleware{
		tur.CheckIfDir, tur.CollectPathSize}...))
	if err != nil {
		t.Fatalf("%v: did you generate testdata via './do.sh gentest?'", err)
	}

	fgot, _ := os.OpenFile("testdata/got.tur",
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)

	syscall.Fallocate(int(fgot.Fd()), 0, 0, tur.GetTotalSize())

	tur.Process(fgot, 0, runtime.NumCPU())

	fgot.Close()

	fgot, _ = os.OpenFile("testdata/got.tur", os.O_RDONLY, 0644)
	fexp, _ := os.OpenFile("testdata/expected.tur", os.O_RDONLY, 0644)

	fgotContent, _ := io.ReadAll(fgot)
	fexpContent, _ := io.ReadAll(fexp)

	if !bytes.Equal(fgotContent, fexpContent) {
		t.Log(`expected.tur and got.tur contents do not match\n`)
		t.Fail()
	}
}
