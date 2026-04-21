package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/itzfelixv/evmgo/internal/rpc"
)

var (
	ErrBlockNotFound       = errors.New("block not found")
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrReceiptNotFound     = errors.New("receipt not found")
)

type BlockResult struct {
	Number           string `json:"number"`
	Hash             string `json:"hash"`
	ParentHash       string `json:"parentHash"`
	Timestamp        string `json:"timestamp"`
	TransactionCount int    `json:"transactionCount"`
}

type TransactionResult struct {
	Hash        string `json:"hash"`
	From        string `json:"from"`
	To          string `json:"to,omitempty"`
	Value       string `json:"value"`
	BlockNumber string `json:"blockNumber,omitempty"`
}

type ReceiptResult struct {
	TransactionHash string `json:"transactionHash"`
	BlockNumber     string `json:"blockNumber"`
	Status          string `json:"status"`
	GasUsed         string `json:"gasUsed"`
	ContractAddress string `json:"contractAddress,omitempty"`
}

func GetBlock(ctx context.Context, client *rpc.Client, ref rpc.BlockRef) (BlockResult, error) {
	var raw struct {
		Number       string   `json:"number"`
		Hash         string   `json:"hash"`
		ParentHash   string   `json:"parentHash"`
		Timestamp    string   `json:"timestamp"`
		Transactions []string `json:"transactions"`
	}
	if err := callJSONResult(ctx, client, "eth_getBlockByNumber", []any{ref.RPCArg(), false}, &raw, ErrBlockNotFound); err != nil {
		return BlockResult{}, err
	}

	return BlockResult{
		Number:           raw.Number,
		Hash:             raw.Hash,
		ParentHash:       raw.ParentHash,
		Timestamp:        raw.Timestamp,
		TransactionCount: len(raw.Transactions),
	}, nil
}

func GetTransaction(ctx context.Context, client *rpc.Client, hash string) (TransactionResult, error) {
	var raw TransactionResult
	if err := callJSONResult(ctx, client, "eth_getTransactionByHash", []any{hash}, &raw, ErrTransactionNotFound); err != nil {
		return TransactionResult{}, err
	}

	return raw, nil
}

func GetReceipt(ctx context.Context, client *rpc.Client, hash string) (ReceiptResult, error) {
	var raw ReceiptResult
	if err := callJSONResult(ctx, client, "eth_getTransactionReceipt", []any{hash}, &raw, ErrReceiptNotFound); err != nil {
		return ReceiptResult{}, err
	}

	return raw, nil
}

func callJSONResult(ctx context.Context, client *rpc.Client, method string, params []any, out any, notFound error) error {
	var raw json.RawMessage
	if err := client.Call(ctx, method, params, &raw); err != nil {
		return err
	}

	if isNullJSON(raw) {
		return notFound
	}

	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode %s result: %w", method, err)
	}

	return nil
}

func isNullJSON(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null"))
}
