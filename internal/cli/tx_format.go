package cli

import (
	"fmt"

	"github.com/itzfelixv/evmgo/internal/actions"
)

func renderTxText(result actions.TxViewResult) []string {
	lines := []string{
		"hash: " + result.Hash,
		"kind: " + result.Kind,
		"from: " + result.From,
	}

	if result.To == "" {
		lines = append(lines, "to: <none>")
	} else {
		lines = append(lines, "to: "+result.To)
	}
	if result.CreatedContract != "" {
		lines = append(lines, "created-contract: "+result.CreatedContract)
	}

	lines = append(lines, "value: "+result.Value)
	if result.BlockNumber != "" {
		lines = append(lines, "block: "+result.BlockNumber)
	}
	if result.Timestamp != "" {
		lines = append(lines, "timestamp: "+result.Timestamp)
	}
	if result.Status != "" {
		lines = append(lines, "status: "+result.Status)
	}
	if result.Type != "" {
		lines = append(lines, "type: "+result.Type)
	}
	if result.Nonce != "" {
		lines = append(lines, "nonce: "+result.Nonce)
	}
	if result.GasLimit != "" {
		lines = append(lines, "gas-limit: "+result.GasLimit)
	}
	if result.GasPrice != "" {
		lines = append(lines, "gas-price: "+result.GasPrice)
	}
	if result.MaxFeePerGas != "" {
		lines = append(lines, "max-fee-per-gas: "+result.MaxFeePerGas)
	}
	if result.MaxPriorityFeePerGas != "" {
		lines = append(lines, "max-priority-fee-per-gas: "+result.MaxPriorityFeePerGas)
	}
	if result.GasUsed != "" {
		lines = append(lines, "gas-used: "+result.GasUsed)
	}

	lines = append(lines, fmt.Sprintf("input-bytes: %d", result.InputBytes))
	if result.Call != nil && result.Call.Selector != "" {
		lines = append(lines, "selector: "+result.Call.Selector)
	}
	if result.Input != nil {
		lines = append(lines, result.Input.Label+": "+result.Input.Data)
	}
	if result.Call != nil {
		lines = append(lines, "", "call:")
		lines = appendDecodeLines(lines, result.Call.Decode)
	}
	if result.Revert != nil {
		lines = append(lines, "", "revert:")
		lines = appendDecodeLines(lines, result.Revert.Decode)
	}

	return lines
}

func appendDecodeLines(lines []string, decode actions.TxDecode) []string {
	if decode.Method != "" {
		lines = append(lines, "  method: "+decode.Method)
	}
	for _, arg := range decode.Args {
		lines = append(lines, fmt.Sprintf("  %s(%s): %s", arg.Name, arg.Type, arg.Value))
	}
	if decode.Reason != "" {
		lines = append(lines, "  reason: "+decode.Reason)
	}
	if decode.Status != "decoded" {
		if decode.Error != "" {
			lines = append(lines, fmt.Sprintf("  decode: %s (%s)", decode.Status, decode.Error))
		} else {
			lines = append(lines, fmt.Sprintf("  decode: %s", decode.Status))
		}
	}
	return lines
}
