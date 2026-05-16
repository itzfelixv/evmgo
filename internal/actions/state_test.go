package actions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
		if len(params) != 2 || params[0] != address {
			t.Fatalf("unexpected params: %#v", params)
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
	wantCalls := []struct {
		Method string
		Params []string
	}{
		{Method: "eth_getBalance", Params: []string{address, "0x64"}},
		{Method: "eth_getTransactionCount", Params: []string{address, "0x64"}},
		{Method: "eth_getCode", Params: []string{address, "0x64"}},
		{Method: "eth_getBalance", Params: []string{address, "0xc8"}},
		{Method: "eth_getTransactionCount", Params: []string{address, "0xc8"}},
		{Method: "eth_getCode", Params: []string{address, "0xc8"}},
	}
	for i, want := range wantCalls {
		if calls[i].Method != want.Method || len(calls[i].Params) != 2 || calls[i].Params[0] != want.Params[0] || calls[i].Params[1] != want.Params[1] {
			t.Fatalf("unexpected call %d: got=%#v want=%#v", i, calls[i], want)
		}
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
		if len(params) != 2 || params[0] != address {
			t.Fatalf("unexpected params: %#v", params)
		}
		calls = append(calls, struct {
			Method string
			Params []string
		}{Method: req.Method, Params: params})

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

	if len(calls) != 6 {
		t.Fatalf("unexpected call count: %d", len(calls))
	}
	wantCalls := []struct {
		Method string
		Params []string
	}{
		{Method: "eth_getBalance", Params: []string{address, "0x64"}},
		{Method: "eth_getTransactionCount", Params: []string{address, "0x64"}},
		{Method: "eth_getCode", Params: []string{address, "0x64"}},
		{Method: "eth_getBalance", Params: []string{address, "0xc8"}},
		{Method: "eth_getTransactionCount", Params: []string{address, "0xc8"}},
		{Method: "eth_getCode", Params: []string{address, "0xc8"}},
	}
	for i, want := range wantCalls {
		if calls[i].Method != want.Method || len(calls[i].Params) != 2 || calls[i].Params[0] != want.Params[0] || calls[i].Params[1] != want.Params[1] {
			t.Fatalf("unexpected call %d: got=%#v want=%#v", i, calls[i], want)
		}
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

func TestDiffStateIncludesStorageSlots(t *testing.T) {
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
		if len(params) < 2 || params[0] != address {
			t.Fatalf("unexpected params: %#v", params)
		}
		calls = append(calls, struct {
			Method string
			Params []string
		}{Method: req.Method, Params: params})

		results := map[string]string{
			"eth_getBalance:0x64":          "0x1",
			"eth_getBalance:0xc8":          "0x1",
			"eth_getTransactionCount:0x64": "0x3",
			"eth_getTransactionCount:0xc8": "0x3",
			"eth_getCode:0x64":             "0x6001",
			"eth_getCode:0xc8":             "0x6001",
			"eth_getStorageAt:0x0:0x64":    "0x00",
			"eth_getStorageAt:0x0:0xc8":    "0x01",
			"eth_getStorageAt:0x1:0x64":    "0x02",
			"eth_getStorageAt:0x1:0xc8":    "0x02",
		}
		key := req.Method + ":" + strings.Join(params[1:], ":")
		result, ok := results[key]
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
		Slots:     []string{"0x0", "0x1"},
	})
	if err != nil {
		t.Fatalf("DiffState returned error: %v", err)
	}

	if len(calls) != 10 {
		t.Fatalf("unexpected call count: %d", len(calls))
	}
	if !result.Changed {
		t.Fatal("expected result to be changed")
	}
	if result.Account.Balance.Changed || result.Account.Nonce.Changed || result.Account.Code.Changed {
		t.Fatalf("expected account diff to be unchanged: %#v", result.Account)
	}
	if len(result.Storage) != 2 {
		t.Fatalf("unexpected storage entry count: %d", len(result.Storage))
	}
	wantStorage := []StorageDiffEntry{
		{Slot: "0x0", Old: "0x00", New: "0x01", Changed: true},
		{Slot: "0x1", Old: "0x02", New: "0x02", Changed: false},
	}
	for i, want := range wantStorage {
		if result.Storage[i] != want {
			t.Fatalf("unexpected storage entry %d: got=%#v want=%#v", i, result.Storage[i], want)
		}
	}
}

func TestDiffStateStorageErrorNamesSlotAndSide(t *testing.T) {
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
		if req.Method == "eth_getStorageAt" {
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"storage unavailable"}}`))
			return
		}

		result := "0x1"
		if req.Method == "eth_getCode" {
			result = "0x"
		}
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"` + result + `"}`))
	}))
	defer server.Close()

	fromBlock, _ := rpc.ParseBlockRef("100")
	toBlock, _ := rpc.ParseBlockRef("200")
	_, err := DiffState(context.Background(), rpc.NewClient(server.URL), StateDiffQuery{
		Address:   address,
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Slots:     []string{"0x0"},
	})
	if err == nil {
		t.Fatal("expected storage error")
	}
	if !strings.Contains(err.Error(), "slot 0x0 at from block") {
		t.Fatalf("expected slot and side in error, got: %v", err)
	}
}

func TestDiffStateMalformedCodeReturnsError(t *testing.T) {
	tests := []struct {
		name       string
		block      string
		code       string
		wantErrSub string
	}{
		{
			name:       "from non hex",
			block:      "0x64",
			code:       "0xzz",
			wantErrSub: "from code",
		},
		{
			name:       "to non hex",
			block:      "0xc8",
			code:       "0xzz",
			wantErrSub: "to code",
		},
		{
			name:       "from odd length",
			block:      "0x64",
			code:       "0x1",
			wantErrSub: "from code",
		},
		{
			name:       "to odd length",
			block:      "0xc8",
			code:       "0x1",
			wantErrSub: "to code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				if len(params) != 2 || params[0] != address {
					t.Fatalf("unexpected params: %#v", params)
				}

				result := "0x1"
				if req.Method == "eth_getCode" {
					result = "0x"
					if params[1] == tt.block {
						result = tt.code
					}
				}

				_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"` + result + `"}`))
			}))
			defer server.Close()

			fromBlock, _ := rpc.ParseBlockRef("100")
			toBlock, _ := rpc.ParseBlockRef("200")
			_, err := DiffState(context.Background(), rpc.NewClient(server.URL), StateDiffQuery{
				Address:   address,
				FromBlock: fromBlock,
				ToBlock:   toBlock,
			})
			if err == nil {
				t.Fatal("expected malformed code to return error")
			}
			if !strings.Contains(err.Error(), tt.wantErrSub) {
				t.Fatalf("expected %q context in error, got: %v", tt.wantErrSub, err)
			}
		})
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
