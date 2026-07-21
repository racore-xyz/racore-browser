package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"github.com/racore/god/pkg/api"
)

func Load() api.Config {
	cfg := api.Config{
		Host:             "127.0.0.1",
		Port:             47831,
		NodeName:         "Racore Desktop",
		MeshEnabled:      true,
		MeshGroup:        "239.255.77.77",
		MeshPort:         47777,
		MeshAPIPort:      47833,
		MeshHeartbeatSec: 5,
		IPFSAPI:          "http://127.0.0.1:5001",
		IPFSGateway:      "http://127.0.0.1:8180",
		DataDir:          DataDir(),
	}

	path := filepath.Join(cfg.DataDir, "settings.json")
	if data, err := os.ReadFile(path); err == nil {
		var raw map[string]any
		if json.Unmarshal(data, &raw) == nil {
			if v, ok := raw["ipfs_gateway"].(string); ok && v == "http://127.0.0.1:8080" {
				raw["ipfs_gateway"] = "http://127.0.0.1:8180"
			}
			b, _ := json.Marshal(raw)
			json.Unmarshal(b, &cfg)
		}
	} else {
		os.MkdirAll(cfg.DataDir, 0700)
		b, _ := json.MarshalIndent(cfg, "", "  ")
		os.WriteFile(path, b, 0600)
	}

	applyEnv(&cfg)
	return cfg
}

func applyEnv(cfg *api.Config) {
	if v := os.Getenv("RACORE_PORT"); v != "" {
		if p := parseInt(v); p > 0 {
			cfg.Port = p
		}
	}
	if v := os.Getenv("RACORE_NODE_NAME"); v != "" {
		cfg.NodeName = v
	}
	if v := os.Getenv("RACORE_MESH_PORT"); v != "" {
		if p := parseInt(v); p > 0 {
			cfg.MeshPort = p
		}
	}
	if v := os.Getenv("RACORE_DATA_DIR"); v != "" {
		cfg.DataDir = v
	}
}

func DataDir() string {
	if v := os.Getenv("RACORE_DATA_DIR"); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "./data"
	}
	if runtime.GOOS == "windows" {
		app := os.Getenv("APPDATA")
		if app != "" {
			return filepath.Join(app, "Racore")
		}
		return filepath.Join(home, "AppData", "Roaming", "Racore")
	}
	xdg := os.Getenv("XDG_DATA_HOME")
	if xdg == "" {
		xdg = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(xdg, "racore")
}

func parseInt(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}
