package actions

import (
	"context"
	"encoding/binary"
	"errors"
	"strings"

	"github.com/itzfelixv/evmgo/internal/eth"
	"github.com/itzfelixv/evmgo/internal/rpc"
)

func replayRevert(ctx context.Context, client *rpc.Client, tx rawTransactionResult) *TxRevertSection {
	payload := map[string]any{
		"from":  tx.From,
		"value": tx.Value,
		"data":  tx.Input,
	}
	if tx.To != "" {
		payload["to"] = tx.To
	}
	if tx.Gas != "" {
		payload["gas"] = tx.Gas
	}
	if tx.GasPrice != "" {
		payload["gasPrice"] = tx.GasPrice
	}
	if tx.MaxFeePerGas != "" {
		payload["maxFeePerGas"] = tx.MaxFeePerGas
	}
	if tx.MaxPriorityFeePerGas != "" {
		payload["maxPriorityFeePerGas"] = tx.MaxPriorityFeePerGas
	}

	blockRef := tx.BlockNumber
	if blockRef == "" {
		blockRef = "latest"
	}

	var raw string
	err := client.Call(ctx, "eth_call", []any{payload, blockRef}, &raw)
	if err == nil {
		return &TxRevertSection{Decode: TxDecode{Status: "unavailable", Error: "no revert data"}}
	}

	return decodeRevertSection(err)
}

func decodeRevertSection(err error) *TxRevertSection {
	var rpcErr *rpc.RPCError
	if !errors.As(err, &rpcErr) {
		return &TxRevertSection{Decode: TxDecode{Status: "unavailable", Error: "no revert data"}}
	}

	if raw, ok := rpcErr.DataString(); ok {
		if reason, ok := decodeStandardRevertReason(raw); ok {
			return &TxRevertSection{Decode: TxDecode{Status: "decoded", Reason: reason}}
		}
	}

	if reason, ok := extractRevertReasonFromMessage(rpcErr.Message); ok {
		return &TxRevertSection{Decode: TxDecode{Status: "decoded", Reason: reason}}
	}

	return &TxRevertSection{Decode: TxDecode{Status: "unavailable", Error: "no revert data"}}
}

func extractRevertReasonFromMessage(message string) (string, bool) {
	const prefix = "execution reverted: "
	if !strings.HasPrefix(message, prefix) {
		return "", false
	}

	reason := strings.TrimSpace(strings.TrimPrefix(message, prefix))
	return reason, reason != ""
}

func decodeStandardRevertReason(raw string) (string, bool) {
	decoded, err := eth.DecodeHexBytes(raw)
	if err != nil || len(decoded) < 68 || eth.EncodeHexBytes(decoded[:4]) != "0x08c379a0" {
		return "", false
	}

	length := binary.BigEndian.Uint64(decoded[60:68])
	start := 68
	end := start + int(length)
	if end > len(decoded) {
		return "", false
	}

	return string(decoded[start:end]), true
}
