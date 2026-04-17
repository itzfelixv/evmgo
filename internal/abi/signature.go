package abiutil

import (
	"fmt"

	"github.com/itzfelixv/evmgo/internal/eth"
)

func MethodSignature(contractABI ABI, method string) (string, error) {
	m, err := contractABI.lookupMethod(method)
	if err != nil {
		return "", err
	}
	return signature(m.Name, m.Inputs), nil
}

func EventSignature(contractABI ABI, event string) (string, error) {
	e, err := contractABI.lookupEvent(event)
	if err != nil {
		return "", err
	}
	return signature(e.Name, e.Inputs), nil
}

func MethodSelector(contractABI ABI, method string) (string, error) {
	signature, err := MethodSignature(contractABI, method)
	if err != nil {
		return "", err
	}
	return "0x" + eth.MethodSelector(signature), nil
}

func EventTopic(contractABI ABI, event string) (string, error) {
	e, err := contractABI.lookupEvent(event)
	if err != nil {
		return "", err
	}
	if e.Anonymous {
		return "", fmt.Errorf("anonymous event %q cannot be used with --event; use --topic filters instead", signature(e.Name, e.Inputs))
	}
	signature := signature(e.Name, e.Inputs)
	return eth.EventTopic(signature), nil
}

func signature(name string, args []Argument) string {
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		parts = append(parts, arg.Type.String())
	}
	return name + "(" + join(parts) + ")"
}

func join(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += "," + parts[i]
	}
	return out
}
