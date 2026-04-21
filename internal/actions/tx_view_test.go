package actions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/itzfelixv/evmgo/internal/rpc"
)

func newRPCServer(t *testing.T, responses map[string]string, hits map[string]int) *httptest.Server {
	t.Helper()

	var mu sync.Mutex

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		mu.Lock()
		hits[req.Method]++
		mu.Unlock()
		payload, ok := responses[req.Method]
		if !ok {
			t.Fatalf("unexpected method: %s", req.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(payload))
	}))
}

func TestGetTransactionViewContractCallWithoutABIDefaultsToUnavailableDecode(t *testing.T) {
	txHash := "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	hits := map[string]int{}

	server := newRPCServer(t, map[string]string{
		"eth_getTransactionByHash":  `{"jsonrpc":"2.0","id":1,"result":{"hash":"` + txHash + `","from":"0x1111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222","value":"0x0","blockNumber":"0x2a","type":"0x2","nonce":"0x15","gas":"0x186a0","maxFeePerGas":"0x59682f00","maxPriorityFeePerGas":"0x3b9aca00","input":"0xa9059cbb00000000000000000000000011111111111111111111111111111111111111110000000000000000000000000000000000000000000000000de0b6b3a7640000"}}`,
		"eth_getTransactionReceipt": `{"jsonrpc":"2.0","id":1,"result":{"transactionHash":"` + txHash + `","blockNumber":"0x2a","status":"0x1","gasUsed":"0x5208","contractAddress":""}}`,
		"eth_getBlockByNumber":      `{"jsonrpc":"2.0","id":1,"result":{"number":"0x2a","hash":"0xabc","parentHash":"0xdef","timestamp":"0x67d9d2f0","transactions":[]}}`,
	}, hits)
	defer server.Close()

	got, err := GetTransactionView(context.Background(), rpc.NewClient(server.URL), txHash, "", false)
	if err != nil {
		t.Fatalf("GetTransactionView returned error: %v", err)
	}
	if got.Kind != "contract-call" {
		t.Fatalf("unexpected kind: %q", got.Kind)
	}
	if got.Status != "success" {
		t.Fatalf("unexpected status: %q", got.Status)
	}
	if got.Timestamp != "0x67d9d2f0" {
		t.Fatalf("unexpected timestamp: %q", got.Timestamp)
	}
	if got.InputBytes != 68 {
		t.Fatalf("unexpected input bytes: %d", got.InputBytes)
	}
	if got.Call == nil {
		t.Fatal("expected call section")
	}
	if got.Call.Selector != "0xa9059cbb" {
		t.Fatalf("unexpected selector: %q", got.Call.Selector)
	}
	if got.Call.Decode.Status != "unavailable" || got.Call.Decode.Error != "no ABI" {
		t.Fatalf("unexpected decode state: %#v", got.Call.Decode)
	}
	if got.Input != nil {
		t.Fatalf("expected input to be hidden by default, got %#v", got.Input)
	}
}

func TestGetTransactionRemainsCLIStable(t *testing.T) {
	txHash := "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"

	server := newRPCServer(t, map[string]string{
		"eth_getTransactionByHash": `{"jsonrpc":"2.0","id":1,"result":{"hash":"` + txHash + `","from":"0x1111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222","value":"0x10","blockNumber":"0x2a","type":"0x2","nonce":"0x15","gas":"0x5208","maxFeePerGas":"0x59682f00","maxPriorityFeePerGas":"0x3b9aca00","input":"0xabcdef"}}`,
	}, map[string]int{})
	defer server.Close()

	got, err := GetTransaction(context.Background(), rpc.NewClient(server.URL), txHash)
	if err != nil {
		t.Fatalf("GetTransaction returned error: %v", err)
	}

	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(encoded, &payload); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	for _, field := range []string{"type", "nonce", "gas", "gasPrice", "maxFeePerGas", "maxPriorityFeePerGas", "input"} {
		if _, ok := payload[field]; ok {
			t.Fatalf("expected GetTransaction JSON to omit %q, got %s", field, encoded)
		}
	}
}

func TestGetTransactionViewContractCreationInputIsOptional(t *testing.T) {
	txHash := "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	hits := map[string]int{}

	server := newRPCServer(t, map[string]string{
		"eth_getTransactionByHash":  `{"jsonrpc":"2.0","id":1,"result":{"hash":"` + txHash + `","from":"0x1111111111111111111111111111111111111111","value":"0x0","blockNumber":"0x2a","type":"0x2","nonce":"0x15","gas":"0x1fbd0","maxFeePerGas":"0x59682f00","maxPriorityFeePerGas":"0x3b9aca00","input":"0x608060405234801561001057600080fd5b50"}}`,
		"eth_getTransactionReceipt": `{"jsonrpc":"2.0","id":1,"result":{"transactionHash":"` + txHash + `","blockNumber":"0x2a","status":"0x1","gasUsed":"0x1a2b3","contractAddress":"0x2222222222222222222222222222222222222222"}}`,
		"eth_getBlockByNumber":      `{"jsonrpc":"2.0","id":1,"result":{"number":"0x2a","hash":"0xabc","parentHash":"0xdef","timestamp":"0x67d9d2f0","transactions":[]}}`,
	}, hits)
	defer server.Close()

	withoutInput, err := GetTransactionView(context.Background(), rpc.NewClient(server.URL), txHash, "", false)
	if err != nil {
		t.Fatalf("GetTransactionView returned error: %v", err)
	}
	if withoutInput.Kind != "contract-creation" {
		t.Fatalf("unexpected kind: %q", withoutInput.Kind)
	}
	if withoutInput.Input != nil {
		t.Fatalf("expected nil input section, got %#v", withoutInput.Input)
	}

	withInput, err := GetTransactionView(context.Background(), rpc.NewClient(server.URL), txHash, "", true)
	if err != nil {
		t.Fatalf("GetTransactionView returned error: %v", err)
	}
	if withInput.Input == nil || withInput.Input.Label != "initcode" {
		t.Fatalf("unexpected input section: %#v", withInput.Input)
	}
	if withInput.CreatedContract != "0x2222222222222222222222222222222222222222" {
		t.Fatalf("unexpected created contract: %q", withInput.CreatedContract)
	}
}

func TestGetTransactionViewPendingTransactionSkipsBlockLookup(t *testing.T) {
	txHash := "0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	hits := map[string]int{}

	server := newRPCServer(t, map[string]string{
		"eth_getTransactionByHash":  `{"jsonrpc":"2.0","id":1,"result":{"hash":"` + txHash + `","from":"0x1111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222","value":"0x0","type":"0x2","nonce":"0x15","gas":"0x186a0","maxFeePerGas":"0x59682f00","maxPriorityFeePerGas":"0x3b9aca00","input":"0x"}}`,
		"eth_getTransactionReceipt": `{"jsonrpc":"2.0","id":1,"result":null}`,
	}, hits)
	defer server.Close()

	got, err := GetTransactionView(context.Background(), rpc.NewClient(server.URL), txHash, "", false)
	if err != nil {
		t.Fatalf("GetTransactionView returned error: %v", err)
	}
	if got.Status != "pending" {
		t.Fatalf("unexpected status: %q", got.Status)
	}
	if got.Timestamp != "" {
		t.Fatalf("expected empty timestamp, got %q", got.Timestamp)
	}
	if hits["eth_getBlockByNumber"] != 0 {
		t.Fatalf("expected no block lookup, got %d", hits["eth_getBlockByNumber"])
	}
}

func TestGetTransactionViewTransferClassification(t *testing.T) {
	txHash := "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	hits := map[string]int{}

	server := newRPCServer(t, map[string]string{
		"eth_getTransactionByHash":  `{"jsonrpc":"2.0","id":1,"result":{"hash":"` + txHash + `","from":"0x1111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222","value":"0x10","blockNumber":"0x2a","type":"0x0","nonce":"0x1","gas":"0x5208","gasPrice":"0x3b9aca00","input":"0x"}}`,
		"eth_getTransactionReceipt": `{"jsonrpc":"2.0","id":1,"result":{"transactionHash":"` + txHash + `","blockNumber":"0x2a","status":"0x1","gasUsed":"0x5208","contractAddress":""}}`,
		"eth_getBlockByNumber":      `{"jsonrpc":"2.0","id":1,"result":{"number":"0x2a","hash":"0xabc","parentHash":"0xdef","timestamp":"0x6","transactions":[]}}`,
	}, hits)
	defer server.Close()

	got, err := GetTransactionView(context.Background(), rpc.NewClient(server.URL), txHash, "", false)
	if err != nil {
		t.Fatalf("GetTransactionView returned error: %v", err)
	}
	if got.Kind != "transfer" {
		t.Fatalf("unexpected kind: %q", got.Kind)
	}
	if got.Call != nil {
		t.Fatalf("expected no call section, got %#v", got.Call)
	}
}

func TestGetTransactionViewMinedWithoutReceiptOmitsStatus(t *testing.T) {
	txHash := "0xdddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
	hits := map[string]int{}

	server := newRPCServer(t, map[string]string{
		"eth_getTransactionByHash":  `{"jsonrpc":"2.0","id":1,"result":{"hash":"` + txHash + `","from":"0x1111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222","value":"0x10","blockNumber":"0x2a","type":"0x0","nonce":"0x1","gas":"0x5208","gasPrice":"0x3b9aca00","input":"0x"}}`,
		"eth_getTransactionReceipt": `{"jsonrpc":"2.0","id":1,"result":null}`,
		"eth_getBlockByNumber":      `{"jsonrpc":"2.0","id":1,"result":{"number":"0x2a","hash":"0xabc","parentHash":"0xdef","timestamp":"0x5","transactions":[]}}`,
	}, hits)
	defer server.Close()

	got, err := GetTransactionView(context.Background(), rpc.NewClient(server.URL), txHash, "", false)
	if err != nil {
		t.Fatalf("GetTransactionView returned error: %v", err)
	}
	if got.Status != "" {
		t.Fatalf("expected empty status, got %q", got.Status)
	}
	if got.Timestamp != "0x5" {
		t.Fatalf("unexpected timestamp: %q", got.Timestamp)
	}
}

func TestGetTransactionViewRevertedStatus(t *testing.T) {
	txHash := "0x9999999999999999999999999999999999999999999999999999999999999999"
	hits := map[string]int{}

	server := newRPCServer(t, map[string]string{
		"eth_getTransactionByHash":  `{"jsonrpc":"2.0","id":1,"result":{"hash":"` + txHash + `","from":"0x1111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222","value":"0x0","blockNumber":"0x2a","type":"0x2","nonce":"0x15","gas":"0x186a0","maxFeePerGas":"0x59682f00","maxPriorityFeePerGas":"0x3b9aca00","input":"0xa9059cbb00000000000000000000000011111111111111111111111111111111111111110000000000000000000000000000000000000000000000000de0b6b3a7640000"}}`,
		"eth_getTransactionReceipt": `{"jsonrpc":"2.0","id":1,"result":{"transactionHash":"` + txHash + `","blockNumber":"0x2a","status":"0x0","gasUsed":"0x5208","contractAddress":""}}`,
		"eth_getBlockByNumber":      `{"jsonrpc":"2.0","id":1,"result":{"number":"0x2a","hash":"0xabc","parentHash":"0xdef","timestamp":"0x7","transactions":[]}}`,
	}, hits)
	defer server.Close()

	got, err := GetTransactionView(context.Background(), rpc.NewClient(server.URL), txHash, "", false)
	if err != nil {
		t.Fatalf("GetTransactionView returned error: %v", err)
	}
	if got.Status != "reverted" {
		t.Fatalf("unexpected status: %q", got.Status)
	}
}
