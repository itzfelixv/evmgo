package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const diffStateAddress = "0x1111111111111111111111111111111111111111"

type diffStateRPCCall struct {
	Method string
	Params []string
}

func TestDiffStateCommandTextChangedOnly(t *testing.T) {
	server, calls := newDiffStateRPCServer(t, map[string]string{
		"eth_getBalance:0x64":          "0x1",
		"eth_getBalance:0xc8":          "0x2",
		"eth_getTransactionCount:0x64": "0x3",
		"eth_getTransactionCount:0xc8": "0x3",
		"eth_getCode:0x64":             "0x6001",
		"eth_getCode:0xc8":             "0x6001",
		"eth_getStorageAt:0x0:0x64":    "0x00",
		"eth_getStorageAt:0x0:0xc8":    "0x01",
	})
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{
		"--rpc", server.URL,
		"diff", "state", diffStateAddress,
		"--from-block", "100",
		"--to-block", "200",
		"--slot", "0x0",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}

	got := stdout.String()
	for _, visible := range []string{"from: 0x64", "to:   0xc8", "balance:", "storage:", "0x0:"} {
		if !strings.Contains(got, visible) {
			t.Fatalf("stdout missing %q:\n%s", visible, got)
		}
	}
	if strings.Contains(got, "nonce:") {
		t.Fatalf("stdout unexpectedly contains unchanged nonce:\n%s", got)
	}
	if len(*calls) != 8 {
		t.Fatalf("unexpected RPC call count: %d", len(*calls))
	}
}

func TestDiffStateCommandJSONUsesEnvRPCAndIncludesAllFields(t *testing.T) {
	server, _ := newDiffStateRPCServer(t, map[string]string{
		"eth_getBalance:0x64":          "0x1",
		"eth_getBalance:0xc8":          "0x2",
		"eth_getTransactionCount:0x64": "0x3",
		"eth_getTransactionCount:0xc8": "0x3",
		"eth_getCode:0x64":             "0x6001",
		"eth_getCode:0xc8":             "0x6001",
		"eth_getStorageAt:0x0:0x64":    "0x00",
		"eth_getStorageAt:0x0:0xc8":    "0x01",
	})
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(key string) string {
		if key == "EVMGO_RPC" {
			return server.URL
		}
		return ""
	})
	cmd.SetArgs([]string{
		"--json",
		"diff", "state", diffStateAddress,
		"--from-block", "100",
		"--to-block", "200",
		"--slot", "0x0",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}

	var result struct {
		Address string `json:"address"`
		From    string `json:"from_block"`
		To      string `json:"to_block"`
		Changed bool   `json:"changed"`
		Account struct {
			Balance struct {
				Old     string `json:"old"`
				New     string `json:"new"`
				Changed bool   `json:"changed"`
			} `json:"balance"`
			Nonce struct {
				Old     string `json:"old"`
				New     string `json:"new"`
				Changed bool   `json:"changed"`
			} `json:"nonce"`
			Code struct {
				OldHash string `json:"old_hash"`
				NewHash string `json:"new_hash"`
				Changed bool   `json:"changed"`
			} `json:"code"`
		} `json:"account"`
		Storage []struct {
			Slot    string `json:"slot"`
			Old     string `json:"old"`
			New     string `json:"new"`
			Changed bool   `json:"changed"`
		} `json:"storage"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode JSON output: %v\n%s", err, stdout.String())
	}

	if result.Address != diffStateAddress || result.From != "0x64" || result.To != "0xc8" || !result.Changed {
		t.Fatalf("unexpected JSON result header: %#v", result)
	}
	if result.Account.Nonce.Old != "0x3" || result.Account.Nonce.New != "0x3" || result.Account.Nonce.Changed {
		t.Fatalf("expected unchanged nonce in JSON output, got %#v", result.Account.Nonce)
	}
	if len(result.Storage) != 1 || result.Storage[0].Slot != "0x0" || !result.Storage[0].Changed {
		t.Fatalf("expected storage array in JSON output, got %#v", result.Storage)
	}
}

func TestDiffStateCommandRejectsDuplicateSlotBeforeRPC(t *testing.T) {
	server, calls := newDiffStateRPCServer(t, map[string]string{})
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{
		"--rpc", server.URL,
		"diff", "state", diffStateAddress,
		"--from-block", "100",
		"--to-block", "200",
		"--slot", "0x0",
		"--slot", "0x0",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected duplicate slot error")
	}
	if err.Error() != `duplicate slot "0x0"` {
		t.Fatalf("unexpected error: %q", err.Error())
	}
	if len(*calls) != 0 {
		t.Fatalf("expected no RPC calls, got %d", len(*calls))
	}
}

func TestDiffStateCommandNoChangePrintsConciseMessage(t *testing.T) {
	server, _ := newDiffStateRPCServer(t, map[string]string{
		"eth_getBalance:0x64":          "0x1",
		"eth_getBalance:0xc8":          "0x1",
		"eth_getTransactionCount:0x64": "0x3",
		"eth_getTransactionCount:0xc8": "0x3",
		"eth_getCode:0x64":             "0x6001",
		"eth_getCode:0xc8":             "0x6001",
	})
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{
		"--rpc", server.URL,
		"diff", "state", diffStateAddress,
		"--from-block", "100",
		"--to-block", "200",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.String() != "no changes detected between blocks 0x64 and 0xc8\n" {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func newDiffStateRPCServer(t *testing.T, results map[string]string) (*httptest.Server, *[]diffStateRPCCall) {
	t.Helper()

	calls := []diffStateRPCCall{}
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
		calls = append(calls, diffStateRPCCall{Method: req.Method, Params: params})

		key := req.Method + ":" + strings.Join(params[1:], ":")
		result, ok := results[key]
		if !ok {
			t.Fatalf("unexpected request: method=%q params=%#v", req.Method, params)
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"` + result + `"}`))
	}))

	return server, &calls
}
