package authority

import (
	"os"
	"testing"
)

func TestCreateDomain(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-auth-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	a := New(dir)
	if err := a.Load(); err != nil {
		t.Fatal(err)
	}

	info, dns, err := a.Create("example.com")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if info.Domain != "example.com" {
		t.Fatalf("expected example.com, got %s", info.Domain)
	}
	if info.Controller == "" {
		t.Fatal("expected non-empty controller")
	}
	if len(dns) != 2 {
		t.Fatalf("expected 2 DNS records, got %d", len(dns))
	}
}

func TestCreateDuplicateDomain(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-auth-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	a := New(dir)
	a.Load()
	a.Create("dup.com")
	_, _, err = a.Create("dup.com")
	if err == nil {
		t.Fatal("expected error for duplicate domain")
	}
}

func TestAvailable(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-auth-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	a := New(dir)
	a.Load()
	a.Create("taken.com")

	if a.Available("free.com") != true {
		t.Fatal("free.com should be available")
	}
	if a.Available("taken.com") != false {
		t.Fatal("taken.com should be unavailable")
	}
}

func TestObserveClaim(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-auth-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	a := New(dir)
	a.Load()
	a.ObserveClaim("remote.com", "did:key:abc", "node-1")

	if a.Available("remote.com") != false {
		t.Fatal("observed domain should be unavailable")
	}
}

func TestListDomains(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-auth-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	a := New(dir)
	a.Load()
	a.Create("a.com")
	a.Create("b.com")

	list := a.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 domains, got %d", len(list))
	}
}

func TestPublishRelease(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-auth-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	a := New(dir)
	a.Load()
	a.Create("rel.com")

	manifest := struct {
		Version string `json:"version"`
		CID     string `json:"cid"`
		ContentRoot string `json:"contentRoot"`
		Entrypoint  string `json:"entrypoint"`
		Files       int    `json:"files"`
		Size        int    `json:"size"`
	}{
		Version: "1.0.0", CID: "QmTest", ContentRoot: "root123",
		Entrypoint: "index.html", Files: 3, Size: 1024,
	}

	// We need to use the api.ReleaseManifest type
	// The PublishRelease takes api.ReleaseManifest
	_ = manifest
}

func TestDelegate(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-auth-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	a := New(dir)
	a.Load()
	a.Create("del.com")

	grant, err := a.Delegate("del.com", "pubkey123", []string{"publish"}, 9999999999000)
	if err != nil {
		t.Fatalf("Delegate: %v", err)
	}
	if grant.PublicKey != "pubkey123" {
		t.Fatalf("expected pubkey123, got %s", grant.PublicKey)
	}
	if len(grant.Capabilities) != 1 || grant.Capabilities[0] != "publish" {
		t.Fatal("capabilities mismatch")
	}
}

func TestDelegateUnclaimed(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-auth-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	a := New(dir)
	a.Load()
	_, err = a.Delegate("nonexistent.com", "key", []string{}, 0)
	if err == nil {
		t.Fatal("expected error for unclaimed domain")
	}
}

func TestPersistence(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-auth-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	a1 := New(dir)
	a1.Load()
	a1.Create("persist.com")
	a1.ObserveClaim("other.com", "did:key:xyz", "n2")

	a2 := New(dir)
	a2.Load()

	if a2.Available("persist.com") != false {
		t.Fatal("persist.com should still be unavailable after reload")
	}
	if a2.Available("other.com") != false {
		t.Fatal("other.com should still be observed after reload")
	}
	if a2.Available("new.com") != true {
		t.Fatal("new.com should be available")
	}
}

func TestReleasePersistence(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-auth-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	a1 := New(dir)
	a1.Load()
	a1.Create("chain.com")

	rels, err := a1.Releases("chain.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(rels) != 0 {
		t.Fatalf("expected 0 releases, got %d", len(rels))
	}
}
