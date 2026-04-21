package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseBlockRef(t *testing.T) {
	ref, err := ParseBlockRef("42")
	if err != nil {
		t.Fatalf("ParseBlockRef returned error: %v", err)
	}
	if got := ref.RPCArg(); got != "0x2a" {
		t.Fatalf("expected 0x2a, got %q", got)
	}
}

func TestParseBlockRefRejectsNegativeDecimal(t *testing.T) {
	if _, err := ParseBlockRef("-1"); err == nil {
		t.Fatal("expected ParseBlockRef to reject negative decimal selector")
	}
}

func TestCallSendsJSONRPCEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			JSONRPC string          `json:"jsonrpc"`
			ID      int             `json:"id"`
			Method  string          `json:"method"`
			Params  json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		if req.JSONRPC != "2.0" {
			t.Fatalf("unexpected jsonrpc version: %q", req.JSONRPC)
		}
		if req.ID != 1 {
			t.Fatalf("unexpected id: %d", req.ID)
		}
		if req.Method != "eth_blockNumber" {
			t.Fatalf("unexpected method: %v", req.Method)
		}
		if string(req.Params) != "[]" {
			t.Fatalf("unexpected params: %s", req.Params)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x10"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	var out string
	if err := client.Call(context.Background(), "eth_blockNumber", nil, &out); err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if out != "0x10" {
		t.Fatalf("expected 0x10, got %q", out)
	}
}

func TestCallReturnsHTTPStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("backend unavailable"))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	err := client.Call(context.Background(), "eth_blockNumber", nil, new(string))
	if err == nil {
		t.Fatal("expected Call to return an HTTP status error")
	}
	got := err.Error()
	if !strings.Contains(got, "503 Service Unavailable") {
		t.Fatalf("expected status code in error, got %q", got)
	}
	if strings.Contains(got, "decode") {
		t.Fatalf("expected non-2xx response to avoid decode error, got %q", got)
	}
}

func TestCallPreservesRPCErrorData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"error":{"code":3,"message":"execution reverted","data":"0x08c379a00000000000000000000000000000000000000000000000000000000000000020"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	err := client.Call(context.Background(), "eth_call", []any{map[string]any{"to": "0x2222222222222222222222222222222222222222"}, "latest"}, new(string))
	if err == nil {
		t.Fatal("expected Call to return an RPC error")
	}

	var rpcErr *RPCError
	if !errors.As(err, &rpcErr) {
		t.Fatalf("expected RPCError, got %T", err)
	}

	if got, ok := rpcErr.DataString(); !ok || got != "0x08c379a00000000000000000000000000000000000000000000000000000000000000020" {
		t.Fatalf("unexpected RPC error data: %q, %v", got, ok)
	}
}
