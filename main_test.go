package main

import (
	"path/filepath"
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
		err := t.Gather("testdata/bscrap_srv/", t.Chain([]turyn.Middleware{t.Default}...))

		b.StopTimer()
		if err != nil {
			b.FailNow()
		}
		t.Clear()
		b.StartTimer()
	}

}
