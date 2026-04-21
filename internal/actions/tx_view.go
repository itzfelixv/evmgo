package actions

import (
	"context"
	"errors"
	"fmt"
	"sync"

	abiutil "github.com/itzfelixv/evmgo/internal/abi"
	"github.com/itzfelixv/evmgo/internal/eth"
	"github.com/itzfelixv/evmgo/internal/rpc"
)

type TxDecode struct {
	Status string         `json:"status"`
	Error  string         `json:"error,omitempty"`
	Method string         `json:"method,omitempty"`
	Args   []TxDecodedArg `json:"args,omitempty"`
	Reason string         `json:"reason,omitempty"`
}

type TxDecodedArg struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type TxCallSection struct {
	Selector string   `json:"selector,omitempty"`
	Decode   TxDecode `json:"decode"`
}

type TxRevertSection struct {
	Decode TxDecode `json:"decode"`
}

type TxInputSection struct {
	Label string `json:"label"`
	Data  string `json:"data"`
}

type TxViewResult struct {
	Hash                 string           `json:"hash"`
	Kind                 string           `json:"kind"`
	From                 string           `json:"from"`
	To                   string           `json:"to,omitempty"`
	CreatedContract      string           `json:"createdContract,omitempty"`
	Value                string           `json:"value"`
	BlockNumber          string           `json:"blockNumber,omitempty"`
	Timestamp            string           `json:"timestamp,omitempty"`
	Status               string           `json:"status,omitempty"`
	Type                 string           `json:"type,omitempty"`
	Nonce                string           `json:"nonce,omitempty"`
	GasLimit             string           `json:"gasLimit,omitempty"`
	GasPrice             string           `json:"gasPrice,omitempty"`
	MaxFeePerGas         string           `json:"maxFeePerGas,omitempty"`
	MaxPriorityFeePerGas string           `json:"maxPriorityFeePerGas,omitempty"`
	GasUsed              string           `json:"gasUsed,omitempty"`
	InputBytes           int              `json:"inputBytes"`
	Input                *TxInputSection  `json:"input,omitempty"`
	Call                 *TxCallSection   `json:"call,omitempty"`
	Revert               *TxRevertSection `json:"revert,omitempty"`
}

type rawTransactionResult struct {
	Hash                 string `json:"hash"`
	From                 string `json:"from"`
	To                   string `json:"to,omitempty"`
	Value                string `json:"value"`
	BlockNumber          string `json:"blockNumber,omitempty"`
	Type                 string `json:"type,omitempty"`
	Nonce                string `json:"nonce,omitempty"`
	Gas                  string `json:"gas,omitempty"`
	GasPrice             string `json:"gasPrice,omitempty"`
	MaxFeePerGas         string `json:"maxFeePerGas,omitempty"`
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas,omitempty"`
	Input                string `json:"input,omitempty"`
}

func GetTransactionView(ctx context.Context, client *rpc.Client, hash string, abiPath string, includeInput bool) (TxViewResult, error) {
	var (
		tx      rawTransactionResult
		txErr   error
		receipt ReceiptResult
		rcptErr error
	)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		tx, txErr = getTransactionViewRaw(ctx, client, hash)
	}()

	go func() {
		defer wg.Done()
		receipt, rcptErr = GetReceipt(ctx, client, hash)
	}()

	wg.Wait()

	if txErr != nil {
		return TxViewResult{}, txErr
	}
	if rcptErr != nil && !errors.Is(rcptErr, ErrReceiptNotFound) {
		return TxViewResult{}, rcptErr
	}

	view := TxViewResult{
		Hash:                 tx.Hash,
		Kind:                 classifyTransaction(tx),
		From:                 tx.From,
		To:                   tx.To,
		Value:                tx.Value,
		BlockNumber:          tx.BlockNumber,
		Type:                 tx.Type,
		Nonce:                tx.Nonce,
		GasLimit:             tx.Gas,
		GasPrice:             tx.GasPrice,
		MaxFeePerGas:         tx.MaxFeePerGas,
		MaxPriorityFeePerGas: tx.MaxPriorityFeePerGas,
		GasUsed:              receipt.GasUsed,
		CreatedContract:      receipt.ContractAddress,
		InputBytes:           countInputBytes(tx.Input),
	}

	if errors.Is(rcptErr, ErrReceiptNotFound) && tx.BlockNumber == "" {
		view.Status = "pending"
	} else if receipt.Status == "0x1" {
		view.Status = "success"
	} else if receipt.Status == "0x0" {
		view.Status = "reverted"
	}

	if tx.BlockNumber != "" {
		ref, err := rpc.ParseBlockRef(tx.BlockNumber)
		if err == nil {
			block, err := GetBlock(ctx, client, ref)
			if err == nil {
				view.Timestamp = block.Timestamp
			}
		}
	}

	if view.Kind == "contract-call" && tx.Input != "" && tx.Input != "0x" {
		view.Call = &TxCallSection{
			Selector: selectorFromInput(tx.Input),
			Decode: TxDecode{
				Status: "unavailable",
				Error:  "no ABI",
			},
		}

		if abiPath != "" {
			view.Call.Decode = decodeTxCall(abiPath, tx.Input)
		}
	}

	if includeInput && tx.Input != "" && tx.Input != "0x" {
		view.Input = &TxInputSection{
			Label: inputLabel(view.Kind),
			Data:  tx.Input,
		}
	}

	return view, nil
}

func classifyTransaction(tx rawTransactionResult) string {
	if tx.To == "" {
		return "contract-creation"
	}
	if tx.Input == "" || tx.Input == "0x" {
		return "transfer"
	}
	return "contract-call"
}

func getTransactionViewRaw(ctx context.Context, client *rpc.Client, hash string) (rawTransactionResult, error) {
	var raw rawTransactionResult
	if err := callJSONResult(ctx, client, "eth_getTransactionByHash", []any{hash}, &raw, ErrTransactionNotFound); err != nil {
		return rawTransactionResult{}, err
	}

	return raw, nil
}

func inputLabel(kind string) string {
	if kind == "contract-creation" {
		return "initcode"
	}
	return "calldata"
}

func selectorFromInput(raw string) string {
	decoded, err := eth.DecodeHexBytes(raw)
	if err != nil || len(decoded) < 4 {
		return ""
	}
	return eth.EncodeHexBytes(decoded[:4])
}

func countInputBytes(raw string) int {
	decoded, err := eth.DecodeHexBytes(raw)
	if err != nil {
		return 0
	}
	return len(decoded)
}

func decodeTxCall(abiPath string, rawInput string) TxDecode {
	contractABI, err := abiutil.LoadFile(abiPath)
	if err != nil {
		return TxDecode{
			Status: "unavailable",
			Error:  err.Error(),
		}
	}

	decoded, err := abiutil.DecodeMethodInput(contractABI, rawInput)
	if err != nil {
		switch {
		case errors.Is(err, abiutil.ErrSelectorNotFound):
			return TxDecode{
				Status: "unavailable",
				Error:  err.Error(),
			}
		case errors.Is(err, abiutil.ErrInputDecodeMismatch):
			return TxDecode{
				Status: "failed",
				Error:  fmt.Sprintf("ABI mismatch: %v", err),
			}
		default:
			return TxDecode{
				Status: "failed",
				Error:  err.Error(),
			}
		}
	}

	args := make([]TxDecodedArg, 0, len(decoded.Args))
	for _, arg := range decoded.Args {
		args = append(args, TxDecodedArg{
			Name:  arg.Name,
			Type:  arg.Type,
			Value: arg.Value,
		})
	}

	return TxDecode{
		Status: "decoded",
		Method: decoded.Method,
		Args:   args,
	}
}
