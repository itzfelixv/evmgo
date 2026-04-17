package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRPCCommandJSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("unexpected method: %s", got)
		}
		if got, want := r.Header.Get("Content-Type"), "application/json"; got != want {
			t.Fatalf("unexpected content-type: %s", got)
		}

		var req struct {
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if req.Method != "eth_blockNumber" {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if string(req.Params) != `["latest",9007199254740993]` {
			t.Fatalf("unexpected params: %s", req.Params)
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":9007199254740993}`))
	}))
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(key string) string {
		if key == "EVMGO_RPC" {
			return server.URL
		}
		return ""
	})
	cmd.SetArgs([]string{"rpc", "eth_blockNumber", `"latest"`, `9007199254740993`, "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	want := "{\n  \"result\": 9007199254740993\n}\n"
	if stdout.String() != want {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
}

func TestRPCCommandTextOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if req.Method != "eth_blockNumber" {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if string(req.Params) != `["latest"]` {
			t.Fatalf("unexpected params: %s", req.Params)
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":9007199254740993}`))
	}))
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(key string) string {
		if key == "EVMGO_RPC" {
			return server.URL
		}
		return ""
	})
	cmd.SetArgs([]string{"rpc", "eth_blockNumber", `"latest"`})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	want := "9007199254740993\n"
	if stdout.String() != want {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
}

func TestRPCCommandUsesCommandContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(100 * time.Millisecond):
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x1"}`))
		case <-r.Context().Done():
		}
	}))
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(key string) string {
		if key == "EVMGO_RPC" {
			return server.URL
		}
		return ""
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"rpc", "eth_blockNumber"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected Execute to return a context error")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("expected context deadline exceeded, got %v", err)
	}
}
