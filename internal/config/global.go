package config

import (
	"errors"
	"strings"
)

type GlobalFlags struct {
	RPC  string
	JSON bool
}

func ResolveRPCURL(flags GlobalFlags, lookup func(string) string) (string, error) {
	if value := strings.TrimSpace(flags.RPC); value != "" {
		return value, nil
	}

	if value := strings.TrimSpace(lookup("EVMGO_RPC")); value != "" {
		return value, nil
	}

	return "", errors.New("rpc endpoint required: pass --rpc or set EVMGO_RPC")
}
