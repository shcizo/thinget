package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Cache struct {
	dir string
}

func NewCache(dir string) *Cache {
	return &Cache{dir: dir}
}

func (c *Cache) packagePath(id, version string) string {
	lowerID := strings.ToLower(id)
	lowerVer := strings.ToLower(version)
	return filepath.Join(c.dir, lowerID, lowerVer, fmt.Sprintf("%s.%s.nupkg", lowerID, lowerVer))
}

func (c *Cache) HasPackage(id, version string) bool {
	_, err := os.Stat(c.packagePath(id, version))
	return err == nil
}

func (c *Cache) GetPackage(id, version string) (io.ReadCloser, error) {
	return os.Open(c.packagePath(id, version))
}

func (c *Cache) PutPackage(id, version string, r io.Reader) error {
	finalPath := c.packagePath(id, version)
	dir := filepath.Dir(finalPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, "*.nupkg.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := io.Copy(tmp, r); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename to final path: %w", err)
	}

	return nil
}
