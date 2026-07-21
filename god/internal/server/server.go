package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/racore/god/internal/authority"
	"github.com/racore/god/internal/ipfs"
	"github.com/racore/god/internal/kubo"
	"github.com/racore/god/internal/mesh"
	"github.com/racore/god/internal/providers"
	"github.com/racore/god/internal/vault"
	"github.com/racore/god/pkg/api"
)

type Server struct {
	cfg     api.Config
	vault   *vault.Vault
	gateway *providers.Gateway
	ipfs    *ipfs.Bridge
	kubo    *kubo.Manager
	author  *authority.Authority
	mesh    *mesh.MeshNode
	hub     *Hub
	version string
}

func New(cfg api.Config) *Server {
	v := vault.New(cfg.DataDir)
	if err := v.Load(); err != nil {
		log.Printf("vault load: %v", err)
	}

	gateway := providers.NewGateway(func(id string) (string, error) {
		key, ok := v.Get(id)
		if !ok {
			return "", fmt.Errorf("not connected")
		}
		return key, nil
	})

	ipfsBridge := ipfs.New(cfg.IPFSAPI, cfg.IPFSGateway)
	kuboMgr := kubo.New(cfg.IPFSAPI, cfg.IPFSGateway, cfg.DataDir)
	auth := authority.New(cfg.DataDir)
	if err := auth.Load(); err != nil {
		log.Printf("authority load: %v", err)
	}

	mn := mesh.NewMeshNode(cfg)

	s := &Server{
		cfg:     cfg,
		vault:   v,
		gateway: gateway,
		ipfs:    ipfsBridge,
		kubo:    kuboMgr,
		author:  auth,
		mesh:    mn,
		hub:     NewHub(),
		version: "0.2.0",
	}

	mn.SetEventSink(func(ev map[string]any) {
		if t, _ := ev["type"].(string); t == "mesh.message" {
			if mt, _ := ev["messageType"].(string); mt == "domain.claim" {
				if data, ok := ev["data"].(map[string]any); ok {
					domain, _ := data["domain"].(string)
					controller, _ := data["controller"].(string)
					nodeID, _ := ev["nodeId"].(string)
					auth.ObserveClaim(domain, controller, nodeID)
				}
			}
		}
		s.hub.Broadcast(ev)
	})

	return s
}

func (s *Server) Start(ctx context.Context) error {
	if _, err := s.kubo.Start(ctx); err != nil {
		log.Printf("kubo start: %v", err)
	}
	if s.cfg.MeshEnabled {
		if err := s.mesh.Start(ctx); err != nil {
			log.Printf("mesh start: %v", err)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/v1/providers", s.providersHandler)
	mux.HandleFunc("/v1/providers/", s.providerByIDHandler)
	mux.HandleFunc("/v1/chat", s.chatHandler)
	mux.HandleFunc("/v1/ipfs/status", s.ipfsStatusHandler)
	mux.HandleFunc("/v1/ipfs/add", s.ipfsAddHandler)
	mux.HandleFunc("/v1/ipfs/pin/", s.ipfsPinHandler)
	mux.HandleFunc("/v1/mesh/status", s.meshStatusHandler)
	mux.HandleFunc("/v1/mesh/peers", s.meshPeersHandler)
	mux.HandleFunc("/v1/mesh/broadcast", s.meshBroadcastHandler)
	mux.HandleFunc("/v1/authority/domains", s.domainsHandler)
	mux.HandleFunc("/v1/authority/domains/", s.domainByIDHandler)
	mux.HandleFunc("/v1/events", s.wsHandler)

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	httpSrv := &http.Server{
		Addr:         addr,
		Handler:     corsMiddleware(mux),
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout: 120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("racored listening on %s", addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-time.After(50 * time.Millisecond):
		return nil
	}
}

func (s *Server) Stop() {
	s.mesh.Stop()
	s.kubo.Stop()
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := map[string]bool{
			"http://localhost:3000":        true,
			"http://127.0.0.1:3000":        true,
			"http://[::1]:3000":            true,
			"http://127.0.0.1:47832":       true,
			"https://racore.xyz":           true,
		}
		if allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"detail": msg})
}
