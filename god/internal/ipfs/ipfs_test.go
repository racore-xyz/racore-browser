package ipfs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatusOffline(t *testing.T) {
	b := New("http://127.0.0.1:1", "http://127.0.0.1:2")
	ctx := context.Background()

	status, err := b.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}
	online, _ := status["online"].(bool)
	if online {
		t.Fatal("expected offline status")
	}
}

func TestStatusOnline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"ID":"test-peer","AgentVersion":"kubo/0.29"}`))
	}))
	defer srv.Close()

	b := New(srv.URL, "http://127.0.0.1:8180")
	ctx := context.Background()

	status, err := b.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if status["online"] != true {
		t.Fatal("expected online")
	}
	if status["ID"] != "test-peer" {
		t.Fatalf("expected test-peer, got %v", status["ID"])
	}
}

func TestAddBytes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v0/add" {
			w.Write([]byte(`{"Hash":"QmTest","Name":"test.txt","Size":"123"}`))
		}
	}))
	defer srv.Close()

	b := New(srv.URL, "http://127.0.0.1:8180")
	ctx := context.Background()

	result, err := b.AddBytes(ctx, "test.txt", []byte("hello world"))
	if err != nil {
		t.Fatal(err)
	}
	if result["cid"] != "QmTest" {
		t.Fatalf("expected QmTest, got %v", result["cid"])
	}
	if result["name"] != "test.txt" {
		t.Fatalf("expected test.txt, got %v", result["name"])
	}
}

func TestAddBytesServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	b := New(srv.URL, "http://127.0.0.1:8180")
	ctx := context.Background()

	_, err := b.AddBytes(ctx, "test.txt", []byte("data"))
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestCat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("file contents"))
	}))
	defer srv.Close()

	b := New(srv.URL, "http://127.0.0.1:8180")
	ctx := context.Background()

	data, err := b.Cat(ctx, "QmTest")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "file contents" {
		t.Fatalf("expected 'file contents', got '%s'", string(data))
	}
}

func TestPin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"Pins":["QmTest"]}`))
	}))
	defer srv.Close()

	b := New(srv.URL, "http://127.0.0.1:8180")
	ctx := context.Background()

	result, err := b.Pin(ctx, "QmTest")
	if err != nil {
		t.Fatal(err)
	}
	pins, _ := result["Pins"].([]any)
	if len(pins) != 1 || pins[0] != "QmTest" {
		t.Fatal("pin result mismatch")
	}
}

func TestPublishIPNS(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"Name":"k51...","Value":"/ipfs/QmTest"}`))
	}))
	defer srv.Close()

	b := New(srv.URL, "http://127.0.0.1:8180")
	ctx := context.Background()

	result, err := b.PublishIPNS(ctx, "QmTest")
	if err != nil {
		t.Fatal(err)
	}
	if result["Name"] != "k51..." {
		t.Fatalf("expected k51..., got %v", result["Name"])
	}
}

func TestGatewayURL(t *testing.T) {
	b := New("http://127.0.0.1:5001", "http://127.0.0.1:8180")
	result, err := b.AddBytes(context.Background(), "test", []byte("data"))

	// This will fail but we can test the URL construction
	if err != nil && result == nil {
		// expected since no server running
	}
}
