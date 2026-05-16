package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/itzfelixv/evmgo/internal/eth"
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

type StateDiffQuery struct {
	Address   string
	FromBlock rpc.BlockRef
	ToBlock   rpc.BlockRef
	Slots     []string
}

type StateDiffResult struct {
	Address   string             `json:"address"`
	FromBlock string             `json:"from_block"`
	ToBlock   string             `json:"to_block"`
	Changed   bool               `json:"changed"`
	Account   AccountDiffResult  `json:"account"`
	Storage   []StorageDiffEntry `json:"storage,omitempty"`
}

type AccountDiffResult struct {
	Balance StringDiff `json:"balance"`
	Nonce   StringDiff `json:"nonce"`
	Code    CodeDiff   `json:"code"`
}

type StringDiff struct {
	Old     string `json:"old"`
	New     string `json:"new"`
	Changed bool   `json:"changed"`
}

type CodeDiff struct {
	OldHash string `json:"old_hash"`
	NewHash string `json:"new_hash"`
	Changed bool   `json:"changed"`
}

type StorageDiffEntry struct {
	Slot    string `json:"slot"`
	Old     string `json:"old"`
	New     string `json:"new"`
	Changed bool   `json:"changed"`
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

func DiffState(ctx context.Context, client *rpc.Client, query StateDiffQuery) (StateDiffResult, error) {
	fromBlock := query.FromBlock.RPCArg()
	toBlock := query.ToBlock.RPCArg()

	from, err := readAccountState(ctx, client, query.Address, fromBlock, "from")
	if err != nil {
		return StateDiffResult{}, err
	}
	to, err := readAccountState(ctx, client, query.Address, toBlock, "to")
	if err != nil {
		return StateDiffResult{}, err
	}

	account := AccountDiffResult{
		Balance: stringDiff(from.balance, to.balance),
		Nonce:   stringDiff(from.nonce, to.nonce),
		Code:    codeDiff(from.codeHash, to.codeHash),
	}
	storage, err := readStorageDiffs(ctx, client, query.Address, query.Slots, fromBlock, toBlock)
	if err != nil {
		return StateDiffResult{}, err
	}
	changed := account.Balance.Changed || account.Nonce.Changed || account.Code.Changed
	for _, entry := range storage {
		changed = changed || entry.Changed
	}

	return StateDiffResult{
		Address:   query.Address,
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Changed:   changed,
		Account:   account,
		Storage:   storage,
	}, nil
}

type accountState struct {
	balance  string
	nonce    string
	code     string
	codeHash string
}

func readAccountState(ctx context.Context, client *rpc.Client, address string, block string, side string) (accountState, error) {
	var state accountState
	if err := client.Call(ctx, "eth_getBalance", []any{address, block}, &state.balance); err != nil {
		return accountState{}, fmt.Errorf("%s balance: %w", side, err)
	}
	if err := client.Call(ctx, "eth_getTransactionCount", []any{address, block}, &state.nonce); err != nil {
		return accountState{}, fmt.Errorf("%s nonce: %w", side, err)
	}
	if err := client.Call(ctx, "eth_getCode", []any{address, block}, &state.code); err != nil {
		return accountState{}, fmt.Errorf("%s code: %w", side, err)
	}
	codeHash, err := codeHash(state.code)
	if err != nil {
		return accountState{}, fmt.Errorf("%s code: %w", side, err)
	}
	state.codeHash = codeHash
	return state, nil
}

func readStorageDiffs(ctx context.Context, client *rpc.Client, address string, slots []string, fromBlock string, toBlock string) ([]StorageDiffEntry, error) {
	storage := make([]StorageDiffEntry, 0, len(slots))
	for _, slot := range slots {
		var oldValue string
		if err := client.Call(ctx, "eth_getStorageAt", []any{address, slot, fromBlock}, &oldValue); err != nil {
			return nil, fmt.Errorf("slot %s at from block: %w", slot, err)
		}
		var newValue string
		if err := client.Call(ctx, "eth_getStorageAt", []any{address, slot, toBlock}, &newValue); err != nil {
			return nil, fmt.Errorf("slot %s at to block: %w", slot, err)
		}
		storage = append(storage, StorageDiffEntry{
			Slot:    slot,
			Old:     oldValue,
			New:     newValue,
			Changed: oldValue != newValue,
		})
	}
	return storage, nil
}

func stringDiff(oldValue string, newValue string) StringDiff {
	return StringDiff{
		Old:     oldValue,
		New:     newValue,
		Changed: oldValue != newValue,
	}
}

func codeDiff(oldHash string, newHash string) CodeDiff {
	return CodeDiff{
		OldHash: oldHash,
		NewHash: newHash,
		Changed: oldHash != newHash,
	}
}

func codeHash(code string) (string, error) {
	body := strings.TrimPrefix(code, "0x")
	if !strings.HasPrefix(code, "0x") || len(body)%2 == 1 {
		return "", fmt.Errorf("invalid code bytes %q", code)
	}
	value, err := eth.DecodeHexBytes(code)
	if err != nil {
		return "", err
	}
	return eth.EncodeHexBytes(eth.Keccak256(value)), nil
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
