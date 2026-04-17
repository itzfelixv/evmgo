package abiutil

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strings"

	"github.com/itzfelixv/evmgo/internal/eth"
)

var twoTo256 = new(big.Int).Lsh(big.NewInt(1), 256)

func PackMethod(contractABI ABI, method string, rawArgs []string) (string, error) {
	m, err := contractABI.lookupMethod(method)
	if err != nil {
		return "", err
	}
	if len(rawArgs) != len(m.Inputs) {
		return "", fmt.Errorf("method %s expects %d args, got %d", method, len(m.Inputs), len(rawArgs))
	}

	values := make([]any, 0, len(rawArgs))
	for i, input := range m.Inputs {
		value, err := parseCLIArg(input.Type, rawArgs[i])
		if err != nil {
			return "", fmt.Errorf("arg %d (%s): %w", i, input.Type.String(), err)
		}
		values = append(values, value)
	}

	selector := "0x" + eth.MethodSelector(signature(m.Name, m.Inputs))

	encoded, err := encodeArguments(m.Inputs, values)
	if err != nil {
		return "", err
	}

	return selector + hex.EncodeToString(encoded), nil
}

func parseCLIArg(typ Type, raw string) (any, error) {
	if typ.Kind == TypeKindTuple || typ.isArray() {
		value, err := parseJSONLiteral(raw)
		if err != nil {
			return nil, err
		}
		return parseJSONValue(typ, value)
	}

	return parseScalarValue(typ, raw)
}

func parseJSONLiteral(raw string) (any, error) {
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()

	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("invalid JSON literal: %w", err)
	}
	if err := decoder.Decode(new(any)); err == nil {
		return nil, fmt.Errorf("invalid JSON literal: trailing content")
	} else if err != io.EOF {
		return nil, fmt.Errorf("invalid JSON literal: trailing content")
	}

	return value, nil
}

func parseJSONValue(typ Type, raw any) (any, error) {
	switch typ.Kind {
	case TypeKindTuple:
		values, ok := raw.([]any)
		if !ok {
			return nil, fmt.Errorf("expected JSON array for tuple")
		}
		if len(values) != len(typ.Components) {
			return nil, fmt.Errorf("tuple expects %d values, got %d", len(typ.Components), len(values))
		}

		out := make([]any, 0, len(values))
		for i, component := range typ.Components {
			value, err := parseJSONValue(component.Type, values[i])
			if err != nil {
				return nil, fmt.Errorf("tuple element %d: %w", i, err)
			}
			out = append(out, value)
		}
		return out, nil
	case TypeKindArray, TypeKindFixedArray:
		values, ok := raw.([]any)
		if !ok {
			return nil, fmt.Errorf("expected JSON array for %s", typ.String())
		}
		if typ.Kind == TypeKindFixedArray && len(values) != typ.Length {
			return nil, fmt.Errorf("expected %d values, got %d", typ.Length, len(values))
		}
		if typ.Elem == nil {
			return nil, fmt.Errorf("invalid ABI type %q", typ.String())
		}

		out := make([]any, 0, len(values))
		for i, value := range values {
			parsed, err := parseJSONValue(*typ.Elem, value)
			if err != nil {
				return nil, fmt.Errorf("array element %d: %w", i, err)
			}
			out = append(out, parsed)
		}
		return out, nil
	case TypeKindString:
		value, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("expected string")
		}
		return value, nil
	case TypeKindBool:
		switch value := raw.(type) {
		case bool:
			return value, nil
		case string:
			return parseScalarValue(typ, value)
		default:
			return nil, fmt.Errorf("expected bool")
		}
	case TypeKindAddress, TypeKindBytes, TypeKindInt, TypeKindUint:
		switch value := raw.(type) {
		case string:
			return parseScalarValue(typ, value)
		case json.Number:
			return parseScalarValue(typ, value.String())
		default:
			return nil, fmt.Errorf("expected %s", typ.String())
		}
	default:
		return nil, fmt.Errorf("unsupported ABI type %q", typ.String())
	}
}

func parseScalarValue(typ Type, raw string) (any, error) {
	switch typ.Kind {
	case TypeKindAddress:
		normalized, err := eth.NormalizeAddress(raw)
		if err != nil {
			return nil, err
		}
		decoded, err := hex.DecodeString(strings.TrimPrefix(normalized, "0x"))
		if err != nil {
			return nil, fmt.Errorf("invalid address %q", raw)
		}
		return decoded, nil
	case TypeKindBool:
		switch strings.ToLower(strings.TrimSpace(raw)) {
		case "true":
			return true, nil
		case "false":
			return false, nil
		default:
			return nil, fmt.Errorf("invalid bool %q", raw)
		}
	case TypeKindString:
		return raw, nil
	case TypeKindBytes:
		decoded, err := eth.DecodeHexBytes(raw)
		if err != nil {
			return nil, err
		}
		if typ.Size > 0 && len(decoded) != typ.Size {
			return nil, fmt.Errorf("expected %d bytes, got %d", typ.Size, len(decoded))
		}
		return decoded, nil
	case TypeKindUint, TypeKindInt:
		value, ok := new(big.Int).SetString(strings.TrimSpace(raw), 10)
		if !ok {
			return nil, fmt.Errorf("invalid integer %q", raw)
		}
		return value, nil
	default:
		return nil, fmt.Errorf("unsupported ABI type %q", typ.String())
	}
}

func encodeArguments(args []Argument, values []any) ([]byte, error) {
	types := make([]Type, 0, len(args))
	for _, arg := range args {
		types = append(types, arg.Type)
	}
	return encodeSequence(types, values)
}

func encodeSequence(types []Type, values []any) ([]byte, error) {
	if len(types) != len(values) {
		return nil, fmt.Errorf("type/value length mismatch: %d != %d", len(types), len(values))
	}

	headSize := 0
	for _, typ := range types {
		if isDynamicType(typ) {
			headSize += 32
			continue
		}
		headSize += staticSize(typ)
	}

	head := make([]byte, 0, headSize)
	tail := make([]byte, 0)
	offset := headSize

	for i, typ := range types {
		if isDynamicType(typ) {
			encoded, err := encodeValue(typ, values[i])
			if err != nil {
				return nil, err
			}
			head = append(head, encodeWord(big.NewInt(int64(offset)))...)
			tail = append(tail, encoded...)
			offset += len(encoded)
			continue
		}

		encoded, err := encodeValue(typ, values[i])
		if err != nil {
			return nil, err
		}
		head = append(head, encoded...)
	}

	return append(head, tail...), nil
}

func encodeValue(typ Type, value any) ([]byte, error) {
	switch typ.Kind {
	case TypeKindAddress:
		address, ok := value.([]byte)
		if !ok || len(address) != 20 {
			return nil, fmt.Errorf("invalid address value")
		}
		return leftPad(address), nil
	case TypeKindBool:
		flag, ok := value.(bool)
		if !ok {
			return nil, fmt.Errorf("invalid bool value")
		}
		if flag {
			return encodeWord(big.NewInt(1)), nil
		}
		return make([]byte, 32), nil
	case TypeKindString:
		text, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("invalid string value")
		}
		return encodeDynamicBytes([]byte(text)), nil
	case TypeKindBytes:
		data, ok := value.([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid bytes value")
		}
		if typ.Size > 0 {
			if len(data) != typ.Size {
				return nil, fmt.Errorf("expected %d bytes, got %d", typ.Size, len(data))
			}
			return rightPad(data), nil
		}
		return encodeDynamicBytes(data), nil
	case TypeKindUint:
		number, ok := value.(*big.Int)
		if !ok {
			return nil, fmt.Errorf("invalid uint value")
		}
		return encodeUintValue(typ.Size, number)
	case TypeKindInt:
		number, ok := value.(*big.Int)
		if !ok {
			return nil, fmt.Errorf("invalid int value")
		}
		return encodeIntValue(typ.Size, number)
	case TypeKindArray:
		values, ok := value.([]any)
		if !ok {
			return nil, fmt.Errorf("invalid array value")
		}
		if typ.Elem == nil {
			return nil, fmt.Errorf("invalid ABI type %q", typ.String())
		}
		encoded, err := encodeRepeated(*typ.Elem, values)
		if err != nil {
			return nil, err
		}
		return append(encodeWord(big.NewInt(int64(len(values)))), encoded...), nil
	case TypeKindFixedArray:
		values, ok := value.([]any)
		if !ok {
			return nil, fmt.Errorf("invalid array value")
		}
		if len(values) != typ.Length {
			return nil, fmt.Errorf("expected %d values, got %d", typ.Length, len(values))
		}
		if typ.Elem == nil {
			return nil, fmt.Errorf("invalid ABI type %q", typ.String())
		}
		return encodeRepeated(*typ.Elem, values)
	case TypeKindTuple:
		values, ok := value.([]any)
		if !ok {
			return nil, fmt.Errorf("invalid tuple value")
		}
		if len(values) != len(typ.Components) {
			return nil, fmt.Errorf("tuple expects %d values, got %d", len(typ.Components), len(values))
		}
		types := make([]Type, 0, len(typ.Components))
		for _, component := range typ.Components {
			types = append(types, component.Type)
		}
		return encodeSequence(types, values)
	default:
		return nil, fmt.Errorf("unsupported ABI type %q", typ.String())
	}
}

func encodeRepeated(elem Type, values []any) ([]byte, error) {
	types := make([]Type, len(values))
	for i := range values {
		types[i] = elem
	}
	return encodeSequence(types, values)
}

func isDynamicType(typ Type) bool {
	switch typ.Kind {
	case TypeKindString:
		return true
	case TypeKindBytes:
		return typ.Size == 0
	case TypeKindArray:
		return true
	case TypeKindFixedArray:
		return typ.Elem != nil && isDynamicType(*typ.Elem)
	case TypeKindTuple:
		for _, component := range typ.Components {
			if isDynamicType(component.Type) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func staticSize(typ Type) int {
	switch typ.Kind {
	case TypeKindFixedArray:
		if typ.Elem == nil {
			return 0
		}
		return typ.Length * staticSize(*typ.Elem)
	case TypeKindTuple:
		total := 0
		for _, component := range typ.Components {
			total += staticSize(component.Type)
		}
		return total
	default:
		return 32
	}
}

func encodeDynamicBytes(data []byte) []byte {
	out := make([]byte, 0, 32+((len(data)+31)/32)*32)
	out = append(out, encodeWord(big.NewInt(int64(len(data))))...)
	out = append(out, data...)
	padding := (32 - (len(data) % 32)) % 32
	if padding > 0 {
		out = append(out, make([]byte, padding)...)
	}
	return out
}

func encodeUintValue(size int, value *big.Int) ([]byte, error) {
	if value == nil {
		return nil, fmt.Errorf("invalid uint value")
	}
	if value.Sign() < 0 {
		return nil, fmt.Errorf("unsigned integer cannot be negative")
	}
	if size == 0 {
		size = 256
	}
	if value.BitLen() > size {
		return nil, fmt.Errorf("integer overflows uint%d", size)
	}
	return encodeWord(value), nil
}

func encodeIntValue(size int, value *big.Int) ([]byte, error) {
	if value == nil {
		return nil, fmt.Errorf("invalid int value")
	}
	if size == 0 {
		size = 256
	}

	limit := new(big.Int).Lsh(big.NewInt(1), uint(size-1))
	min := new(big.Int).Neg(limit)
	max := new(big.Int).Sub(new(big.Int).Set(limit), big.NewInt(1))
	if value.Cmp(min) < 0 || value.Cmp(max) > 0 {
		return nil, fmt.Errorf("integer overflows int%d", size)
	}

	if value.Sign() >= 0 {
		return encodeWord(value), nil
	}

	encoded := new(big.Int).Add(twoTo256, value)
	return encodeWord(encoded), nil
}

func encodeWord(value *big.Int) []byte {
	out := make([]byte, 32)
	if value == nil {
		return out
	}
	bytes := value.Bytes()
	copy(out[32-len(bytes):], bytes)
	return out
}

func leftPad(data []byte) []byte {
	out := make([]byte, 32)
	copy(out[32-len(data):], data)
	return out
}

func rightPad(data []byte) []byte {
	out := make([]byte, 32)
	copy(out, data)
	return out
}
