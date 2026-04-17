package cli

import (
	"fmt"
	"strings"

	"github.com/itzfelixv/evmgo/internal/eth"
)

func validateTransactionHash(raw string) error {
	if raw != strings.TrimSpace(raw) {
		return fmt.Errorf("invalid transaction hash %q", raw)
	}
	if !eth.IsHash(raw) {
		return fmt.Errorf("invalid transaction hash %q", raw)
	}
	return nil
}

func validateStorageSlot(raw string) error {
	if raw != strings.TrimSpace(raw) {
		return fmt.Errorf("invalid slot %q", raw)
	}
	if _, err := eth.DecodeQuantity(raw); err == nil {
		return nil
	}

	decoded, err := eth.DecodeHexBytes(raw)
	if err == nil && len(decoded) <= 32 {
		return nil
	}

	return fmt.Errorf("invalid slot %q", raw)
}

func normalizeStateAddress(raw string) (string, error) {
	if raw != strings.TrimSpace(raw) {
		return "", fmt.Errorf("invalid address %q", raw)
	}

	switch {
	case len(raw) == 40:
		raw = "0x" + raw
	case len(raw) == 42 && strings.HasPrefix(raw, "0X"):
		raw = "0x" + raw[2:]
	}

	address, err := eth.NormalizeAddress(raw)
	if err != nil {
		return "", fmt.Errorf("invalid address %q", raw)
	}

	return address, nil
}
