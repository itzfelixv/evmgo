package actions

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itzfelixv/evmgo/internal/rpc"
)

func TestGetBlock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"number":"0x2a","hash":"0xabc","parentHash":"0xdef","timestamp":"0x5","transactions":["0x1","0x2"]}}`))
	}))
	defer server.Close()

	client := rpc.NewClient(server.URL)
	ref, _ := rpc.ParseBlockRef("42")

	got, err := GetBlock(context.Background(), client, ref)
	if err != nil {
		t.Fatalf("GetBlock returned error: %v", err)
	}
	if got.TransactionCount != 2 {
		t.Fatalf("expected 2 transactions, got %d", got.TransactionCount)
	}
}

func TestGetBlockReturnsNotFoundOnNullResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":null}`))
	}))
	defer server.Close()

	client := rpc.NewClient(server.URL)
	ref, _ := rpc.ParseBlockRef("42")

	_, err := GetBlock(context.Background(), client, ref)
	if err == nil {
		t.Fatal("expected GetBlock to return an error for null result")
	}
	if !errors.Is(err, ErrBlockNotFound) {
		t.Fatalf("expected ErrBlockNotFound, got %v", err)
	}
}

func TestGetTransaction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"hash":"0xaaa","from":"0x1111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222","value":"0x10","blockNumber":"0x2a"}}`))
	}))
	defer server.Close()

	client := rpc.NewClient(server.URL)

	got, err := GetTransaction(context.Background(), client, "0xaaa")
	if err != nil {
		t.Fatalf("GetTransaction returned error: %v", err)
	}
	if got.Value != "0x10" {
		t.Fatalf("expected 0x10, got %q", got.Value)
	}
}

func TestGetTransactionReturnsNotFoundOnNullResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":null}`))
	}))
	defer server.Close()

	client := rpc.NewClient(server.URL)

	_, err := GetTransaction(context.Background(), client, "0xaaa")
	if err == nil {
		t.Fatal("expected GetTransaction to return an error for null result")
	}
	if !errors.Is(err, ErrTransactionNotFound) {
		t.Fatalf("expected ErrTransactionNotFound, got %v", err)
	}
}

func TestGetReceipt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"transactionHash":"0xaaa","blockNumber":"0x2a","status":"0x1","gasUsed":"0x5208","contractAddress":"0x0000000000000000000000000000000000000000"}}`))
	}))
	defer server.Close()

	client := rpc.NewClient(server.URL)

	got, err := GetReceipt(context.Background(), client, "0xaaa")
	if err != nil {
		t.Fatalf("GetReceipt returned error: %v", err)
	}
	if got.Status != "0x1" {
		t.Fatalf("expected 0x1, got %q", got.Status)
	}
}

func TestGetReceiptReturnsNotFoundOnNullResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":null}`))
	}))
	defer server.Close()

	client := rpc.NewClient(server.URL)

	_, err := GetReceipt(context.Background(), client, "0xaaa")
	if err == nil {
		t.Fatal("expected GetReceipt to return an error for null result")
	}
	if !errors.Is(err, ErrReceiptNotFound) {
		t.Fatalf("expected ErrReceiptNotFound, got %v", err)
	}
}
