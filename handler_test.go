package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
