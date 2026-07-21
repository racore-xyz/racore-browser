package api

type Config struct {
	Host               string   `json:"host"`
	Port               int      `json:"port"`
	NodeName           string   `json:"nodeName"`
	MeshEnabled        bool     `json:"meshEnabled"`
	MeshGroup          string   `json:"meshGroup"`
	MeshPort           int      `json:"meshPort"`
	MeshAPIPort        int      `json:"meshApiPort"`
	MeshHeartbeatSec   int      `json:"meshHeartbeatSeconds"`
	IPFSAPI            string   `json:"ipfsApi"`
	IPFSGateway        string   `json:"ipfsGateway"`
	BootstrapPeers     []string `json:"bootstrapPeers"`
	DataDir            string   `json:"dataDir"`
}

type Provider struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	BaseURL     string `json:"baseURL"`
	DefaultModel string `json:"defaultModel"`
	Auth        string `json:"auth"`
	Free        bool   `json:"free"`
	Local       bool   `json:"local"`
	Connected   bool   `json:"connected"`
	MaskedKey   string `json:"maskedKey"`
}

type ChatRequest struct {
	Provider string              `json:"provider"`
	Model    string              `json:"model,omitempty"`
	Messages []map[string]string `json:"messages"`
	System   string              `json:"system,omitempty"`
}

type ChatResponse struct {
	Text      string `json:"text"`
	Usage     any    `json:"usage,omitempty"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	LatencyMs int64  `json:"latencyMs"`
	RawID     string `json:"rawId,omitempty"`
}

type HealthResponse struct {
	OK        bool        `json:"ok"`
	Version   string      `json:"version"`
	Mesh      any         `json:"mesh,omitempty"`
	IPFS      any         `json:"ipfs,omitempty"`
	Providers int         `json:"providers"`
}

type DomainRecord struct {
	Domain     string `json:"domain"`
	Controller string `json:"controller"`
	CreatedAt  int64  `json:"createdAt"`
	Status     string `json:"status"`
}

type ReleaseManifest struct {
	Protocol    string `json:"protocol"`
	Domain      string `json:"domain"`
	Version     string `json:"version"`
	CID         string `json:"cid"`
	ContentRoot string `json:"contentRoot"`
	Entrypoint  string `json:"entrypoint"`
	Files       int    `json:"files"`
	Size        int    `json:"size"`
	CreatedAt   int64  `json:"createdAt"`
	Parent      string `json:"parent,omitempty"`
	Signature   string `json:"signature"`
}

type DelegateGrant struct {
	PublicKey    string   `json:"publicKey"`
	Capabilities []string `json:"capabilities"`
	ExpiresAt    int64    `json:"expiresAt"`
	CreatedAt    int64    `json:"createdAt"`
}
