package actions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itzfelixv/evmgo/internal/rpc"
)

func TestDiffStateAccountOnlyChanged(t *testing.T) {
	address := "0x1111111111111111111111111111111111111111"
	var calls []struct {
		Method string
		Params []string
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		var params []string
		if err := json.Unmarshal(req.Params, &params); err != nil {
			t.Fatalf("failed to decode params: %v", err)
		}
		calls = append(calls, struct {
			Method string
			Params []string
		}{Method: req.Method, Params: params})

		results := map[string]string{
			"eth_getBalance:0x64":          "0x1",
			"eth_getBalance:0xc8":          "0x2",
			"eth_getTransactionCount:0x64": "0x3",
			"eth_getTransactionCount:0xc8": "0x4",
			"eth_getCode:0x64":             "0x6001",
			"eth_getCode:0xc8":             "0x6002",
		}
		result, ok := results[req.Method+":"+params[len(params)-1]]
		if !ok {
			t.Fatalf("unexpected request: method=%q params=%#v", req.Method, params)
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"` + result + `"}`))
	}))
	defer server.Close()

	fromBlock, _ := rpc.ParseBlockRef("100")
	toBlock, _ := rpc.ParseBlockRef("200")
	result, err := DiffState(context.Background(), rpc.NewClient(server.URL), StateDiffQuery{
		Address:   address,
		FromBlock: fromBlock,
		ToBlock:   toBlock,
	})
	if err != nil {
		t.Fatalf("DiffState returned error: %v", err)
	}

	if len(calls) != 6 {
		t.Fatalf("unexpected call count: %d", len(calls))
	}
	if result.Address != address {
		t.Fatalf("unexpected address: %q", result.Address)
	}
	if result.FromBlock != "0x64" || result.ToBlock != "0xc8" {
		t.Fatalf("unexpected blocks: from=%q to=%q", result.FromBlock, result.ToBlock)
	}
	if !result.Changed {
		t.Fatal("expected result to be changed")
	}
	if !result.Account.Balance.Changed || result.Account.Balance.Old != "0x1" || result.Account.Balance.New != "0x2" {
		t.Fatalf("unexpected balance diff: %#v", result.Account.Balance)
	}
	if !result.Account.Nonce.Changed || result.Account.Nonce.Old != "0x3" || result.Account.Nonce.New != "0x4" {
		t.Fatalf("unexpected nonce diff: %#v", result.Account.Nonce)
	}
	if !result.Account.Code.Changed {
		t.Fatalf("expected code diff to be changed: %#v", result.Account.Code)
	}
	if result.Account.Code.OldHash == "" || result.Account.Code.NewHash == "" {
		t.Fatalf("expected non-empty code hashes: %#v", result.Account.Code)
	}
	if result.Account.Code.OldHash == result.Account.Code.NewHash {
		t.Fatalf("expected different code hashes: %#v", result.Account.Code)
	}
}

func TestDiffStateAccountOnlyUnchanged(t *testing.T) {
	address := "0x1111111111111111111111111111111111111111"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		var params []string
		if err := json.Unmarshal(req.Params, &params); err != nil {
			t.Fatalf("failed to decode params: %v", err)
		}

		results := map[string]string{
			"eth_getBalance":          "0x1",
			"eth_getTransactionCount": "0x3",
			"eth_getCode":             "0x6001",
		}
		result, ok := results[req.Method]
		if !ok {
			t.Fatalf("unexpected request: method=%q params=%#v", req.Method, params)
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"` + result + `"}`))
	}))
	defer server.Close()

	fromBlock, _ := rpc.ParseBlockRef("100")
	toBlock, _ := rpc.ParseBlockRef("200")
	result, err := DiffState(context.Background(), rpc.NewClient(server.URL), StateDiffQuery{
		Address:   address,
		FromBlock: fromBlock,
		ToBlock:   toBlock,
	})
	if err != nil {
		t.Fatalf("DiffState returned error: %v", err)
	}

	if result.Changed {
		t.Fatal("expected result to be unchanged")
	}
	if result.Account.Balance.Changed || result.Account.Balance.Old != "0x1" || result.Account.Balance.New != "0x1" {
		t.Fatalf("unexpected balance diff: %#v", result.Account.Balance)
	}
	if result.Account.Nonce.Changed || result.Account.Nonce.Old != "0x3" || result.Account.Nonce.New != "0x3" {
		t.Fatalf("unexpected nonce diff: %#v", result.Account.Nonce)
	}
	if result.Account.Code.Changed {
		t.Fatalf("unexpected code diff: %#v", result.Account.Code)
	}
	if result.Account.Code.OldHash == "" || result.Account.Code.NewHash == "" {
		t.Fatalf("expected non-empty code hashes: %#v", result.Account.Code)
	}
	if result.Account.Code.OldHash != result.Account.Code.NewHash {
		t.Fatalf("expected equal code hashes: %#v", result.Account.Code)
	}
}

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
