package server

import (
	"encoding/json"
	"frontman/internal/stats"
	"net/http"
	"os"
	"path/filepath"
)

// DashboardHandler serves the dashboard HTML and stats JSON
func DashboardHandler(apiStats *stats.ApiStats, webDir string) http.Handler {
	mux := http.NewServeMux()

	// Serve the dashboard HTML
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Only serve index.html for root
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		f, err := os.Open(filepath.Join(webDir, "index.html"))
		if err != nil {
			http.Error(w, "dashboard not found", http.StatusNotFound)
			return
		}
		defer f.Close()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fi := fStat(f)
		http.ServeContent(w, r, "index.html", fi.ModTime(), f)
	})

	// Serve the stats as JSON
	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		snapshot := apiStats.Snapshot()
		json.NewEncoder(w).Encode(snapshot)
	})

	return mux
}

// Helper to get file stat
func fStat(f *os.File) (fi os.FileInfo) {
	fi, _ = f.Stat()
	return
}
