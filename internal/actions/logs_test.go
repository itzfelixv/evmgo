package actions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/itzfelixv/evmgo/internal/rpc"
)

func TestGetLogs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if req.Method != "eth_getLogs" {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if len(req.Params) != 1 {
			t.Fatalf("unexpected param count: %d", len(req.Params))
		}

		wantFilter := `{"address":"0x2222222222222222222222222222222222222222","fromBlock":"0x1","toBlock":"0x2a","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"]}`
		if string(req.Params[0]) != wantFilter {
			t.Fatalf("unexpected log filter:\nwant %s\ngot  %s", wantFilter, req.Params[0])
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":[{"address":"0x2222222222222222222222222222222222222222","blockNumber":"0x2a","transactionHash":"0xaaa","data":"0x00","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"]}]}`))
	}))
	defer server.Close()

	from, _ := rpc.ParseBlockRef("1")
	to, _ := rpc.ParseBlockRef("42")

	result, err := GetLogs(context.Background(), rpc.NewClient(server.URL), LogQuery{
		Address: "0x2222222222222222222222222222222222222222",
		ABIPath: "../../testdata/abi/erc20.json",
		Event:   "Transfer",
		From:    from,
		To:      to,
	})
	if err != nil {
		t.Fatalf("GetLogs returned error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected one log, got %d", len(result))
	}
	entry := result[0]
	if entry.Address != "0x2222222222222222222222222222222222222222" {
		t.Fatalf("unexpected address: %q", entry.Address)
	}
	if entry.BlockNumber != "0x2a" {
		t.Fatalf("unexpected block number: %q", entry.BlockNumber)
	}
	if entry.TransactionHash != "0xaaa" {
		t.Fatalf("unexpected transaction hash: %q", entry.TransactionHash)
	}
	if entry.Data != "0x00" {
		t.Fatalf("unexpected data: %q", entry.Data)
	}
	if len(entry.Topics) != 1 || entry.Topics[0] != "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
		t.Fatalf("unexpected topics: %#v", entry.Topics)
	}
}

func TestGetLogsRejectsAnonymousEventLookup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("GetLogs should reject anonymous --event lookups before RPC")
	}))
	defer server.Close()

	abiPath := filepath.Join(t.TempDir(), "anonymous.json")
	if err := os.WriteFile(abiPath, []byte(`[
  {
    "type": "event",
    "name": "AnonymousFilled",
    "anonymous": true,
    "inputs": [{ "name": "maker", "type": "address", "indexed": true }]
  }
]`), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	from, _ := rpc.ParseBlockRef("1")
	to, _ := rpc.ParseBlockRef("2")

	_, err := GetLogs(context.Background(), rpc.NewClient(server.URL), LogQuery{
		Address: "0x2222222222222222222222222222222222222222",
		ABIPath: abiPath,
		Event:   "AnonymousFilled",
		From:    from,
		To:      to,
	})
	if err == nil || !strings.Contains(err.Error(), "anonymous event") {
		t.Fatalf("expected anonymous event error, got %v", err)
	}
}

func TestGetLogsSupportsExactOverloadedEventSignature(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if req.Method != "eth_getLogs" {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if len(req.Params) != 1 {
			t.Fatalf("unexpected param count: %d", len(req.Params))
		}

		wantFilter := `{"address":"0x2222222222222222222222222222222222222222","fromBlock":"0x1","toBlock":"0x2","topics":["0x165a4c09b8f9b447ff3e85c37a38fb6ce8b4f5251f471c4abb77bb4c24c20700"]}`
		if string(req.Params[0]) != wantFilter {
			t.Fatalf("unexpected log filter:\nwant %s\ngot  %s", wantFilter, req.Params[0])
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":[{"address":"0x2222222222222222222222222222222222222222","blockNumber":"0x2","transactionHash":"0xbbb","data":"0x00","topics":["0x165a4c09b8f9b447ff3e85c37a38fb6ce8b4f5251f471c4abb77bb4c24c20700"]}]}`))
	}))
	defer server.Close()

	abiPath := filepath.Join(t.TempDir(), "overloaded-events.json")
	if err := os.WriteFile(abiPath, []byte(`[
  {
    "type": "event",
    "name": "Filled",
    "inputs": [{ "name": "amount", "type": "uint256", "indexed": true }]
  },
  {
    "type": "event",
    "name": "Filled",
    "inputs": [{ "name": "maker", "type": "address", "indexed": true }]
  }
]`), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	from, _ := rpc.ParseBlockRef("1")
	to, _ := rpc.ParseBlockRef("2")

	result, err := GetLogs(context.Background(), rpc.NewClient(server.URL), LogQuery{
		Address: "0x2222222222222222222222222222222222222222",
		ABIPath: abiPath,
		Event:   "Filled(address)",
		From:    from,
		To:      to,
	})
	if err != nil {
		t.Fatalf("GetLogs returned error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected one log, got %d", len(result))
	}
}

func TestGetLogsRejectsAmbiguousBareEventName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("GetLogs should reject ambiguous --event lookups before RPC")
	}))
	defer server.Close()

	abiPath := filepath.Join(t.TempDir(), "overloaded-events.json")
	if err := os.WriteFile(abiPath, []byte(`[
  {
    "type": "event",
    "name": "Filled",
    "inputs": [{ "name": "amount", "type": "uint256", "indexed": true }]
  },
  {
    "type": "event",
    "name": "Filled",
    "inputs": [{ "name": "maker", "type": "address", "indexed": true }]
  }
]`), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	from, _ := rpc.ParseBlockRef("1")
	to, _ := rpc.ParseBlockRef("2")

	_, err := GetLogs(context.Background(), rpc.NewClient(server.URL), LogQuery{
		Address: "0x2222222222222222222222222222222222222222",
		ABIPath: abiPath,
		Event:   "Filled",
		From:    from,
		To:      to,
	})
	if err == nil || !strings.Contains(err.Error(), "ambiguous event") {
		t.Fatalf("expected ambiguous event error, got %v", err)
	}
}
