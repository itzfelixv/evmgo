package actions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itzfelixv/evmgo/internal/rpc"
)

func TestGetBalance(t *testing.T) {
	var gotMethod string
	var gotParams []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		gotMethod = req.Method
		if err := json.Unmarshal(req.Params, &gotParams); err != nil {
			t.Fatalf("failed to decode params: %v", err)
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0xde0b6b3a7640000"}`))
	}))
	defer server.Close()

	result, err := GetBalance(context.Background(), rpc.NewClient(server.URL), "0x1111111111111111111111111111111111111111")
	if err != nil {
		t.Fatalf("GetBalance returned error: %v", err)
	}
	if gotMethod != "eth_getBalance" {
		t.Fatalf("unexpected method: %q", gotMethod)
	}
	if len(gotParams) != 2 || gotParams[0] != "0x1111111111111111111111111111111111111111" || gotParams[1] != "latest" {
		t.Fatalf("unexpected params: %#v", gotParams)
	}
	if result.Balance != "0xde0b6b3a7640000" {
		t.Fatalf("unexpected balance: %q", result.Balance)
	}
}

func TestGetCode(t *testing.T) {
	var gotMethod string
	var gotParams []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		gotMethod = req.Method
		if err := json.Unmarshal(req.Params, &gotParams); err != nil {
			t.Fatalf("failed to decode params: %v", err)
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x60016000"}`))
	}))
	defer server.Close()

	ref, _ := rpc.ParseBlockRefOrLatest("")
	result, err := GetCode(context.Background(), rpc.NewClient(server.URL), "0x1111111111111111111111111111111111111111", ref)
	if err != nil {
		t.Fatalf("GetCode returned error: %v", err)
	}
	if gotMethod != "eth_getCode" {
		t.Fatalf("unexpected method: %q", gotMethod)
	}
	if len(gotParams) != 2 || gotParams[0] != "0x1111111111111111111111111111111111111111" || gotParams[1] != "latest" {
		t.Fatalf("unexpected params: %#v", gotParams)
	}
	if result.Code != "0x60016000" {
		t.Fatalf("unexpected code: %q", result.Code)
	}
}

func TestGetStorage(t *testing.T) {
	var gotMethod string
	var gotParams []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		gotMethod = req.Method
		if err := json.Unmarshal(req.Params, &gotParams); err != nil {
			t.Fatalf("failed to decode params: %v", err)
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x01"}`))
	}))
	defer server.Close()

	ref, _ := rpc.ParseBlockRef("latest")
	result, err := GetStorage(context.Background(), rpc.NewClient(server.URL), "0x1111111111111111111111111111111111111111", "0x0", ref)
	if err != nil {
		t.Fatalf("GetStorage returned error: %v", err)
	}
	if gotMethod != "eth_getStorageAt" {
		t.Fatalf("unexpected method: %q", gotMethod)
	}
	if len(gotParams) != 3 || gotParams[0] != "0x1111111111111111111111111111111111111111" || gotParams[1] != "0x0" || gotParams[2] != "latest" {
		t.Fatalf("unexpected params: %#v", gotParams)
	}
	if result.Value != "0x01" {
		t.Fatalf("unexpected slot value: %q", result.Value)
	}
}
