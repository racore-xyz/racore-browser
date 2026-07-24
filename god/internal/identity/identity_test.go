package identity

import (
	"encoding/base64"
	"os"
	"testing"
)

func TestNewNodeIdentity(t *testing.T) {
	dir, err := os.MkdirTemp("", "gomesh-identity-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	id, err := NewNodeIdentity(dir)
	if err != nil {
		t.Fatalf("NewNodeIdentity: %v", err)
	}

	if id.NodeID() == "" {
		t.Fatal("empty nodeId")
	}

	if len(id.PublicKeyBytes()) != 32 {
		t.Fatalf("expected 32 bytes public key, got %d", len(id.PublicKeyBytes()))
	}

	// should load existing key
	id2, err := NewNodeIdentity(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}

	if id.NodeID() != id2.NodeID() {
		t.Fatal("nodeId mismatch after reload")
	}
}

func TestSignVerify(t *testing.T) {
	dir, err := os.MkdirTemp("", "gomesh-identity-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	id, err := NewNodeIdentity(dir)
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte(`{"hello":"world"}`)
	sig, err := id.Sign(payload)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	pub, err := Verify(id.PublicKeyBase64(), payload, sig)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}

	if len(pub) != 32 {
		t.Fatal("invalid public key from verify")
	}
}

func TestNodeIDFromPublicKey(t *testing.T) {
	dir, err := os.MkdirTemp("", "gomesh-identity-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	id, err := NewNodeIdentity(dir)
	if err != nil {
		t.Fatal(err)
	}

	derived := NodeIDFromPublicKey(id.PublicKeyBytes())
	if derived != id.NodeID() {
		t.Fatal("NodeIDFromPublicKey mismatch")
	}
}

func TestTamperedSignature(t *testing.T) {
	dir, err := os.MkdirTemp("", "gomesh-identity-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	id, err := NewNodeIdentity(dir)
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte("test")
	sig, err := id.Sign(payload)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Verify(id.PublicKeyBase64(), []byte("tampered"), sig); err == nil {
		t.Fatal("expected verify to fail for tampered payload")
	}
}

func TestVerifyRejectsMalformedKeyAndSignature(t *testing.T) {
	dir, err := os.MkdirTemp("", "gomesh-identity-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	id, err := NewNodeIdentity(dir)
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte("test")
	sig, err := id.Sign(payload)
	if err != nil {
		t.Fatal(err)
	}

	shortKey := base64.RawURLEncoding.EncodeToString([]byte("too-short"))
	if _, err := Verify(shortKey, payload, sig); err == nil {
		t.Fatal("expected error for wrong-length public key")
	}

	shortSig := base64.RawURLEncoding.EncodeToString([]byte("bad"))
	if _, err := Verify(id.PublicKeyBase64(), payload, shortSig); err == nil {
		t.Fatal("expected error for wrong-length signature")
	}
}
