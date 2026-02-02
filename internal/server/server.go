package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"go-proxy-manager/internal/store"
)

type Server struct {
	port  int
	store *store.Store
}

func NewServer(port int, s *store.Store) *Server {
	return &Server{
		port:  port,
		store: s,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/proxies", s.handleGetProxies)
	mux.HandleFunc("GET /health", s.handleHealth)

	addr := fmt.Sprintf("0.0.0.0:%d", s.port)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) handleGetProxies(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	format := r.URL.Query().Get("format")

	limit := 0
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	proxies := s.store.GetSorted(limit)

	if format == "text" {
		w.Header().Set("Content-Type", "text/plain")
		for _, p := range proxies {
			// Return as protocol://ip:port
			fmt.Fprintf(w, "%s://%s:%d\n", p.Protocol, p.IP, p.Port)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(proxies)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
