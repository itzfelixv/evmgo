package eth

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

func DecodeQuantity(raw string) (*big.Int, error) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "0x") {
		return nil, fmt.Errorf("invalid quantity %q", raw)
	}

	body := strings.TrimPrefix(raw, "0x")
	if body == "" || strings.HasPrefix(body, "-") || strings.HasPrefix(body, "+") {
		return nil, fmt.Errorf("invalid quantity %q", raw)
	}

	value, ok := new(big.Int).SetString(body, 16)
	if !ok {
		return nil, fmt.Errorf("invalid quantity %q", raw)
	}

	return value, nil
}

func EncodeQuantity(value *big.Int) string {
	if value == nil {
		return "0x0"
	}
	if value.Sign() < 0 {
		panic(fmt.Sprintf("invalid quantity %q", value.String()))
	}
	if value.Sign() == 0 {
		return "0x0"
	}

	return "0x" + value.Text(16)
}

func DecodeHexBytes(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "0x") {
		return nil, fmt.Errorf("invalid hex %q", raw)
	}

	body := strings.TrimPrefix(raw, "0x")
	if len(body)%2 == 1 {
		body = "0" + body
	}

	value, err := hex.DecodeString(body)
	if err != nil {
		return nil, fmt.Errorf("invalid hex %q", raw)
	}

	return value, nil
}

func EncodeHexBytes(value []byte) string {
	return "0x" + hex.EncodeToString(value)
}
