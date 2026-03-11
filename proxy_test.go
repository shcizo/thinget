package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxyFetchVersionIndex(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3-flatcontainer/newtonsoft.json/index.json" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"versions":["13.0.1","13.0.3"]}`))
	}))
	defer upstream.Close()

	p := NewProxy(upstream.URL)

	data, statusCode, err := p.FetchVersionIndex("newtonsoft.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if statusCode != 200 {
		t.Errorf("expected 200, got %d", statusCode)
	}
	if string(data) != `{"versions":["13.0.1","13.0.3"]}` {
		t.Errorf("unexpected body: %s", string(data))
	}
}

func TestProxyFetchPackage(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := "/v3-flatcontainer/foo/1.0.0/foo.1.0.0.nupkg"
		if r.URL.Path != expected {
			t.Errorf("expected path %s, got %s", expected, r.URL.Path)
		}
		w.Write([]byte("fake-nupkg-bytes"))
	}))
	defer upstream.Close()

	p := NewProxy(upstream.URL)

	body, statusCode, err := p.FetchPackage("foo", "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer body.Close()

	if statusCode != 200 {
		t.Errorf("expected 200, got %d", statusCode)
	}

	data, _ := io.ReadAll(body)
	if string(data) != "fake-nupkg-bytes" {
		t.Errorf("unexpected body: %s", string(data))
	}
}

func TestProxyFetchPackage_NotFound(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer upstream.Close()

	p := NewProxy(upstream.URL)

	body, statusCode, err := p.FetchPackage("nonexistent", "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body != nil {
		body.Close()
	}
	if statusCode != 404 {
		t.Errorf("expected 404, got %d", statusCode)
	}
}
