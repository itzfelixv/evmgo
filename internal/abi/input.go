package abiutil

import (
	"errors"
	"fmt"

	"github.com/itzfelixv/evmgo/internal/eth"
)

var (
	ErrSelectorNotFound    = errors.New("selector not found in ABI")
	ErrSelectorAmbiguous   = errors.New("ambiguous selector in ABI")
	ErrInputDecodeMismatch = errors.New("input does not match ABI")
)

type DecodedInput struct {
	Method string
	Args   []DecodedArgument
}

type DecodedArgument struct {
	Name  string
	Type  string
	Value string
}

func DecodeMethodInput(contractABI ABI, rawHex string) (DecodedInput, error) {
	raw, err := eth.DecodeHexBytes(rawHex)
	if err != nil {
		return DecodedInput{}, err
	}
	if len(raw) < 4 {
		return DecodedInput{}, fmt.Errorf("%w: calldata shorter than selector", ErrInputDecodeMismatch)
	}

	selector := eth.EncodeHexBytes(raw[:4])
	method, err := lookupMethodBySelector(contractABI, selector)
	if err != nil {
		return DecodedInput{}, err
	}

	types := make([]Type, 0, len(method.Inputs))
	for _, arg := range method.Inputs {
		types = append(types, arg.Type)
	}

	values, err := decodeSequence(types, raw[4:], 0)
	if err != nil {
		return DecodedInput{}, fmt.Errorf("%w: %v", ErrInputDecodeMismatch, err)
	}

	args := make([]DecodedArgument, 0, len(method.Inputs))
	for i, arg := range method.Inputs {
		rendered, err := renderDecodedValue(arg.Type, values[i], false)
		if err != nil {
			return DecodedInput{}, fmt.Errorf("%w: %v", ErrInputDecodeMismatch, err)
		}

		name := arg.Name
		if name == "" {
			name = fmt.Sprintf("arg[%d]", i)
		}

		args = append(args, DecodedArgument{
			Name:  name,
			Type:  arg.Type.String(),
			Value: rendered,
		})
	}

	return DecodedInput{
		Method: signature(method.Name, method.Inputs),
		Args:   args,
	}, nil
}

func lookupMethodBySelector(contractABI ABI, selector string) (Method, error) {
	matches := make([]Method, 0, 1)
	for sig, method := range contractABI.methodBySignature {
		if "0x"+eth.MethodSelector(sig) == selector {
			matches = append(matches, method)
		}
	}

	switch len(matches) {
	case 0:
		return Method{}, fmt.Errorf("%w: %s", ErrSelectorNotFound, selector)
	case 1:
		return matches[0], nil
	default:
		return Method{}, fmt.Errorf("%w: %s", ErrSelectorAmbiguous, selector)
	}
}
