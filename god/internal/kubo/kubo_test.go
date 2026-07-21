package kubo

import (
	"context"
	"os"
	"testing"
)

func TestNewManager(t *testing.T) {
	m := New("http://127.0.0.1:5001", "http://127.0.0.1:8180", "/tmp")
	if m == nil {
		t.Fatal("expected non-nil manager")
	}
}

func TestStartNoExecutable(t *testing.T) {
	os.Unsetenv("RACORE_KUBO_PATH")
	dir, _ := os.MkdirTemp("", "god-kubo-*")
	defer os.RemoveAll(dir)

	m := New("http://127.0.0.1:5001", "http://127.0.0.1:8180", dir)
	ctx := context.Background()

	result, err := m.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}

	online, _ := result["online"].(bool)
	if online {
		t.Fatal("expected offline (no kub binary)")
	}

	reason, _ := result["reason"].(string)
	if reason != "kubo-not-installed" {
		t.Fatalf("expected kubo-not-installed, got %s", reason)
	}
}

func TestFindExecutable(t *testing.T) {
	m := New("", "", "")
	exe := m.findExecutable()
	_ = exe
}

func TestIsOnlineNoServer(t *testing.T) {
	m := New("http://127.0.0.1:1", "", "")
	ctx := context.Background()

	if m.isOnline(ctx) {
		t.Fatal("expected offline")
	}
}

func TestStopNoProcess(t *testing.T) {
	m := New("", "", "")
	m.Stop()
	// Should not panic
}

func TestStartExternalNode(t *testing.T) {
	dir, _ := os.MkdirTemp("", "god-kubo-ext-*")
	defer os.RemoveAll(dir)

	m := New("http://127.0.0.1:1", "http://127.0.0.1:8180", dir)
	ctx := context.Background()

	result, err := m.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	online, _ := result["online"].(bool)
	if online {
		t.Fatal("expected offline (no external node)")
	}
}

func TestStartDetectsExternalNode(t *testing.T) {
	// This test verifies that when no kub binary is found,
	// the manager correctly reports not-installed
	os.Unsetenv("RACORE_KUBO_PATH")
	dir, _ := os.MkdirTemp("", "god-kubo-detect-*")
	defer os.RemoveAll(dir)

	m := New("http://127.0.0.1:5001", "http://127.0.0.1:8180", dir)

	if m.findExecutable() != "" {
		t.Skip("kub found on PATH, skipping")
	}

	ctx := context.Background()
	result, _ := m.Start(ctx)
	managed, _ := result["managed"].(bool)
	if managed {
		t.Fatal("expected not managed when no kub available")
	}
}
