package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBlockCommandAliasJSONOutputUsesEnvRPC(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if req.Method != "eth_getBlockByNumber" {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if string(req.Params) != `["0x2a",false]` {
			t.Fatalf("unexpected params: %s", req.Params)
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"number":"0x2a","hash":"0xabc","parentHash":"0xdef","timestamp":"0x5","transactions":["0x1","0x2"]}}`))
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
	cmd.SetArgs([]string{"blk", "42", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	want := "{\n  \"number\": \"0x2a\",\n  \"hash\": \"0xabc\",\n  \"parentHash\": \"0xdef\",\n  \"timestamp\": \"0x5\",\n  \"transactionCount\": 2\n}\n"
	if stdout.String() != want {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestBlockCommandHelpShowsAcceptedSelectors(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{"block", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	help := stdout.String()
	for _, needle := range []string{"block <number|0xnumber|latest|earliest|pending|safe|finalized>", "Fetch a block by number or tag"} {
		if !strings.Contains(help, needle) {
			t.Fatalf("help output missing %q\n%s", needle, help)
		}
	}
}

func TestTxCommandTextOutputUsesFlagRPC(t *testing.T) {
	txHash := "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if req.Method != "eth_getTransactionByHash" {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if string(req.Params) != `["`+txHash+`"]` {
			t.Fatalf("unexpected params: %s", req.Params)
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"hash":"` + txHash + `","from":"0x1111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222","value":"0x10","blockNumber":"0x2a"}}`))
	}))
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{"--rpc", server.URL, "tx", txHash})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	want := "hash: " + txHash + "\nfrom: 0x1111111111111111111111111111111111111111\nto: 0x2222222222222222222222222222222222222222\nvalue: 0x10\nblock: 0x2a\n"
	if stdout.String() != want {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestReceiptCommandAliasJSONOutputUsesEnvRPC(t *testing.T) {
	txHash := "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if req.Method != "eth_getTransactionReceipt" {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if string(req.Params) != `["`+txHash+`"]` {
			t.Fatalf("unexpected params: %s", req.Params)
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"transactionHash":"` + txHash + `","blockNumber":"0x2a","status":"0x1","gasUsed":"0x5208","contractAddress":"0x0000000000000000000000000000000000000000"}}`))
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
	cmd.SetArgs([]string{"rcpt", txHash, "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	want := "{\n  \"transactionHash\": \"" + txHash + "\",\n  \"blockNumber\": \"0x2a\",\n  \"status\": \"0x1\",\n  \"gasUsed\": \"0x5208\",\n  \"contractAddress\": \"0x0000000000000000000000000000000000000000\"\n}\n"
	if stdout.String() != want {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestTxCommandRejectsInvalidHashBeforeRPC(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"hash":"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","from":"0x1111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222","value":"0x10","blockNumber":"0x2a"}}`))
	}))
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{"--rpc", server.URL, "tx", "0xaaa"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected invalid hash error")
	}
	if got := err.Error(); got != `invalid transaction hash "0xaaa"` {
		t.Fatalf("unexpected error: %v", err)
	}
	if hits != 0 {
		t.Fatalf("expected no RPC calls, got %d", hits)
	}
}

func TestReceiptCommandRejectsInvalidHashBeforeRPC(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"transactionHash":"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","blockNumber":"0x2a","status":"0x1","gasUsed":"0x5208","contractAddress":"0x0000000000000000000000000000000000000000"}}`))
	}))
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{"--rpc", server.URL, "receipt", "0xaaa"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected invalid hash error")
	}
	if got := err.Error(); got != `invalid transaction hash "0xaaa"` {
		t.Fatalf("unexpected error: %v", err)
	}
	if hits != 0 {
		t.Fatalf("expected no RPC calls, got %d", hits)
	}
}
