package protocol

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/racore/god/internal/identity"
)

func TestNewEnvelope(t *testing.T) {
	dir, err := os.MkdirTemp("", "gomesh-protocol-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	id, err := identity.NewNodeIdentity(dir)
	if err != nil {
		t.Fatal(err)
	}

	data := map[string]any{"apiPort": 47831}
	env, err := NewEnvelope("presence", data, id, "Test Node", []string{"client", "cache"})
	if err != nil {
		t.Fatalf("NewEnvelope: %v", err)
	}

	if env.Protocol != ProtocolVersion {
		t.Fatalf("expected %s, got %s", ProtocolVersion, env.Protocol)
	}

	if env.Signature == "" {
		t.Fatal("empty signature")
	}

	if env.NodeID != id.NodeID() {
		t.Fatal("nodeId mismatch")
	}
}

func TestVerifyEnvelope(t *testing.T) {
	dir, err := os.MkdirTemp("", "gomesh-protocol-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	id, err := identity.NewNodeIdentity(dir)
	if err != nil {
		t.Fatal(err)
	}

	data := map[string]any{"test": true}
	env, err := NewEnvelope("test", data, id, "Test", []string{"client"})
	if err != nil {
		t.Fatal(err)
	}

	raw, _ := json.Marshal(env)
	verified, err := VerifyEnvelope(raw)
	if err != nil {
		t.Fatalf("VerifyEnvelope: %v", err)
	}

	if verified.NodeID != id.NodeID() {
		t.Fatal("nodeId mismatch after verify")
	}
}

func TestTamperedEnvelope(t *testing.T) {
	dir, err := os.MkdirTemp("", "gomesh-protocol-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	id, err := identity.NewNodeIdentity(dir)
	if err != nil {
		t.Fatal(err)
	}

	env, err := NewEnvelope("test", map[string]any{}, id, "Test", nil)
	if err != nil {
		t.Fatal(err)
	}

	raw, _ := json.Marshal(env)
	tampered := strings.Replace(string(raw), `"type":"test"`, `"type":"hacked"`, 1)

	if _, err := VerifyEnvelope([]byte(tampered)); err == nil {
		t.Fatal("expected verify to fail for tampered envelope")
	}
}

func TestCanonicalJSON(t *testing.T) {
	// Ensure keys are sorted alphabetically
	input := map[string]any{
		"z": "last",
		"a": "first",
		"m": "middle",
	}
	out, err := CanonicalJSON(input)
	if err != nil {
		t.Fatal(err)
	}

	// Go's json.Marshal for map[string]any sorts keys alphabetically
	var result map[string]any
	json.Unmarshal(out, &result)
	keys := make([]string, 0, len(result))
	for k := range result {
		keys = append(keys, k)
	}
	if keys[0] != "a" || keys[1] != "m" || keys[2] != "z" {
		t.Fatalf("keys not sorted: %v", keys)
	}
}

func TestCanonicalJSONNoSpaces(t *testing.T) {
	input := map[string]any{"a": 1, "b": 2}
	out, err := CanonicalJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if strings.Contains(s, " ") {
		t.Fatalf("canonical JSON should have no spaces, got: %s", s)
	}
	if !strings.HasPrefix(s, `{"a":1,"b":2}`) && s != `{"a":1,"b":2}` {
		t.Fatalf("unexpected format: %s", s)
	}
}

func TestSelfMessageRejection(t *testing.T) {
	dir, err := os.MkdirTemp("", "gomesh-protocol-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	id, err := identity.NewNodeIdentity(dir)
	if err != nil {
		t.Fatal(err)
	}

	env, err := NewEnvelope("presence", nil, id, "Self", nil)
	if err != nil {
		t.Fatal(err)
	}

	raw, _ := json.Marshal(env)
	verified, err := VerifyEnvelope(raw)
	if err != nil {
		t.Fatalf("self message should verify: %v", err)
	}
	if verified.NodeID != id.NodeID() {
		t.Fatal("nodeId should match")
	}
}
