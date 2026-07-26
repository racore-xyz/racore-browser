package api

type Config struct {
	Host             string   `json:"host"`
	Port             int      `json:"port"`
	NodeName         string   `json:"nodeName"`
	MeshEnabled      bool     `json:"meshEnabled"`
	MeshGroup        string   `json:"meshGroup"`
	MeshPort         int      `json:"meshPort"`
	MeshAPIPort      int      `json:"meshApiPort"`
	MeshHeartbeatSec int      `json:"meshHeartbeatSeconds"`
	IPFSAPI          string   `json:"ipfsApi"`
	IPFSGateway      string   `json:"ipfsGateway"`
	BootstrapPeers   []string `json:"bootstrapPeers"`
	DataDir          string   `json:"dataDir"`
}

type ChatRequest struct {
	Messages []map[string]string `json:"messages"`
	System   string              `json:"system,omitempty"`
}

type BrowserAction struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type ChatResponse struct {
	Text      string          `json:"text"`
	Model     string          `json:"model"`
	Engine    string          `json:"engine"`
	LatencyMs int64           `json:"latencyMs"`
	Actions   []BrowserAction `json:"actions,omitempty"`
}

type LocalAIStatus struct {
	Ready        bool   `json:"ready"`
	State        string `json:"state"`
	Model        string `json:"model"`
	Label        string `json:"label"`
	Engine       string `json:"engine"`
	Parameters   int64  `json:"parameters"`
	Quantization string `json:"quantization"`
	LocalOnly    bool   `json:"localOnly"`
	Error        string `json:"error,omitempty"`
}

type HealthResponse struct {
	OK      bool          `json:"ok"`
	Version string        `json:"version"`
	Mesh    any           `json:"mesh,omitempty"`
	IPFS    any           `json:"ipfs,omitempty"`
	LocalAI LocalAIStatus `json:"localAI"`
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
