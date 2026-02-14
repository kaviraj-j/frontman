package server

import (
	"frontman/internal/config"
	"frontman/internal/engine"
	"io"
	"log"
	"net/http"
	"strconv"
)

type Server struct {
	proxyEngine *engine.Engine
	cfg         *config.Config
}

func NewServer(eng *engine.Engine, cfg *config.Config) *Server {
	return &Server{proxyEngine: eng, cfg: cfg}
}

func (s *Server) Run() {
	proxy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Proxy received request method=%s path=%s", r.Method, r.URL.Path)

		resp, err := s.proxyEngine.HandleRequest(r)
		if err != nil {
			log.Printf("engine error for path %s: %v", r.URL.Path, err)
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// copy response headers
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}

		// write status code then body
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	log.Printf("listening on port %d", s.cfg.Server.Listen)
	addr := ":" + strconv.Itoa(s.cfg.Server.Listen)
	err := http.ListenAndServe(addr, proxy)
	if err != nil {
		log.Printf("server error: %v", err)
	}
}
