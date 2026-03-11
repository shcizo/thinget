package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestCacheHasPackage_Miss(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(dir)

	if c.HasPackage("newtonsoft.json", "13.0.1") {
		t.Error("expected cache miss for nonexistent package")
	}
}

func TestCacheGetPackage(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(dir)

	// Manually place a package file
	pkgDir := filepath.Join(dir, "foo", "1.0.0")
	os.MkdirAll(pkgDir, 0755)
	os.WriteFile(filepath.Join(pkgDir, "foo.1.0.0.nupkg"), []byte("fake-nupkg"), 0644)

	reader, err := c.GetPackage("foo", "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer reader.Close()

	data, _ := io.ReadAll(reader)
	if string(data) != "fake-nupkg" {
		t.Errorf("expected fake-nupkg, got %s", string(data))
	}
}

func TestCachePutPackage(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(dir)

	data := []byte("nupkg-content")
	err := c.PutPackage("bar", "2.0.0", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !c.HasPackage("bar", "2.0.0") {
		t.Error("expected package to exist after put")
	}

	reader, _ := c.GetPackage("bar", "2.0.0")
	defer reader.Close()
	got, _ := io.ReadAll(reader)
	if string(got) != "nupkg-content" {
		t.Errorf("expected nupkg-content, got %s", string(got))
	}
}
