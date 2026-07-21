package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewVault(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-vault-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	v := New(dir)
	if err := v.Load(); err != nil {
		t.Fatal(err)
	}
}

func TestSetGet(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-vault-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	v := New(dir)
	v.Load()

	if err := v.Set("openai", "sk-test123456789"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	val, ok := v.Get("openai")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if val != "sk-test123456789" {
		t.Fatalf("expected sk-test123456789, got %s", val)
	}
}

func TestGetMissing(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-vault-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	v := New(dir)
	v.Load()

	_, ok := v.Get("nonexistent")
	if ok {
		t.Fatal("expected missing key")
	}
}

func TestDelete(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-vault-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	v := New(dir)
	v.Load()
	v.Set("test", "secret123")
	v.Delete("test")

	_, ok := v.Get("test")
	if ok {
		t.Fatal("expected key to be deleted")
	}
}

func TestConnected(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-vault-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	v := New(dir)
	v.Load()
	v.Set("a", "1")
	v.Set("b", "2")
	v.Set("c", "3")

	conn := v.Connected()
	if len(conn) != 3 {
		t.Fatalf("expected 3, got %d", len(conn))
	}
	// Should be sorted
	if conn[0] != "a" || conn[1] != "b" || conn[2] != "c" {
		t.Fatalf("expected sorted [a b c], got %v", conn)
	}
}

func TestMasked(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-vault-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	v := New(dir)
	v.Load()

	v.Set("test", "sk-test123456789")
	masked := v.Masked("test")
	if masked != "sk-...6789" {
		t.Fatalf("expected sk-...6789, got %s", masked)
	}

	maskedMissing := v.Masked("missing")
	if maskedMissing != "" {
		t.Fatalf("expected empty, got %s", maskedMissing)
	}

	v.Set("short", "ab")
	shortMasked := v.Masked("short")
	if shortMasked != "***" {
		t.Fatalf("expected ***, got %s", shortMasked)
	}
}

func TestPersistence(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-vault-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	v1 := New(dir)
	v1.Load()
	v1.Set("persist", "keep-me")

	v2 := New(dir)
	v2.Load()

	val, ok := v2.Get("persist")
	if !ok {
		t.Fatal("expected persisted key to exist")
	}
	if val != "keep-me" {
		t.Fatalf("expected keep-me, got %s", val)
	}
}

func TestKeyFileCreated(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-vault-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	v := New(dir)
	v.Load()

	keyPath := filepath.Join(dir, ".vault-key")
	if _, err := os.Stat(keyPath); err != nil {
		t.Fatalf("key file not created: %v", err)
	}

	vaultPath := filepath.Join(dir, "credentials.vault")
	if _, err := os.Stat(vaultPath); os.IsExist(err) {
		t.Fatal("vault file should not exist before first save")
	}

	v.Set("x", "y")
	if _, err := os.Stat(vaultPath); err != nil {
		t.Fatalf("vault file should exist after save: %v", err)
	}
}

func TestConcurrentAccess(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-vault-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	v := New(dir)
	v.Load()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			key := string(rune('a' + n))
			v.Set(key, "val")
			v.Get(key)
			v.Connected()
			v.Masked(key)
			done <- true
		}(i)
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}
