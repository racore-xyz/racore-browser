package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

type NodeIdentity struct {
	nodeID     string
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

func NewNodeIdentity(root string) (*NodeIdentity, error) {
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	path := filepath.Join(root, "mesh-identity.pem")
	var privateKey ed25519.PrivateKey
	if data, err := os.ReadFile(path); err == nil {
		block, _ := pem.Decode(data)
		if block == nil {
			return nil, fmt.Errorf("invalid pem in %s", path)
		}
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		var ok bool
		privateKey, ok = key.(ed25519.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("key is not Ed25519")
		}
	} else {
		_, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("generate key: %w", err)
		}
		privateKey = priv
		b, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			return nil, fmt.Errorf("marshal key: %w", err)
		}
		block := &pem.Block{Type: "PRIVATE KEY", Bytes: b}
		if err := os.WriteFile(path, pem.EncodeToMemory(block), 0600); err != nil {
			return nil, fmt.Errorf("write key: %w", err)
		}
	}
	publicKey := privateKey.Public().(ed25519.PublicKey)
	hash := sha256.Sum256(publicKey)
	nodeID := hex.EncodeToString(hash[:16])
	return &NodeIdentity{
		nodeID:     nodeID,
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

func (id *NodeIdentity) NodeID() string {
	return id.nodeID
}

func (id *NodeIdentity) PublicKeyBytes() []byte {
	return id.publicKey
}

func (id *NodeIdentity) PublicKeyBase64() string {
	return base64.RawURLEncoding.EncodeToString(id.publicKey)
}

func (id *NodeIdentity) Sign(payload []byte) (string, error) {
	sig := ed25519.Sign(id.privateKey, payload)
	return base64.RawURLEncoding.EncodeToString(sig), nil
}

func Verify(publicKeyB64 string, payload []byte, signatureB64 string) (ed25519.PublicKey, error) {
	pubBytes, err := base64.RawURLEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return nil, fmt.Errorf("decode public key: %w", err)
	}
	sig, err := base64.RawURLEncoding.DecodeString(signatureB64)
	if err != nil {
		return nil, fmt.Errorf("decode signature: %w", err)
	}
	pub := ed25519.PublicKey(pubBytes)
	if !ed25519.Verify(pub, payload, sig) {
		return nil, fmt.Errorf("invalid signature")
	}
	return pub, nil
}

func NodeIDFromPublicKey(pub ed25519.PublicKey) string {
	hash := sha256.Sum256(pub)
	return hex.EncodeToString(hash[:16])
}
