package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type Handler struct {
	cache  *Cache
	proxy  *Proxy
	config Config
	mux    *http.ServeMux
}

func NewHandler(cache *Cache, proxy *Proxy, cfg Config) *Handler {
	h := &Handler{
		cache:  cache,
		proxy:  proxy,
		config: cfg,
		mux:    http.NewServeMux(),
	}
	h.mux.HandleFunc("/health", h.handleHealth)
	h.mux.HandleFunc("/v3/index.json", h.handleServiceIndex)
	h.mux.HandleFunc("/v3/flat/", h.handleFlat)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) handleServiceIndex(w http.ResponseWriter, r *http.Request) {
	selfBase := fmt.Sprintf("http://%s/v3/flat/", r.Host)
	upstream := strings.TrimRight(h.config.Upstream, "/")

	index := map[string]interface{}{
		"version": "3.0.0",
		"resources": []map[string]string{
			{
				"@id":     selfBase,
				"@type":   "PackageBaseAddress/3.0.0",
				"comment": "Package content (cached by ThinGet)",
			},
			{
				"@id":     upstream + "/v3/registration5-semver1/",
				"@type":   "RegistrationsBaseUrl",
				"comment": "Package metadata (upstream passthrough)",
			},
			{
				"@id":     upstream + "/v3/registration5-semver1/",
				"@type":   "RegistrationsBaseUrl/3.0.0-rc",
				"comment": "Package metadata (upstream passthrough)",
			},
			{
				"@id":     upstream + "/v3/registration5-gz-semver2/",
				"@type":   "RegistrationsBaseUrl/3.6.0",
				"comment": "Package metadata (upstream passthrough)",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(index)
}

func (h *Handler) handleFlat(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v3/flat/")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) == 2 && parts[1] == "index.json" {
		h.handleVersionIndex(w, r, parts[0])
		return
	}

	if len(parts) == 3 && strings.HasSuffix(parts[2], ".nupkg") {
		h.handlePackageDownload(w, r, parts[0], parts[1])
		return
	}

	http.NotFound(w, r)
}

func (h *Handler) handleVersionIndex(w http.ResponseWriter, r *http.Request, id string) {
	data, statusCode, err := h.proxy.FetchVersionIndex(id)
	if err != nil {
		http.Error(w, "upstream error", http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(data)
}

func (h *Handler) handlePackageDownload(w http.ResponseWriter, r *http.Request, id, version string) {
	reader, err := h.cache.GetPackage(id, version)
	if err == nil {
		defer reader.Close()
		w.Header().Set("Content-Type", "application/octet-stream")
		io.Copy(w, reader)
		return
	}
	if !errors.Is(err, os.ErrNotExist) {
		http.Error(w, "cache read error", http.StatusInternalServerError)
		return
	}

	body, statusCode, err := h.proxy.FetchPackage(id, version)
	if err != nil {
		http.Error(w, "upstream error", http.StatusBadGateway)
		return
	}
	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
		return
	}
	defer body.Close()

	if err := h.cache.PutPackage(id, version, body); err != nil {
		log.Printf("cache write failed: %v", err)
		http.Error(w, "cache write error", http.StatusInternalServerError)
		return
	}

	reader, err = h.cache.GetPackage(id, version)
	if err != nil {
		http.Error(w, "cache read error", http.StatusInternalServerError)
		return
	}
	defer reader.Close()
	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(w, reader)
}
