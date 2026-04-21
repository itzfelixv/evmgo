package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/itzfelixv/evmgo/internal/actions"
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

		switch req.Method {
		case "eth_getTransactionByHash":
			if string(req.Params) != `["`+txHash+`"]` {
				t.Fatalf("unexpected params: %s", req.Params)
			}
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"hash":"` + txHash + `","from":"0x1111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222","value":"0x10","blockNumber":"0x2a","type":"0x0","nonce":"0x1","gas":"0x5208","gasPrice":"0x3b9aca00","input":"0x"}}`))
		case "eth_getTransactionReceipt":
			if string(req.Params) != `["`+txHash+`"]` {
				t.Fatalf("unexpected params: %s", req.Params)
			}
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"transactionHash":"` + txHash + `","blockNumber":"0x2a","status":"0x1","gasUsed":"0x5208","contractAddress":""}}`))
		case "eth_getBlockByNumber":
			if string(req.Params) != `["0x2a",false]` {
				t.Fatalf("unexpected params: %s", req.Params)
			}
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"number":"0x2a","hash":"0xabc","parentHash":"0xdef","timestamp":"0x5","transactions":[]}}`))
		default:
			t.Fatalf("unexpected method: %s", req.Method)
		}
	}))
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{"--rpc", server.URL, "tx", txHash})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	want := "hash: " + txHash + "\nkind: transfer\nfrom: 0x1111111111111111111111111111111111111111\nto: 0x2222222222222222222222222222222222222222\nvalue: 0x10\nblock: 0x2a\ntimestamp: 0x5\nstatus: success\ntype: 0x0\nnonce: 0x1\ngas-limit: 0x5208\ngas-price: 0x3b9aca00\ngas-used: 0x5208\ninput-bytes: 0\n"
	if stdout.String() != want {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestAppendDecodeLinesShowsStatusWithoutError(t *testing.T) {
	lines := appendDecodeLines(nil, actions.TxDecode{Status: "unavailable"})

	want := []string{"  decode: unavailable"}
	if len(lines) != len(want) {
		t.Fatalf("unexpected line count: got %d want %d (%q)", len(lines), len(want), lines)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Fatalf("unexpected line %d: got %q want %q", i, lines[i], want[i])
		}
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

func TestTxCommandTextOutputShowsEnrichedTransfer(t *testing.T) {
	txHash := "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		switch req.Method {
		case "eth_getTransactionByHash":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"hash":"` + txHash + `","from":"0x1111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222","value":"0x10","blockNumber":"0x2a","type":"0x0","nonce":"0x1","gas":"0x5208","gasPrice":"0x3b9aca00","input":"0x"}}`))
		case "eth_getTransactionReceipt":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"transactionHash":"` + txHash + `","blockNumber":"0x2a","status":"0x1","gasUsed":"0x5208","contractAddress":""}}`))
		case "eth_getBlockByNumber":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"number":"0x2a","hash":"0xabc","parentHash":"0xdef","timestamp":"0x5","transactions":[]}}`))
		default:
			t.Fatalf("unexpected method: %s", req.Method)
		}
	}))
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{"--rpc", server.URL, "tx", txHash})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	want := "hash: " + txHash + "\nkind: transfer\nfrom: 0x1111111111111111111111111111111111111111\nto: 0x2222222222222222222222222222222222222222\nvalue: 0x10\nblock: 0x2a\ntimestamp: 0x5\nstatus: success\ntype: 0x0\nnonce: 0x1\ngas-limit: 0x5208\ngas-price: 0x3b9aca00\ngas-used: 0x5208\ninput-bytes: 0\n"
	if stdout.String() != want {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestTxCommandTextOutputShowsDefaultCallSectionAndCalldata(t *testing.T) {
	txHash := "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		switch req.Method {
		case "eth_getTransactionByHash":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"hash":"` + txHash + `","from":"0x1111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222","value":"0x0","blockNumber":"0x2a","type":"0x2","nonce":"0x15","gas":"0x186a0","maxFeePerGas":"0x59682f00","maxPriorityFeePerGas":"0x3b9aca00","input":"0xa9059cbb00000000000000000000000011111111111111111111111111111111111111110000000000000000000000000000000000000000000000000de0b6b3a7640000"}}`))
		case "eth_getTransactionReceipt":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"transactionHash":"` + txHash + `","blockNumber":"0x2a","status":"0x1","gasUsed":"0x5208","contractAddress":""}}`))
		case "eth_getBlockByNumber":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"number":"0x2a","hash":"0xabc","parentHash":"0xdef","timestamp":"0x67d9d2f0","transactions":[]}}`))
		default:
			t.Fatalf("unexpected method: %s", req.Method)
		}
	}))
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{"--rpc", server.URL, "tx", txHash, "--input"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	want := "hash: " + txHash + "\nkind: contract-call\nfrom: 0x1111111111111111111111111111111111111111\nto: 0x2222222222222222222222222222222222222222\nvalue: 0x0\nblock: 0x2a\ntimestamp: 0x67d9d2f0\nstatus: success\ntype: 0x2\nnonce: 0x15\ngas-limit: 0x186a0\nmax-fee-per-gas: 0x59682f00\nmax-priority-fee-per-gas: 0x3b9aca00\ngas-used: 0x5208\ninput-bytes: 68\nselector: 0xa9059cbb\ncalldata: 0xa9059cbb00000000000000000000000011111111111111111111111111111111111111110000000000000000000000000000000000000000000000000de0b6b3a7640000\n\ncall:\n  decode: unavailable (no ABI)\n"
	if stdout.String() != want {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestTxCommandTextOutputShowsDecodedCallSection(t *testing.T) {
	txHash := "0xdddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		switch req.Method {
		case "eth_getTransactionByHash":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"hash":"` + txHash + `","from":"0x1111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222","value":"0x0","blockNumber":"0x2a","type":"0x2","nonce":"0x15","gas":"0x186a0","maxFeePerGas":"0x59682f00","maxPriorityFeePerGas":"0x3b9aca00","input":"0xa9059cbb00000000000000000000000011111111111111111111111111111111111111110000000000000000000000000000000000000000000000000de0b6b3a7640000"}}`))
		case "eth_getTransactionReceipt":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"transactionHash":"` + txHash + `","blockNumber":"0x2a","status":"0x1","gasUsed":"0x5208","contractAddress":""}}`))
		case "eth_getBlockByNumber":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"number":"0x2a","hash":"0xabc","parentHash":"0xdef","timestamp":"0x67d9d2f0","transactions":[]}}`))
		default:
			t.Fatalf("unexpected method: %s", req.Method)
		}
	}))
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{"--rpc", server.URL, "tx", txHash, "--abi", "../../testdata/abi/erc20.json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	want := "hash: " + txHash + "\nkind: contract-call\nfrom: 0x1111111111111111111111111111111111111111\nto: 0x2222222222222222222222222222222222222222\nvalue: 0x0\nblock: 0x2a\ntimestamp: 0x67d9d2f0\nstatus: success\ntype: 0x2\nnonce: 0x15\ngas-limit: 0x186a0\nmax-fee-per-gas: 0x59682f00\nmax-priority-fee-per-gas: 0x3b9aca00\ngas-used: 0x5208\ninput-bytes: 68\nselector: 0xa9059cbb\n\ncall:\n  method: transfer(address,uint256)\n  to(address): 0x1111111111111111111111111111111111111111\n  amount(uint256): 1000000000000000000\n"
	if stdout.String() != want {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestTxCommandJSONOutputIncludesInitcodeWhenRequested(t *testing.T) {
	txHash := "0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		switch req.Method {
		case "eth_getTransactionByHash":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"hash":"` + txHash + `","from":"0x1111111111111111111111111111111111111111","value":"0x0","blockNumber":"0x2a","type":"0x2","nonce":"0x15","gas":"0x1fbd0","maxFeePerGas":"0x59682f00","maxPriorityFeePerGas":"0x3b9aca00","input":"0x608060405234801561001057600080fd5b50"}}`))
		case "eth_getTransactionReceipt":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"transactionHash":"` + txHash + `","blockNumber":"0x2a","status":"0x1","gasUsed":"0x1a2b3","contractAddress":"0x2222222222222222222222222222222222222222"}}`))
		case "eth_getBlockByNumber":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"number":"0x2a","hash":"0xabc","parentHash":"0xdef","timestamp":"0x67d9d2f0","transactions":[]}}`))
		default:
			t.Fatalf("unexpected method: %s", req.Method)
		}
	}))
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{"--rpc", server.URL, "tx", txHash, "--json", "--input"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	want := "{\n  \"hash\": \"" + txHash + "\",\n  \"kind\": \"contract-creation\",\n  \"from\": \"0x1111111111111111111111111111111111111111\",\n  \"createdContract\": \"0x2222222222222222222222222222222222222222\",\n  \"value\": \"0x0\",\n  \"blockNumber\": \"0x2a\",\n  \"timestamp\": \"0x67d9d2f0\",\n  \"status\": \"success\",\n  \"type\": \"0x2\",\n  \"nonce\": \"0x15\",\n  \"gasLimit\": \"0x1fbd0\",\n  \"maxFeePerGas\": \"0x59682f00\",\n  \"maxPriorityFeePerGas\": \"0x3b9aca00\",\n  \"gasUsed\": \"0x1a2b3\",\n  \"inputBytes\": 18,\n  \"input\": {\n    \"label\": \"initcode\",\n    \"data\": \"0x608060405234801561001057600080fd5b50\"\n  }\n}\n"
	if stdout.String() != want {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestTxCommandJSONOutputShowsDecodedCallSection(t *testing.T) {
	txHash := "0xdddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		switch req.Method {
		case "eth_getTransactionByHash":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"hash":"` + txHash + `","from":"0x1111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222","value":"0x0","blockNumber":"0x2a","type":"0x2","nonce":"0x15","gas":"0x186a0","maxFeePerGas":"0x59682f00","maxPriorityFeePerGas":"0x3b9aca00","input":"0xa9059cbb00000000000000000000000011111111111111111111111111111111111111110000000000000000000000000000000000000000000000000de0b6b3a7640000"}}`))
		case "eth_getTransactionReceipt":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"transactionHash":"` + txHash + `","blockNumber":"0x2a","status":"0x1","gasUsed":"0x5208","contractAddress":""}}`))
		case "eth_getBlockByNumber":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"number":"0x2a","hash":"0xabc","parentHash":"0xdef","timestamp":"0x67d9d2f0","transactions":[]}}`))
		default:
			t.Fatalf("unexpected method: %s", req.Method)
		}
	}))
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{"--rpc", server.URL, "tx", txHash, "--json", "--abi", "../../testdata/abi/erc20.json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	want := "{\n  \"hash\": \"" + txHash + "\",\n  \"kind\": \"contract-call\",\n  \"from\": \"0x1111111111111111111111111111111111111111\",\n  \"to\": \"0x2222222222222222222222222222222222222222\",\n  \"value\": \"0x0\",\n  \"blockNumber\": \"0x2a\",\n  \"timestamp\": \"0x67d9d2f0\",\n  \"status\": \"success\",\n  \"type\": \"0x2\",\n  \"nonce\": \"0x15\",\n  \"gasLimit\": \"0x186a0\",\n  \"maxFeePerGas\": \"0x59682f00\",\n  \"maxPriorityFeePerGas\": \"0x3b9aca00\",\n  \"gasUsed\": \"0x5208\",\n  \"inputBytes\": 68,\n  \"call\": {\n    \"selector\": \"0xa9059cbb\",\n    \"decode\": {\n      \"status\": \"decoded\",\n      \"method\": \"transfer(address,uint256)\",\n      \"args\": [\n        {\n          \"name\": \"to\",\n          \"type\": \"address\",\n          \"value\": \"0x1111111111111111111111111111111111111111\"\n        },\n        {\n          \"name\": \"amount\",\n          \"type\": \"uint256\",\n          \"value\": \"1000000000000000000\"\n        }\n      ]\n    }\n  }\n}\n"
	if stdout.String() != want {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestTxCommandJSONOutputShowsUnavailableDecodeState(t *testing.T) {
	txHash := "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		switch req.Method {
		case "eth_getTransactionByHash":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"hash":"` + txHash + `","from":"0x1111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222","value":"0x0","blockNumber":"0x2a","type":"0x2","nonce":"0x15","gas":"0x186a0","maxFeePerGas":"0x59682f00","maxPriorityFeePerGas":"0x3b9aca00","input":"0xffffffff0000000000000000000000001111111111111111111111111111111111111111"}}`))
		case "eth_getTransactionReceipt":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"transactionHash":"` + txHash + `","blockNumber":"0x2a","status":"0x1","gasUsed":"0x5208","contractAddress":""}}`))
		case "eth_getBlockByNumber":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"number":"0x2a","hash":"0xabc","parentHash":"0xdef","timestamp":"0x67d9d2f0","transactions":[]}}`))
		default:
			t.Fatalf("unexpected method: %s", req.Method)
		}
	}))
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{"--rpc", server.URL, "tx", txHash, "--json", "--abi", "../../testdata/abi/erc20.json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	want := "{\n  \"hash\": \"" + txHash + "\",\n  \"kind\": \"contract-call\",\n  \"from\": \"0x1111111111111111111111111111111111111111\",\n  \"to\": \"0x2222222222222222222222222222222222222222\",\n  \"value\": \"0x0\",\n  \"blockNumber\": \"0x2a\",\n  \"timestamp\": \"0x67d9d2f0\",\n  \"status\": \"success\",\n  \"type\": \"0x2\",\n  \"nonce\": \"0x15\",\n  \"gasLimit\": \"0x186a0\",\n  \"maxFeePerGas\": \"0x59682f00\",\n  \"maxPriorityFeePerGas\": \"0x3b9aca00\",\n  \"gasUsed\": \"0x5208\",\n  \"inputBytes\": 36,\n  \"call\": {\n    \"selector\": \"0xffffffff\",\n    \"decode\": {\n      \"status\": \"unavailable\",\n      \"error\": \"selector not found in ABI: 0xffffffff\"\n    }\n  }\n}\n"
	if stdout.String() != want {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
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
