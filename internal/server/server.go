package server

import (
	"frontman/internal/config"
	"frontman/internal/engine"
	"frontman/internal/stats"
	"io"
	"log"
	"net/http"
	"strconv"
)

type Server struct {
	proxyEngine *engine.Engine
	cfg         *config.Config
	apiStats    *stats.ApiStats
}

func NewServer(eng *engine.Engine, cfg *config.Config, apiStats *stats.ApiStats) *Server {
	return &Server{proxyEngine: eng, cfg: cfg, apiStats: apiStats}
}

func (s *Server) Run() {
	// Main proxy server
	proxy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Proxy received request method=%s path=%s", r.Method, r.URL.Path)
		resp, err := s.proxyEngine.HandleRequest(r)
		if err != nil {
			log.Printf("engine error for path %s: %v", r.URL.Path, err)
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	// Dashboard server (port 8081)
	go func() {
		dashboardAddr := ":8081"
		log.Printf("Dashboard running on http://localhost%s", dashboardAddr)
		webDir := "web"
		handler := DashboardHandler(s.apiStats, webDir)
		if err := http.ListenAndServe(dashboardAddr, handler); err != nil {
			log.Printf("dashboard server error: %v", err)
		}
	}()

	log.Printf("listening on port %d", s.cfg.Server.Listen)
	addr := ":" + strconv.Itoa(s.cfg.Server.Listen)
	if err := http.ListenAndServe(addr, proxy); err != nil {
		log.Printf("server error: %v", err)
	}
}
