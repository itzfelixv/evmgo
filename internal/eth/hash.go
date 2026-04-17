package eth

import (
	"encoding/hex"
	"fmt"
	"strings"
)

func NormalizeHash(raw string) (string, error) {
	raw = strings.TrimSpace(raw)

	switch {
	case strings.HasPrefix(raw, "0x"):
		raw = strings.TrimPrefix(raw, "0x")
	case strings.HasPrefix(raw, "0X"):
		raw = strings.TrimPrefix(raw, "0X")
	}

	if len(raw) != 64 {
		return "", fmt.Errorf("invalid hash %q", raw)
	}
	if _, err := hex.DecodeString(raw); err != nil {
		return "", fmt.Errorf("invalid hash %q", raw)
	}

	return "0x" + strings.ToLower(raw), nil
}

func IsHash(raw string) bool {
	_, err := NormalizeHash(raw)
	return err == nil
}
