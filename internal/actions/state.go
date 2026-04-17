package actions

import (
	"context"

	"github.com/itzfelixv/evmgo/internal/rpc"
)

type BalanceResult struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
}

type CodeResult struct {
	Address string `json:"address"`
	Block   string `json:"block"`
	Code    string `json:"code"`
}

type StorageResult struct {
	Address string `json:"address"`
	Slot    string `json:"slot"`
	Block   string `json:"block"`
	Value   string `json:"value"`
}

func GetBalance(ctx context.Context, client *rpc.Client, address string) (BalanceResult, error) {
	var balance string
	if err := client.Call(ctx, "eth_getBalance", []any{address, "latest"}, &balance); err != nil {
		return BalanceResult{}, err
	}

	return BalanceResult{
		Address: address,
		Balance: balance,
	}, nil
}

func GetCode(ctx context.Context, client *rpc.Client, address string, ref rpc.BlockRef) (CodeResult, error) {
	var code string
	if err := client.Call(ctx, "eth_getCode", []any{address, ref.RPCArg()}, &code); err != nil {
		return CodeResult{}, err
	}

	return CodeResult{
		Address: address,
		Block:   ref.RPCArg(),
		Code:    code,
	}, nil
}

func GetStorage(ctx context.Context, client *rpc.Client, address string, slot string, ref rpc.BlockRef) (StorageResult, error) {
	var value string
	if err := client.Call(ctx, "eth_getStorageAt", []any{address, slot, ref.RPCArg()}, &value); err != nil {
		return StorageResult{}, err
	}

	return StorageResult{
		Address: address,
		Slot:    slot,
		Block:   ref.RPCArg(),
		Value:   value,
	}, nil
}
