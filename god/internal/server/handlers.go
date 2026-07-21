package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/racore/god/pkg/api"
)

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	ipfsStatus, _ := s.ipfs.Status(r.Context())
	writeJSON(w, 200, api.HealthResponse{
		OK:        true,
		Version:   s.version,
		Mesh:      s.mesh.Status(),
		IPFS:      ipfsStatus,
		Providers: len(s.gateway.Catalog()),
	})
}

func (s *Server) providersHandler(w http.ResponseWriter, r *http.Request) {
	catalog := s.gateway.Catalog()
	writeJSON(w, 200, catalog)
}

func (s *Server) providerByIDHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/providers/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, 404, "provider not specified")
		return
	}
	providerID := parts[0]

	switch r.Method {
	case "PUT":
		if len(parts) > 1 && parts[1] == "connect" {
			var body struct {
				APIKey string `json:"api_key"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeError(w, 400, "invalid body")
				return
			}
			if len(body.APIKey) < 3 || len(body.APIKey) > 4096 {
				writeError(w, 400, "invalid key length")
				return
			}
			s.vault.Set(providerID, body.APIKey)
			health, _ := s.gateway.Health(r.Context(), providerID)
			s.hub.Broadcast(map[string]any{"type": "provider.connected", "provider": providerID})
			writeJSON(w, 200, map[string]any{
				"connected": true, "provider": providerID,
				"health": health, "maskedKey": s.vault.Masked(providerID),
			})
			return
		}
	case "DELETE":
		s.vault.Delete(providerID)
		writeJSON(w, 200, map[string]any{"connected": false, "provider": providerID})
		return
	case "GET":
		if len(parts) > 1 && parts[1] == "health" {
			health, err := s.gateway.Health(r.Context(), providerID)
			if err != nil {
				writeError(w, 404, err.Error())
				return
			}
			writeJSON(w, 200, health)
			return
		}
	}

	writeError(w, 405, "method not allowed")
}

func (s *Server) chatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}

	var req api.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid body")
		return
	}

	s.hub.Broadcast(map[string]any{"type": "agent.started", "provider": req.Provider, "model": req.Model})

	resp, err := s.gateway.Chat(r.Context(), req.Provider, req.Model, req.Messages, req.System)
	if err != nil {
		writeError(w, 502, err.Error())
		return
	}

	s.hub.Broadcast(map[string]any{
		"type": "agent.completed", "provider": req.Provider,
		"model": resp.Model, "latencyMs": resp.LatencyMs,
	})

	writeJSON(w, 200, resp)
}

func (s *Server) ipfsStatusHandler(w http.ResponseWriter, r *http.Request) {
	status, err := s.ipfs.Status(r.Context())
	if err != nil {
		writeJSON(w, 200, map[string]any{"online": false})
		return
	}
	writeJSON(w, 200, status)
}

func (s *Server) ipfsAddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}

	if err := r.ParseMultipartForm(100 << 20); err != nil {
		writeError(w, 400, "multipart parse error")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, 400, "file required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, 500, "read error")
		return
	}

	name := header.Filename
	if name == "" {
		name = "upload.bin"
	}

	result, err := s.ipfs.AddBytes(r.Context(), name, data)
	if err != nil {
		writeError(w, 503, "kubo unavailable: "+err.Error())
		return
	}

	s.mesh.Broadcast("release.available", map[string]any{
		"cid": result["cid"], "name": result["name"],
	})

	writeJSON(w, 200, result)
}

func (s *Server) ipfsPinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}
	cid := strings.TrimPrefix(r.URL.Path, "/v1/ipfs/pin/")
	if cid == "" {
		writeError(w, 400, "cid required")
		return
	}
	result, err := s.ipfs.Pin(r.Context(), cid)
	if err != nil {
		writeError(w, 503, err.Error())
		return
	}
	writeJSON(w, 200, result)
}

func (s *Server) meshStatusHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, s.mesh.Status())
}

func (s *Server) meshPeersHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, s.mesh.Peers())
}

func (s *Server) meshBroadcastHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}

	var body struct {
		Type string         `json:"type"`
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "invalid body")
		return
	}

	env, err := s.mesh.Broadcast(body.Type, body.Data)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, env)
}

func (s *Server) domainsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		writeJSON(w, 200, s.author.List())
	case "POST":
		var body struct {
			Domain string `json:"domain"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, 400, "invalid body")
			return
		}
		info, dns, err := s.author.Create(body.Domain)
		if err != nil {
			writeError(w, 409, err.Error())
			return
		}
		s.mesh.Broadcast("domain.claim", map[string]any{
			"domain":     body.Domain,
			"controller": info.Controller,
		})
		writeJSON(w, 200, map[string]any{
			"domain": info.Domain, "controller": info.Controller,
			"createdAt": info.CreatedAt, "dnsRecords": dns,
		})
	default:
		writeError(w, 405, "method not allowed")
	}
}

func (s *Server) domainByIDHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/authority/domains/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, 404, "domain not specified")
		return
	}
	domain := parts[0]

	if len(parts) >= 2 && parts[1] == "available" {
		writeJSON(w, 200, map[string]any{
			"domain": domain, "available": s.author.Available(domain),
			"scope": "active-known-mesh",
		})
		return
	}

	if len(parts) >= 2 && parts[1] == "releases" {
		switch r.Method {
		case "GET":
			releases, err := s.author.Releases(domain)
			if err != nil {
				writeError(w, 404, err.Error())
				return
			}
			writeJSON(w, 200, releases)
		case "POST":
			var manifest api.ReleaseManifest
			if err := json.NewDecoder(r.Body).Decode(&manifest); err != nil {
				writeError(w, 400, "invalid body")
				return
			}
			result, err := s.author.PublishRelease(domain, manifest)
			if err != nil {
				writeError(w, 404, err.Error())
				return
			}
			s.mesh.Broadcast("release.available", map[string]any{
				"domain": domain, "releaseId": result.CID, "cid": result.CID,
			})
			writeJSON(w, 200, result)
		default:
			writeError(w, 405, "method not allowed")
		}
		return
	}

	if len(parts) >= 2 && parts[1] == "delegate" {
		if r.Method != "POST" {
			writeError(w, 405, "method not allowed")
			return
		}
		var body struct {
			PublicKey    string   `json:"public_key"`
			Capabilities []string `json:"capabilities"`
			ExpiresAt    int64    `json:"expires_at"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, 400, "invalid body")
			return
		}
		grant, err := s.author.Delegate(domain, body.PublicKey, body.Capabilities, body.ExpiresAt)
		if err != nil {
			writeError(w, 404, err.Error())
			return
		}
		writeJSON(w, 200, grant)
		return
	}

	writeError(w, 404, "unknown endpoint")
}
