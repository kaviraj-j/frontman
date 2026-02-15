package stats

import (
	"frontman/internal/config"
	"sync"
)

type ApiStats struct {
	mu     sync.RWMutex
	Routes map[string]*UpstreamStats `json:"routes"`
}

type UpstreamStats struct {
	UpstreamName string                  `json:"upstream_name"`
	Servers      map[string]*ServerStats `json:"servers"`
}

type ServerStats struct {
	ServerURL           string      `json:"server_url"`
	TotalHits           int         `json:"total_hits"`
	ResponseStatusCount map[int]int `json:"response_status_count"`
}

func NewApiStats(cfg *config.Config) *ApiStats {
	stats := &ApiStats{
		Routes: make(map[string]*UpstreamStats),
	}

	for _, route := range cfg.Routes {
		upstream := findUpstream(cfg, route.Upstream)
		if upstream == nil {
			continue
		}

		upStats := &UpstreamStats{
			UpstreamName: upstream.Name,
			Servers:      make(map[string]*ServerStats),
		}

		for _, srv := range upstream.Servers {
			upStats.Servers[srv] = &ServerStats{
				ServerURL:           srv,
				ResponseStatusCount: make(map[int]int),
			}
		}

		stats.Routes[route.Path] = upStats
	}

	return stats
}

func (a *ApiStats) Update(routePath, serverURL string, statusCode int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	upstream, ok := a.Routes[routePath]
	if !ok {
		return
	}

	server, ok := upstream.Servers[serverURL]
	if !ok {
		return
	}

	server.TotalHits++
	server.ResponseStatusCount[statusCode]++
}

func (a *ApiStats) Snapshot() map[string]*UpstreamStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	copyRoutes := make(map[string]*UpstreamStats, len(a.Routes))

	for routePath, upstream := range a.Routes {
		upCopy := &UpstreamStats{
			UpstreamName: upstream.UpstreamName,
			Servers:      make(map[string]*ServerStats),
		}

		for serverURL, server := range upstream.Servers {
			statusCopy := make(map[int]int, len(server.ResponseStatusCount))
			for code, count := range server.ResponseStatusCount {
				statusCopy[code] = count
			}

			upCopy.Servers[serverURL] = &ServerStats{
				ServerURL:           server.ServerURL,
				TotalHits:           server.TotalHits,
				ResponseStatusCount: statusCopy,
			}
		}

		copyRoutes[routePath] = upCopy
	}

	return copyRoutes
}

func findUpstream(cfg *config.Config, name string) *config.Upstream {
	for i := range cfg.Upstreams {
		if cfg.Upstreams[i].Name == name {
			return &cfg.Upstreams[i]
		}
	}
	return nil
}
