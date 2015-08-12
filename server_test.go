package conveyor

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/context"
)

func TestServer_Ping(t *testing.T) {
	b := func(ctx context.Context, w Logger, opts BuildOptions) (string, error) {
		return "", nil
	}
	s := NewServer(New(BuilderFunc(b)))

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/", nil)
	req.Header.Set("X-GitHub-Event", "ping")

	s.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatal("Expected 200 OK")
	}
}

func TestServer_Push(t *testing.T) {
	var called bool
	b := func(ctx context.Context, w Logger, opts BuildOptions) (string, error) {
		called = true
		expected := BuildOptions{
			Repository: "remind101/acme-inc",
			Branch:     "master",
			Sha:        "abcd",
		}
		if got, want := opts, expected; got != want {
			t.Fatalf("BuildOptions => %v; want %v", got, want)
		}
		return "", nil
	}
	s := NewServer(New(BuilderFunc(b)))
	s.Builder = BuilderFunc(b) // Remove Async

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(`{
  "ref": "refs/heads/master",
  "head_commit": {
    "id": "abcd"
  },
  "repository": {
    "full_name": "remind101/acme-inc"
  }
}`))
	req.Header.Set("X-GitHub-Event", "push")

	s.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatal("Expected 200 OK")
	}

	if !called {
		t.Fatal("Expected builder to have been called")
	}
}

func TestServer_Push_Fork(t *testing.T) {
	var called bool
	b := func(ctx context.Context, w Logger, opts BuildOptions) (string, error) {
		called = true
		return "", nil
	}
	s := NewServer(New(BuilderFunc(b)))
	s.Builder = BuilderFunc(b) // Remove Async

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(`{
  "ref": "refs/heads/master",
  "head_commit": {
    "id": "abcd"
  },
  "repository": {
    "full_name": "remind101/acme-inc",
    "fork": true
  }
}`))
	req.Header.Set("X-GitHub-Event", "push")

	s.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatal("Expected 200 OK")
	}

	if called {
		t.Fatal("Expected builder to have not been called")
	}
}

func TestNoCache(t *testing.T) {
	tests := []struct {
		in  string
		out bool
	}{
		// Use cache
		{"testing", false},

		// Don't use cache
		{"[docker nocache]", true},
		{"this is a commit [docker nocache]", true},
	}

	for _, tt := range tests {
		if got, want := noCache(tt.in), tt.out; got != want {
			t.Fatalf("noCache(%q) => %v; want %v", tt.in, got, want)
		}
	}
}
