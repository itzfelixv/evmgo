package rpc

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/itzfelixv/evmgo/internal/eth"
)

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("rpc error %d: %s", e.Code, e.Message)
}

type BlockRef struct {
	tag    string
	number *big.Int
}

func ParseBlockRef(input string) (BlockRef, error) {
	value := strings.TrimSpace(input)
	switch value {
	case "latest", "earliest", "pending", "safe", "finalized":
		return BlockRef{tag: value}, nil
	}

	if strings.HasPrefix(value, "0x") {
		body := strings.TrimPrefix(value, "0x")
		if len(body) > 1 && strings.HasPrefix(body, "0") {
			return BlockRef{}, fmt.Errorf("invalid block selector %q", input)
		}
		n, err := eth.DecodeQuantity(value)
		if err != nil {
			return BlockRef{}, fmt.Errorf("invalid block selector %q: %w", input, err)
		}
		return BlockRef{number: n}, nil
	}

	if strings.HasPrefix(value, "-") {
		return BlockRef{}, fmt.Errorf("invalid block selector %q", input)
	}

	n := new(big.Int)
	if _, ok := n.SetString(value, 10); !ok {
		return BlockRef{}, fmt.Errorf("invalid block selector %q", input)
	}

	return BlockRef{number: n}, nil
}

func ParseBlockRefOrLatest(input string) (BlockRef, error) {
	if strings.TrimSpace(input) == "" {
		return BlockRef{tag: "latest"}, nil
	}
	return ParseBlockRef(input)
}

func (b BlockRef) RPCArg() string {
	if b.number != nil {
		return eth.EncodeQuantity(b.number)
	}
	return b.tag
}
