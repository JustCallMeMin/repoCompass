package api

import (
	"net/http"
	"runtime"
	"time"
)

var serverStartTime = time.Now()

// handleMetrics exposes lightweight operational metrics at GET /api/v1/metrics.
// This is intentionally minimal and not Prometheus-compatible for the M6 MVP.
// A full observability stack (Prometheus scrape, OpenTelemetry) can replace this in M8+.
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	payload := map[string]any{
		"uptime_seconds": int64(time.Since(serverStartTime).Seconds()),
		"goroutines":     runtime.NumGoroutine(),
		"memory": map[string]any{
			"alloc_bytes":       mem.Alloc,
			"total_alloc_bytes": mem.TotalAlloc,
			"sys_bytes":         mem.Sys,
			"gc_cycles":         mem.NumGC,
		},
		"go_version": runtime.Version(),
	}

	writeData(w, r, http.StatusOK, payload)
}
