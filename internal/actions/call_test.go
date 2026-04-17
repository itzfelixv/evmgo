package actions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/itzfelixv/evmgo/internal/rpc"
)

func TestCallContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if req.Method != "eth_call" {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if len(req.Params) != 2 {
			t.Fatalf("unexpected param count: %d", len(req.Params))
		}

		wantPayload := `{"data":"0x70a082310000000000000000000000001111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222"}`
		if string(req.Params[0]) != wantPayload {
			t.Fatalf("unexpected call payload:\nwant %s\ngot  %s", wantPayload, req.Params[0])
		}
		if string(req.Params[1]) != `"latest"` {
			t.Fatalf("unexpected block param: %s", req.Params[1])
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x000000000000000000000000000000000000000000000000000000000000000a"}`))
	}))
	defer server.Close()

	ref, _ := rpc.ParseBlockRefOrLatest("")

	result, err := CallContract(
		context.Background(),
		rpc.NewClient(server.URL),
		"0x2222222222222222222222222222222222222222",
		"../../testdata/abi/erc20.json",
		"balanceOf",
		[]string{"0x1111111111111111111111111111111111111111"},
		ref,
	)
	if err != nil {
		t.Fatalf("CallContract returned error: %v", err)
	}
	if result.To != "0x2222222222222222222222222222222222222222" {
		t.Fatalf("unexpected to output: %q", result.To)
	}
	if result.Method != "balanceOf" {
		t.Fatalf("unexpected method output: %q", result.Method)
	}
	if result.Data != "0x70a082310000000000000000000000001111111111111111111111111111111111111111" {
		t.Fatalf("unexpected calldata: %q", result.Data)
	}
	if result.Block != "latest" {
		t.Fatalf("unexpected block output: %q", result.Block)
	}
	if result.Raw != "0x000000000000000000000000000000000000000000000000000000000000000a" {
		t.Fatalf("unexpected raw output: %q", result.Raw)
	}
	if len(result.Decoded) != 1 || result.Decoded[0] != "10" {
		t.Fatalf("unexpected decoded output: %#v", result.Decoded)
	}
}

func TestCallContractSupportsFullMethodSignature(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if req.Method != "eth_call" {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if len(req.Params) != 2 {
			t.Fatalf("unexpected param count: %d", len(req.Params))
		}

		wantPayload := `{"data":"0x03438dd00000000000000000000000001111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222"}`
		if string(req.Params[0]) != wantPayload {
			t.Fatalf("unexpected call payload:\nwant %s\ngot  %s", wantPayload, req.Params[0])
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x0000000000000000000000000000000000000000000000000000000000000001"}`))
	}))
	defer server.Close()

	abiPath := filepath.Join(t.TempDir(), "overloaded.json")
	if err := os.WriteFile(abiPath, []byte(`[
  {
    "type": "function",
    "name": "swap",
    "inputs": [{ "name": "amount", "type": "uint256" }],
    "outputs": [{ "name": "ok", "type": "bool" }]
  },
  {
    "type": "function",
    "name": "swap",
    "inputs": [{ "name": "token", "type": "address" }],
    "outputs": [{ "name": "ok", "type": "bool" }]
  }
]`), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	ref, _ := rpc.ParseBlockRefOrLatest("")

	result, err := CallContract(
		context.Background(),
		rpc.NewClient(server.URL),
		"0x2222222222222222222222222222222222222222",
		abiPath,
		"swap(address)",
		[]string{"0x1111111111111111111111111111111111111111"},
		ref,
	)
	if err != nil {
		t.Fatalf("CallContract returned error: %v", err)
	}
	if result.Method != "swap(address)" {
		t.Fatalf("unexpected method output: %q", result.Method)
	}
	if result.Data != "0x03438dd00000000000000000000000001111111111111111111111111111111111111111" {
		t.Fatalf("unexpected calldata: %q", result.Data)
	}
	if len(result.Decoded) != 1 || result.Decoded[0] != "true" {
		t.Fatalf("unexpected decoded output: %#v", result.Decoded)
	}
}
