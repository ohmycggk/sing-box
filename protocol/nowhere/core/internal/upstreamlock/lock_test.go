package upstreamlock

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestCanonicalLockMatchesGeneratedConstantsAndVectors(t *testing.T) {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(source), "..", ".."))
	lock, err := Read(filepath.Join(root, "UPSTREAM-nw2.lock"))
	if err != nil {
		t.Fatal(err)
	}
	if !lock.MatchesGenerated() {
		t.Fatal("UPSTREAM-nw2.lock differs from generated constants")
	}
	got, err := TreeHash(filepath.Join(root, "testdata", "vectors"))
	if err != nil {
		t.Fatal(err)
	}
	if got != lock.VectorTreeSHA256 {
		t.Fatalf("vector tree hash = %s, want %s", got, lock.VectorTreeSHA256)
	}
}

func TestDefaultLockIsNowhere3(t *testing.T) {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(source), "..", ".."))
	lock, err := Read(filepath.Join(root, "UPSTREAM.lock"))
	if err != nil {
		t.Fatal(err)
	}
	if lock.Version != "3.0.0" || lock.Commit != "501138f43a99cb5bbf40f45ee2a55cda5faf1c51" {
		t.Fatalf("UPSTREAM.lock = %s %s", lock.Version, lock.Commit)
	}
}
