package authority

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/racore/god/internal/protocol"
	"github.com/racore/god/pkg/api"
)

type DelegateGrant struct {
	PublicKey    string   `json:"publicKey"`
	Capabilities []string `json:"capabilities"`
	ExpiresAt    int64    `json:"expiresAt"`
	CreatedAt    int64    `json:"createdAt"`
}

type DomainInfo struct {
	Domain     string           `json:"domain"`
	Controller string           `json:"controller"`
	CreatedAt  int64            `json:"createdAt"`
	Delegates  []DelegateGrant  `json:"delegates"`
	Releases   []ReleaseEntry   `json:"releases"`
}

type ReleaseEntry struct {
	Manifest api.ReleaseManifest `json:"manifest"`
	ReleaseID string            `json:"releaseId"`
}

type Authority struct {
	mu        sync.RWMutex
	dataDir   string
	path      string
	domains   map[string]*DomainInfo
	observed  map[string]ObservedClaim
}

type ObservedClaim struct {
	Domain     string `json:"domain"`
	Controller string `json:"controller"`
	NodeID     string `json:"nodeId"`
	Timestamp  int64  `json:"timestamp"`
}

func New(dataDir string) *Authority {
	return &Authority{
		dataDir:  dataDir,
		path:     filepath.Join(dataDir, "authorities.json"),
		domains:  make(map[string]*DomainInfo),
		observed: make(map[string]ObservedClaim),
	}
}

func (a *Authority) Load() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	data, err := os.ReadFile(a.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var store struct {
		Domains  map[string]*DomainInfo `json:"domains"`
		Observed map[string]ObservedClaim `json:"observed"`
	}
	if err := json.Unmarshal(data, &store); err != nil {
		return err
	}
	if store.Domains != nil {
		a.domains = store.Domains
	}
	if store.Observed != nil {
		a.observed = store.Observed
	}
	return nil
}

func (a *Authority) save() error {
	store := map[string]any{
		"domains":  a.domains,
		"observed": a.observed,
	}
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(a.path, data, 0600)
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func (a *Authority) Create(domain string) (*DomainInfo, []map[string]string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	domain = strings.ToLower(strings.Trim(domain, "."))

	if _, exists := a.domains[domain]; exists {
		return nil, nil, fmt.Errorf("domain already claimed")
	}
	if obs, seen := a.observed[domain]; seen {
		return nil, nil, fmt.Errorf("domain observed as claimed by %s", obs.NodeID)
	}

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	pub := priv.Public().(ed25519.PublicKey)
	controller := "did:key:" + base64.RawURLEncoding.EncodeToString(pub)

	pemPath := filepath.Join(a.dataDir, fmt.Sprintf("authority-%s.pem", domain))
	b, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal authority key for %s: %w", domain, err)
	}
	block := &pem.Block{Type: "PRIVATE KEY", Bytes: b}
	if err := writeFileAtomic(pemPath, pem.EncodeToMemory(block), 0600); err != nil {
		return nil, nil, fmt.Errorf("write authority key for %s: %w", domain, err)
	}

	info := &DomainInfo{
		Domain:     domain,
		Controller: controller,
		CreatedAt:  time.Now().UnixMilli(),
		Delegates:  make([]DelegateGrant, 0),
		Releases:   make([]ReleaseEntry, 0),
	}
	a.domains[domain] = info
	if err := a.save(); err != nil {
		log.Printf("authority: save after create %s: %v", domain, err)
	}

	dns := []map[string]string{
		{"type": "TXT", "name": "_racore-key." + domain, "value": controller},
		{"type": "TXT", "name": "_racore." + domain, "value": domain},
	}
	return info, dns, nil
}

func (a *Authority) List() []*DomainInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]*DomainInfo, 0, len(a.domains))
	for _, info := range a.domains {
		out = append(out, info)
	}
	return out
}

func (a *Authority) Available(domain string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	domain = strings.ToLower(strings.Trim(domain, "."))
	_, owned := a.domains[domain]
	_, seen := a.observed[domain]
	return !owned && !seen
}

func (a *Authority) ObserveClaim(domain, controller, nodeID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	domain = strings.ToLower(strings.Trim(domain, "."))
	if _, exists := a.domains[domain]; !exists {
		a.observed[domain] = ObservedClaim{
			Domain: domain, Controller: controller,
			NodeID: nodeID, Timestamp: time.Now().UnixMilli(),
		}
		if err := a.save(); err != nil {
			log.Printf("authority: save after observe %s: %v", domain, err)
		}
	}
}

func (a *Authority) PublishRelease(domain string, manifest api.ReleaseManifest) (*api.ReleaseManifest, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	domain = strings.ToLower(strings.Trim(domain, "."))
	info, ok := a.domains[domain]
	if !ok {
		return nil, fmt.Errorf("domain not claimed")
	}

	priv, err := a.loadDomainKey(domain)
	if err != nil {
		return nil, err
	}

	manifest.Protocol = "rcp/0.2"
	manifest.Domain = domain
	manifest.CreatedAt = time.Now().UnixMilli()

	if len(info.Releases) > 0 {
		manifest.Parent = info.Releases[len(info.Releases)-1].ReleaseID
	}

	canonical, err := canonicalJSON(manifest)
	if err != nil {
		return nil, err
	}

	sig := ed25519.Sign(priv, canonical)
	manifest.Signature = base64.RawURLEncoding.EncodeToString(sig)

	sum := sha256.Sum256(append(canonical, sig...))
	releaseID := "rcp2-" + hex.EncodeToString(sum[:16])

	entry := ReleaseEntry{Manifest: manifest, ReleaseID: releaseID}
	info.Releases = append(info.Releases, entry)
	if err := a.save(); err != nil {
		log.Printf("authority: save after release %s: %v", domain, err)
		return nil, err
	}

	return &manifest, nil
}

func (a *Authority) Releases(domain string) ([]ReleaseEntry, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	domain = strings.ToLower(strings.Trim(domain, "."))
	info, ok := a.domains[domain]
	if !ok {
		return nil, fmt.Errorf("domain not found")
	}
	return info.Releases, nil
}

func (a *Authority) Delegate(domain, publicKey string, capabilities []string, expiresAt int64) (*DelegateGrant, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	domain = strings.ToLower(strings.Trim(domain, "."))
	info, ok := a.domains[domain]
	if !ok {
		return nil, fmt.Errorf("domain not claimed")
	}

	grant := DelegateGrant{
		PublicKey:    publicKey,
		Capabilities: capabilities,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now().UnixMilli(),
	}
	info.Delegates = append(info.Delegates, grant)
	if err := a.save(); err != nil {
		log.Printf("authority: save after delegate %s: %v", domain, err)
		return nil, err
	}

	return &grant, nil
}

func (a *Authority) loadDomainKey(domain string) (ed25519.PrivateKey, error) {
	pemPath := filepath.Join(a.dataDir, fmt.Sprintf("authority-%s.pem", domain))
	data, err := os.ReadFile(pemPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("invalid pem for %s", domain)
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	priv, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not Ed25519")
	}
	return priv, nil
}

func canonicalJSON(v any) ([]byte, error) {
	return protocol.CanonicalJSON(v)
}
