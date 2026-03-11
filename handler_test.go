package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	handler := NewHandler(nil, nil, Config{})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestServiceIndexHandler(t *testing.T) {
	handler := NewHandler(nil, nil, Config{
		Upstream: "https://api.nuget.org",
	})

	req := httptest.NewRequest("GET", "/v3/index.json", nil)
	req.Host = "localhost:5555"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)

	if result["version"] != "3.0.0" {
		t.Errorf("expected version 3.0.0, got %v", result["version"])
	}

	resources, ok := result["resources"].([]interface{})
	if !ok || len(resources) == 0 {
		t.Fatal("expected resources array")
	}

	first := resources[0].(map[string]interface{})
	if first["@type"] != "PackageBaseAddress/3.0.0" {
		t.Errorf("expected PackageBaseAddress, got %v", first["@type"])
	}
	if first["@id"] != "http://localhost:5555/v3/flat/" {
		t.Errorf("expected self base URL, got %v", first["@id"])
	}
}

func TestPackageDownload_CacheHit(t *testing.T) {
	cacheDir := t.TempDir()
	cache := NewCache(cacheDir)

	// Pre-populate cache
	pkgDir := filepath.Join(cacheDir, "foo", "1.0.0")
	os.MkdirAll(pkgDir, 0755)
	os.WriteFile(filepath.Join(pkgDir, "foo.1.0.0.nupkg"), []byte("cached-pkg"), 0644)

	handler := NewHandler(cache, nil, Config{})

	req := httptest.NewRequest("GET", "/v3/flat/foo/1.0.0/foo.1.0.0.nupkg", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "cached-pkg" {
		t.Errorf("expected cached-pkg, got %s", w.Body.String())
	}
}

func TestPackageDownload_CacheMiss(t *testing.T) {
	cacheDir := t.TempDir()
	cache := NewCache(cacheDir)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("upstream-pkg"))
	}))
	defer upstream.Close()

	proxy := NewProxy(upstream.URL)
	handler := NewHandler(cache, proxy, Config{})

	req := httptest.NewRequest("GET", "/v3/flat/bar/2.0.0/bar.2.0.0.nupkg", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "upstream-pkg" {
		t.Errorf("expected upstream-pkg, got %s", w.Body.String())
	}

	// Verify it was cached
	if !cache.HasPackage("bar", "2.0.0") {
		t.Error("expected package to be cached after fetch")
	}
}

func TestVersionIndex_Proxy(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"versions":["1.0.0","2.0.0"]}`))
	}))
	defer upstream.Close()

	proxy := NewProxy(upstream.URL)
	handler := NewHandler(nil, proxy, Config{})

	req := httptest.NewRequest("GET", "/v3/flat/foo/index.json", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != `{"versions":["1.0.0","2.0.0"]}` {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPackageDownload_CacheHit_UpstreamDown(t *testing.T) {
	cacheDir := t.TempDir()
	cache := NewCache(cacheDir)

	pkgDir := filepath.Join(cacheDir, "foo", "1.0.0")
	os.MkdirAll(pkgDir, 0755)
	os.WriteFile(filepath.Join(pkgDir, "foo.1.0.0.nupkg"), []byte("cached-pkg"), 0644)

	proxy := NewProxy("http://localhost:1")
	handler := NewHandler(cache, proxy, Config{})

	req := httptest.NewRequest("GET", "/v3/flat/foo/1.0.0/foo.1.0.0.nupkg", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 from cache even with upstream down, got %d", w.Code)
	}
	if w.Body.String() != "cached-pkg" {
		t.Errorf("expected cached-pkg, got %s", w.Body.String())
	}
}

func TestPackageDownload_UpstreamNotFound(t *testing.T) {
	cacheDir := t.TempDir()
	cache := NewCache(cacheDir)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer upstream.Close()

	proxy := NewProxy(upstream.URL)
	handler := NewHandler(cache, proxy, Config{})

	req := httptest.NewRequest("GET", "/v3/flat/nonexistent/1.0.0/nonexistent.1.0.0.nupkg", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 passthrough from upstream, got %d", w.Code)
	}
}
