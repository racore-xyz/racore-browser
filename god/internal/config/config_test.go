package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDataDirDefault(t *testing.T) {
	dir := DataDir()
	if dir == "" {
		t.Fatal("expected non-empty data dir")
	}
}

func TestDataDirEnvOverride(t *testing.T) {
	os.Setenv("RACORE_DATA_DIR", "/tmp/racore-test-data")
	defer os.Unsetenv("RACORE_DATA_DIR")

	dir := DataDir()
	if dir != "/tmp/racore-test-data" {
		t.Fatalf("expected /tmp/racore-test-data, got %s", dir)
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"47831", 47831},
		{"0", 0},
		{"abc", 0},
		{"", 0},
		{"999", 999},
	}
	for _, tt := range tests {
		got := parseInt(tt.input)
		if got != tt.want {
			t.Errorf("parseInt(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestLoadDefaults(t *testing.T) {
	orig := os.Getenv("RACORE_DATA_DIR")
	os.Unsetenv("RACORE_DATA_DIR")
	defer os.Setenv("RACORE_DATA_DIR", orig)

	td, _ := os.MkdirTemp("", "racore-config-*")
	defer os.RemoveAll(td)
	os.Setenv("RACORE_DATA_DIR", td)

	cfg := Load()
	if cfg.Port != 47831 {
		t.Fatalf("expected 47831, got %d", cfg.Port)
	}
	if cfg.MeshPort != 47777 {
		t.Fatalf("expected 47777, got %d", cfg.MeshPort)
	}
	if cfg.MeshAPIPort != 47833 {
		t.Fatalf("expected 47833, got %d", cfg.MeshAPIPort)
	}
	if !cfg.MeshEnabled {
		t.Fatal("expected mesh enabled")
	}
}

func TestLoadFromFile(t *testing.T) {
	td, _ := os.MkdirTemp("", "racore-config-*")
	defer os.RemoveAll(td)

	os.Setenv("RACORE_DATA_DIR", td)
	defer os.Unsetenv("RACORE_DATA_DIR")

	settingsPath := filepath.Join(td, "settings.json")
	os.WriteFile(settingsPath, []byte(`{"port":47999,"nodeName":"CustomNode"}`), 0600)

	cfg := Load()
	if cfg.Port != 47999 {
		t.Fatalf("expected 47999, got %d", cfg.Port)
	}
	if cfg.NodeName != "CustomNode" {
		t.Fatalf("expected CustomNode, got %s", cfg.NodeName)
	}
}

func TestEnvOverride(t *testing.T) {
	td, _ := os.MkdirTemp("", "racore-config-*")
	defer os.RemoveAll(td)

	os.Setenv("RACORE_DATA_DIR", td)
	os.Setenv("RACORE_PORT", "47777")
	os.Setenv("RACORE_NODE_NAME", "EnvNode")
	os.Setenv("RACORE_MESH_PORT", "47888")
	defer func() {
		os.Unsetenv("RACORE_DATA_DIR")
		os.Unsetenv("RACORE_PORT")
		os.Unsetenv("RACORE_NODE_NAME")
		os.Unsetenv("RACORE_MESH_PORT")
	}()

	cfg := Load()
	if cfg.Port != 47777 {
		t.Fatalf("expected 47777, got %d", cfg.Port)
	}
	if cfg.NodeName != "EnvNode" {
		t.Fatalf("expected EnvNode, got %s", cfg.NodeName)
	}
	if cfg.MeshPort != 47888 {
		t.Fatalf("expected 47888, got %d", cfg.MeshPort)
	}
}
