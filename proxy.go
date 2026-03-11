package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Proxy struct {
	upstream string
	client   *http.Client
}

func NewProxy(upstream string) *Proxy {
	return &Proxy{
		upstream: strings.TrimRight(upstream, "/"),
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *Proxy) FetchVersionIndex(id string) ([]byte, int, error) {
	url := fmt.Sprintf("%s/v3-flatcontainer/%s/index.json", p.upstream, strings.ToLower(id))

	resp, err := p.client.Get(url)
	if err != nil {
		return nil, 0, fmt.Errorf("fetch version index: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}

	return body, resp.StatusCode, nil
}

func (p *Proxy) FetchPackage(id, version string) (io.ReadCloser, int, error) {
	lowerID := strings.ToLower(id)
	lowerVer := strings.ToLower(version)
	url := fmt.Sprintf("%s/v3-flatcontainer/%s/%s/%s.%s.nupkg", p.upstream, lowerID, lowerVer, lowerID, lowerVer)

	resp, err := p.client.Get(url)
	if err != nil {
		return nil, 0, fmt.Errorf("fetch package: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, resp.StatusCode, nil
	}

	return resp.Body, resp.StatusCode, nil
}
