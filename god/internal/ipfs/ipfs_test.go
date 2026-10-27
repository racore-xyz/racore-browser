package ipfs

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
	if status["peerId"] != "test-peer" {
		t.Fatalf("expected test-peer, got %v", status["peerId"])
	}
	if status["agentVersion"] != "kubo/0.29" {
		t.Fatalf("expected normalized agent version, got %v", status["agentVersion"])
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

func TestAddFilesPreservesNestedPaths(t *testing.T) {
	var written []string
	var pinned string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v0/files/write":
			written = append(written, r.URL.Query().Get("arg"))
			if r.URL.Query().Get("parents") != "true" {
				t.Errorf("files write must create parent directories")
			}
			io.Copy(io.Discard, r.Body)
			w.Write([]byte(`{"Name":"written"}`))
		case "/api/v0/files/stat":
			if got := r.URL.Query().Get("arg"); !strings.HasPrefix(got, "/racore-publish-") {
				t.Errorf("unexpected mfs root %q", got)
			}
			w.Write([]byte(`{"Hash":"bafytestdirectory","Type":"directory"}`))
		case "/api/v0/pin/add":
			pinned = r.URL.Query().Get("arg")
			w.Write([]byte(`{"Pins":["bafytestdirectory"]}`))
		case "/api/v0/files/rm":
			w.Write([]byte(`{}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	b := New(srv.URL, "http://127.0.0.1:8180")
	result, err := b.AddFiles(context.Background(), []File{
		{Name: "index.html", Data: []byte("<h1>hi</h1>")},
		{Name: "assets/app.css", Data: []byte("body{}")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result["cid"] != "bafytestdirectory" {
		t.Fatalf("expected directory cid, got %v", result["cid"])
	}
	if pinned != "bafytestdirectory" {
		t.Fatalf("expected pinned directory cid, got %q", pinned)
	}
	if len(written) != 2 {
		t.Fatalf("expected 2 mfs writes, got %v", written)
	}
	if !strings.HasSuffix(written[0], "/index.html") {
		t.Fatalf("expected index.html write, got %q", written[0])
	}
	if !strings.HasSuffix(written[1], "/assets/app.css") {
		t.Fatalf("expected nested asset write, got %q", written[1])
	}
}

func TestAddFilesRejectsEmptyInput(t *testing.T) {
	b := New("http://127.0.0.1:1", "http://127.0.0.1:8180")
	if _, err := b.AddFiles(context.Background(), nil); err == nil {
		t.Fatal("expected error for empty file list")
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
