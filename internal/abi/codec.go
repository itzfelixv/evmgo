package abiutil

import (
	"fmt"

	"github.com/itzfelixv/evmgo/internal/eth"
)

func DecodeMethodOutput(contractABI ABI, method string, rawHex string) ([]string, error) {
	m, err := contractABI.lookupMethod(method)
	if err != nil {
		return nil, err
	}

	raw, err := eth.DecodeHexBytes(rawHex)
	if err != nil {
		return nil, fmt.Errorf("decode output: %w", err)
	}

	return decodeOutputs(m.Outputs, raw)
}
