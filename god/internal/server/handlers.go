package server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/racore/god/internal/authority"
	"github.com/racore/god/internal/ipfs"
	"github.com/racore/god/pkg/api"
)

type releaseView struct {
	Version   string `json:"version"`
	CID       string `json:"cid"`
	ReleaseID string `json:"releaseId"`
	CreatedAt int64  `json:"createdAt"`
	Files     int    `json:"files"`
	Size      int    `json:"size"`
	Verified  bool   `json:"verified"`
}

type domainView struct {
	Domain     string        `json:"domain"`
	Controller string        `json:"controller"`
	CreatedAt  int64         `json:"createdAt"`
	Status     string        `json:"status"`
	Current    string        `json:"current,omitempty"`
	Releases   []releaseView `json:"releases"`
	Delegates  int           `json:"delegates"`
}

func validateReleaseManifest(domain string, manifest api.ReleaseManifest) error {
	if !authority.IsRacoreDomain(domain) {
		return fmt.Errorf("unsupported Racore domain")
	}
	version := strings.TrimSpace(manifest.Version)
	if version == "" || len(version) > 64 || strings.ContainsAny(version, " \t\r\n") {
		return fmt.Errorf("invalid release version")
	}
	if !authority.IsSafeCID(manifest.CID) {
		return fmt.Errorf("invalid content identifier")
	}
	if len(manifest.ContentRoot) != 64 {
		return fmt.Errorf("content root must be a SHA-256 hex digest")
	}
	if _, err := hex.DecodeString(manifest.ContentRoot); err != nil {
		return fmt.Errorf("content root must be a SHA-256 hex digest")
	}
	entrypoint := strings.ReplaceAll(strings.TrimSpace(manifest.Entrypoint), "\\", "/")
	if entrypoint == "" ||
		strings.HasPrefix(entrypoint, "/") ||
		strings.Contains(entrypoint, "..") {
		return fmt.Errorf("invalid entrypoint")
	}
	if manifest.Files < 1 || manifest.Size < 1 {
		return fmt.Errorf("release must contain at least one byte and one file")
	}
	return nil
}

func (s *Server) shutdownHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "shuttingDown": true})
	s.once.Do(func() { close(s.done) })
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	ipfsStatus, _ := s.ipfs.Status(r.Context())
	writeJSON(w, 200, api.HealthResponse{
		OK:      true,
		Version: s.version,
		Mesh:    s.mesh.Status(),
		IPFS:    ipfsStatus,
		LocalAI: s.localAI.Status(r.Context()),
	})
}

func (s *Server) localAIStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}
	writeJSON(w, 200, s.localAI.Status(r.Context()))
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

	s.hub.Broadcast(map[string]any{"type": "agent.started", "engine": "local"})

	resp, err := s.localAI.Chat(r.Context(), req.Messages, req.System)
	if err != nil {
		writeError(w, 502, err.Error())
		return
	}

	s.hub.Broadcast(map[string]any{
		"type": "agent.completed", "engine": resp.Engine,
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

	if err := r.ParseMultipartForm(256 << 20); err != nil {
		writeError(w, 400, "multipart parse error")
		return
	}

	headers := r.MultipartForm.File["file"]
	if len(headers) == 0 {
		writeError(w, 400, "file required")
		return
	}

	// Multipart filenames are reduced to their base name by the server-side
	// parser, so publish clients send the original relative paths separately.
	var paths []string
	if values := r.MultipartForm.Value["paths"]; len(values) > 0 {
		if err := json.Unmarshal([]byte(values[0]), &paths); err != nil {
			writeError(w, 400, "invalid paths manifest")
			return
		}
	}

	files := make([]ipfs.File, 0, len(headers))
	for index, header := range headers {
		file, err := header.Open()
		if err != nil {
			writeError(w, 400, "read error")
			return
		}
		data, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			writeError(w, 400, "read error")
			return
		}
		name := header.Filename
		if index < len(paths) && paths[index] != "" {
			name = paths[index]
		}
		files = append(files, ipfs.File{Name: name, Data: data})
	}

	result, err := s.ipfs.AddFiles(r.Context(), files)
	if err != nil {
		writeError(w, 503, "kubo unavailable: "+err.Error())
		return
	}

	s.mesh.Broadcast("ipfs.added", map[string]any{
		"cid": result["cid"], "files": result["files"],
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
		items := s.author.List()
		out := make([]domainView, 0, len(items))
		for _, info := range items {
			view := domainView{
				Domain:     info.Domain,
				Controller: info.Controller,
				CreatedAt:  info.CreatedAt,
				Status:     "claimed",
				Releases:   make([]releaseView, 0, len(info.Releases)),
				Delegates:  len(info.Delegates),
			}
			for _, release := range info.Releases {
				view.Releases = append(view.Releases, releaseView{
					Version:   release.Manifest.Version,
					CID:       release.Manifest.CID,
					ReleaseID: release.ReleaseID,
					CreatedAt: release.Manifest.CreatedAt,
					Files:     release.Manifest.Files,
					Size:      release.Manifest.Size,
					Verified:  authority.VerifyRelease(info.Controller, release.ReleaseID, release.Manifest) == nil,
				})
			}
			if len(info.Releases) > 0 {
				latest := info.Releases[len(info.Releases)-1]
				view.Status = "live"
				view.Current = latest.Manifest.CID
			}
			out = append(out, view)
		}
		writeJSON(w, 200, out)
	case "POST":
		var body struct {
			Domain string `json:"domain"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, 400, "invalid body")
			return
		}
		domain := strings.ToLower(strings.TrimSpace(body.Domain))
		if !authority.IsRacoreDomain(domain) {
			writeError(w, 400, "use a .racore, .rac, .core, or .ra domain")
			return
		}
		info, dns, err := s.author.Create(domain)
		if err != nil {
			writeError(w, 409, err.Error())
			return
		}
		s.mesh.Broadcast("domain.claim", map[string]any{
			"domain":     info.Domain,
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

func (s *Server) networkDomainsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}
	writeJSON(w, 200, s.author.NetworkDomains())
}

func (s *Server) resolveDomainHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}
	domain := strings.TrimPrefix(r.URL.Path, "/v1/authority/resolve/")
	resolved, err := s.author.ResolveVerifiedNetworkDomain(domain)
	if err != nil {
		writeError(w, 422, err.Error())
		return
	}
	payload := map[string]any{
		"domain": resolved.Domain, "controller": resolved.Controller,
		"source": resolved.Source, "nodeId": resolved.NodeID,
		"releaseId": resolved.ReleaseID, "cid": resolved.CID,
		"verified":   true,
		"gatewayUrl": strings.TrimRight(s.cfg.IPFSGateway, "/") + "/ipfs/" + resolved.CID + "/",
	}
	if resolved.Manifest != nil {
		payload["version"] = resolved.Manifest.Version
		payload["entrypoint"] = resolved.Manifest.Entrypoint
		payload["manifest"] = resolved.Manifest
	}
	writeJSON(w, 200, payload)
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
			if err := validateReleaseManifest(domain, manifest); err != nil {
				writeError(w, 400, err.Error())
				return
			}
			result, err := s.author.PublishRelease(domain, manifest)
			if err != nil {
				writeError(w, 404, err.Error())
				return
			}
			releases, _ := s.author.Releases(domain)
			entry := releases[len(releases)-1]
			s.mesh.Broadcast("release.available", map[string]any{
				"domain": domain, "releaseId": entry.ReleaseID,
				"cid": entry.Manifest.CID, "manifest": entry.Manifest,
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
