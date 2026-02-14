package engine

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"

	"frontman/internal/config"
)

type upstreamState struct {
	servers []string
	mu      sync.Mutex
	next    int
}

type Engine struct {
	cfg       *config.Config
	upstreams map[string]*upstreamState
	routes    []config.Route
}

func NewEngine(cfg *config.Config) *Engine {
	u := make(map[string]*upstreamState)
	for _, up := range cfg.Upstreams {
		// copy servers slice
		servers := make([]string, len(up.Servers))
		copy(servers, up.Servers)
		u[up.Name] = &upstreamState{servers: servers}
	}
	return &Engine{cfg: cfg, upstreams: u, routes: cfg.Routes}
}

// HandleRequest selects a backend for the provided path using round-robin
// and forwards the provided request to that backend, returning the backend
// response. The request's URL path and query are preserved.
func (e *Engine) HandleRequest(r *http.Request) (*http.Response, error) {
	reqPath := r.URL.Path
	route := e.matchRoute(reqPath)
	if route == nil {
		log.Printf("no route matched for path %s", reqPath)
		return nil, errors.New("no route matched")
	}

	ups, ok := e.upstreams[route.Upstream]
	if !ok || len(ups.servers) == 0 {
		log.Printf("upstream %s has no servers", route.Upstream)
		return nil, fmt.Errorf("upstream %s has no servers", route.Upstream)
	}

	// pick server with round-robin
	ups.mu.Lock()
	idx := ups.next
	ups.next = (ups.next + 1) % len(ups.servers)
	ups.mu.Unlock()

	backend := ups.servers[idx]
	log.Printf("selected backend %s for route %s (path=%s)", backend, route.Path, reqPath)

	// Build target URL by joining backend and the request path + query
	targetURL, err := url.Parse(backend)
	if err != nil {
		log.Printf("invalid backend URL %s: %v", backend, err)
		return nil, fmt.Errorf("invalid backend URL %s: %w", backend, err)
	}

	// Preserve the original path suffix after the route prefix.
	// If the backend URL has a path, join it with the requested path
	// e.g., backend http://localhost:3000/api and request /v1/x -> /api/v1/x
	targetPath := path.Join(targetURL.Path, reqPath)
	targetURL.Path = targetPath
	targetURL.RawQuery = r.URL.RawQuery

	// Create a new request to the backend
	outReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL.String(), r.Body)
	if err != nil {
		log.Printf("request to backend %s failed: %v", targetURL.String(), err)
		return nil, err
	}

	// Copy headers
	outReq.Header = make(http.Header)
	for k, vv := range r.Header {
		for _, v := range vv {
			outReq.Header.Add(k, v)
		}
	}

	// Forward
	resp, err := http.DefaultClient.Do(outReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// matchRoute finds the best (longest prefix) route matching the request path.
func (e *Engine) matchRoute(reqPath string) *config.Route {
	var best *config.Route
	bestLen := -1
	for i := range e.routes {
		r := &e.routes[i]
		if strings.HasPrefix(reqPath, r.Path) {
			l := len(r.Path)
			if l > bestLen {
				best = r
				bestLen = l
			}
		}
	}
	return best
}
