package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBalanceCommandAliasJSONOutputUsesEnvRPC(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if req.Method != "eth_getBalance" {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if string(req.Params) != `["0x1111111111111111111111111111111111111111","latest"]` {
			t.Fatalf("unexpected params: %s", req.Params)
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0xde0b6b3a7640000"}`))
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
	cmd.SetArgs([]string{"bal", "0x1111111111111111111111111111111111111111", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	want := "{\n  \"address\": \"0x1111111111111111111111111111111111111111\",\n  \"balance\": \"0xde0b6b3a7640000\"\n}\n"
	if stdout.String() != want {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestBalanceCommandTextOutputUsesFlagRPC(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if req.Method != "eth_getBalance" {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if string(req.Params) != `["0x1111111111111111111111111111111111111111","latest"]` {
			t.Fatalf("unexpected params: %s", req.Params)
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0xde0b6b3a7640000"}`))
	}))
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{"--rpc", server.URL, "balance", "0x1111111111111111111111111111111111111111"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	want := "address: 0x1111111111111111111111111111111111111111\nbalance: 0xde0b6b3a7640000\n"
	if stdout.String() != want {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestBalanceCommandAcceptsBareAddressAndNormalizesIt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if req.Method != "eth_getBalance" {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if string(req.Params) != `["0x1111111111111111111111111111111111111111","latest"]` {
			t.Fatalf("unexpected params: %s", req.Params)
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0xde0b6b3a7640000"}`))
	}))
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{"--rpc", server.URL, "balance", "1111111111111111111111111111111111111111"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	want := "address: 0x1111111111111111111111111111111111111111\nbalance: 0xde0b6b3a7640000\n"
	if stdout.String() != want {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestBalanceCommandAcceptsUppercasePrefixedAddressAndNormalizesIt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if req.Method != "eth_getBalance" {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if string(req.Params) != `["0x1111111111111111111111111111111111111111","latest"]` {
			t.Fatalf("unexpected params: %s", req.Params)
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0xde0b6b3a7640000"}`))
	}))
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{"--rpc", server.URL, "balance", "0X1111111111111111111111111111111111111111"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	want := "address: 0x1111111111111111111111111111111111111111\nbalance: 0xde0b6b3a7640000\n"
	if stdout.String() != want {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestCodeCommandTextOutputParsesBlockFlag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if req.Method != "eth_getCode" {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if string(req.Params) != `["0x1111111111111111111111111111111111111111","0x2a"]` {
			t.Fatalf("unexpected params: %s", req.Params)
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x60016000"}`))
	}))
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{"--rpc", server.URL, "code", "0x1111111111111111111111111111111111111111", "--block", "42"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	want := "address: 0x1111111111111111111111111111111111111111\nblock: 0x2a\ncode: 0x60016000\n"
	if stdout.String() != want {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestStorageCommandJSONOutputParsesBlockFlag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if req.Method != "eth_getStorageAt" {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if string(req.Params) != `["0x1111111111111111111111111111111111111111","0x0","0x2a"]` {
			t.Fatalf("unexpected params: %s", req.Params)
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x01"}`))
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
	cmd.SetArgs([]string{"storage", "0x1111111111111111111111111111111111111111", "0x0", "--block", "42", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	want := "{\n  \"address\": \"0x1111111111111111111111111111111111111111\",\n  \"slot\": \"0x0\",\n  \"block\": \"0x2a\",\n  \"value\": \"0x01\"\n}\n"
	if stdout.String() != want {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestStorageCommandHelpShowsBlockFlag(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{"storage", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	help := stdout.String()
	for _, needle := range []string{"storage <address> <slot>", "--block", "Read a storage slot for a contract"} {
		if !strings.Contains(help, needle) {
			t.Fatalf("help output missing %q\n%s", needle, help)
		}
	}
}

func TestStorageCommandRejectsInvalidSlotBeforeRPC(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x01"}`))
	}))
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{"--rpc", server.URL, "storage", "0x1111111111111111111111111111111111111111", "slot"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected invalid slot error")
	}
	if got := err.Error(); got != `invalid slot "slot"` {
		t.Fatalf("unexpected error: %v", err)
	}
	if hits != 0 {
		t.Fatalf("expected no RPC calls, got %d", hits)
	}
}
