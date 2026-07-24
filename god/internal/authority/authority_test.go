package authority

import (
	"os"
	"strings"
	"testing"

	"github.com/racore/god/pkg/api"
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

	published, err := a.PublishRelease("rel.com", api.ReleaseManifest{
		Version: "1.0.0", CID: "QmTest", ContentRoot: "root123",
		Entrypoint: "index.html", Files: 3, Size: 1024,
	})
	if err != nil {
		t.Fatalf("PublishRelease: %v", err)
	}
	if published.Signature == "" {
		t.Fatal("expected signed manifest")
	}

	rels, err := a.Releases("rel.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(rels) != 1 {
		t.Fatalf("expected 1 release, got %d", len(rels))
	}

	id := rels[0].ReleaseID
	if !strings.HasPrefix(id, "rcp2-") {
		t.Fatalf("expected rcp2- prefix, got %s", id)
	}
	if len(id) != len("rcp2-")+32 {
		t.Fatalf("expected 32 hex chars after prefix, got %s", id)
	}

	second, err := a.PublishRelease("rel.com", api.ReleaseManifest{
		Version: "2.0.0", CID: "QmTest2", ContentRoot: "root456",
		Entrypoint: "index.html", Files: 4, Size: 2048,
	})
	if err != nil {
		t.Fatalf("PublishRelease second: %v", err)
	}
	if second.Parent != id {
		t.Fatalf("expected parent %s, got %s", id, second.Parent)
	}

	rels, _ = a.Releases("rel.com")
	if rels[1].ReleaseID == id {
		t.Fatal("distinct releases must yield distinct release IDs")
	}
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
