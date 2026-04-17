package eth

import (
	"encoding/hex"
	"fmt"
	"strings"
)

func NormalizeAddress(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "0x") {
		return "", fmt.Errorf("invalid address %q", raw)
	}

	body := strings.TrimPrefix(raw, "0x")
	if len(body) != 40 {
		return "", fmt.Errorf("invalid address %q", raw)
	}
	if _, err := hex.DecodeString(body); err != nil {
		return "", fmt.Errorf("invalid address %q", raw)
	}

	return "0x" + strings.ToLower(body), nil
}

func IsAddress(raw string) bool {
	_, err := NormalizeAddress(raw)
	return err == nil
}
