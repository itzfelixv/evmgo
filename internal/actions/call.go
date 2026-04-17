package actions

import (
	"context"

	abiutil "github.com/itzfelixv/evmgo/internal/abi"
	"github.com/itzfelixv/evmgo/internal/eth"
	"github.com/itzfelixv/evmgo/internal/rpc"
)

type CallResult struct {
	To      string   `json:"to"`
	Method  string   `json:"method"`
	Data    string   `json:"data"`
	Block   string   `json:"block"`
	Raw     string   `json:"raw"`
	Decoded []string `json:"decoded"`
}

func CallContract(
	ctx context.Context,
	client *rpc.Client,
	to string,
	abiPath string,
	method string,
	args []string,
	ref rpc.BlockRef,
) (CallResult, error) {
	normalizedTo, err := eth.NormalizeAddress(to)
	if err != nil {
		return CallResult{}, err
	}

	contractABI, err := abiutil.LoadFile(abiPath)
	if err != nil {
		return CallResult{}, err
	}

	data, err := abiutil.PackMethod(contractABI, method, args)
	if err != nil {
		return CallResult{}, err
	}

	payload := map[string]any{
		"to":   normalizedTo,
		"data": data,
	}

	var raw string
	if err := client.Call(ctx, "eth_call", []any{payload, ref.RPCArg()}, &raw); err != nil {
		return CallResult{}, err
	}

	decoded, err := abiutil.DecodeMethodOutput(contractABI, method, raw)
	if err != nil {
		return CallResult{}, err
	}

	return CallResult{
		To:      normalizedTo,
		Method:  method,
		Data:    data,
		Block:   ref.RPCArg(),
		Raw:     raw,
		Decoded: decoded,
	}, nil
}
