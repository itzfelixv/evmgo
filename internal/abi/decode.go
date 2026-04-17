package abiutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/itzfelixv/evmgo/internal/eth"
)

func decodeOutputs(args []Argument, data []byte) ([]string, error) {
	types := make([]Type, 0, len(args))
	for _, arg := range args {
		types = append(types, arg.Type)
	}

	values, err := decodeSequence(types, data, 0)
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(values))
	for i, value := range values {
		rendered, err := renderDecodedValue(args[i].Type, value, false)
		if err != nil {
			return nil, err
		}
		out = append(out, rendered)
	}

	return out, nil
}

func decodeSequence(types []Type, data []byte, base int) ([]any, error) {
	headSize := 0
	for _, typ := range types {
		if isDynamicType(typ) {
			headSize += 32
			continue
		}
		headSize += staticSize(typ)
	}
	if _, err := sliceABI(data, base, headSize); err != nil {
		return nil, err
	}
	headEnd := base + headSize

	values := make([]any, 0, len(types))
	offset := base

	for _, typ := range types {
		if isDynamicType(typ) {
			target, err := decodeOffset(data, offset, base, headEnd)
			if err != nil {
				return nil, err
			}
			value, err := decodeValue(typ, data, target)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
			offset += 32
			continue
		}

		value, err := decodeValue(typ, data, offset)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
		offset += staticSize(typ)
	}

	return values, nil
}

func decodeValue(typ Type, data []byte, offset int) (any, error) {
	switch typ.Kind {
	case TypeKindAddress:
		word, err := readWord(data, offset)
		if err != nil {
			return nil, err
		}
		if !isZeroed(word[:12]) {
			return nil, fmt.Errorf("invalid address padding")
		}
		return eth.EncodeHexBytes(word[12:]), nil
	case TypeKindBool:
		word, err := readWord(data, offset)
		if err != nil {
			return nil, err
		}
		for _, b := range word[:31] {
			if b != 0 {
				return nil, fmt.Errorf("invalid bool encoding")
			}
		}
		switch word[31] {
		case 0:
			return false, nil
		case 1:
			return true, nil
		default:
			return nil, fmt.Errorf("invalid bool encoding")
		}
	case TypeKindString:
		raw, err := decodeDynamicBytes(data, offset)
		if err != nil {
			return nil, err
		}
		return string(raw), nil
	case TypeKindBytes:
		if typ.Size > 0 {
			word, err := readWord(data, offset)
			if err != nil {
				return nil, err
			}
			if !isZeroed(word[typ.Size:]) {
				return nil, fmt.Errorf("invalid bytes padding")
			}
			return append([]byte(nil), word[:typ.Size]...), nil
		}
		return decodeDynamicBytes(data, offset)
	case TypeKindUint:
		word, err := readWord(data, offset)
		if err != nil {
			return nil, err
		}
		value := new(big.Int).SetBytes(word)
		if typ.Size > 0 && value.BitLen() > typ.Size {
			return nil, fmt.Errorf("invalid uint%d encoding", typ.Size)
		}
		return value, nil
	case TypeKindInt:
		word, err := readWord(data, offset)
		if err != nil {
			return nil, err
		}
		value := decodeSignedWord(word)
		if !fitsSignedWidth(value, typ.Size) {
			return nil, fmt.Errorf("invalid int%d encoding", typ.Size)
		}
		return value, nil
	case TypeKindArray:
		lengthWord, err := readWord(data, offset)
		if err != nil {
			return nil, err
		}
		length, err := wordToInt(lengthWord)
		if err != nil {
			return nil, err
		}
		return decodeRepeated(typ, length, data, offset+32)
	case TypeKindFixedArray:
		return decodeRepeated(typ, typ.Length, data, offset)
	case TypeKindTuple:
		types := make([]Type, 0, len(typ.Components))
		for _, component := range typ.Components {
			types = append(types, component.Type)
		}
		return decodeSequence(types, data, offset)
	default:
		return nil, fmt.Errorf("unsupported ABI type %q", typ.String())
	}
}

func decodeRepeated(typ Type, length int, data []byte, base int) ([]any, error) {
	if typ.Elem == nil {
		return nil, fmt.Errorf("invalid ABI type %q", typ.String())
	}

	types := make([]Type, length)
	for i := 0; i < length; i++ {
		types[i] = *typ.Elem
	}

	return decodeSequence(types, data, base)
}

func decodeDynamicBytes(data []byte, offset int) ([]byte, error) {
	lengthWord, err := readWord(data, offset)
	if err != nil {
		return nil, err
	}
	length, err := wordToInt(lengthWord)
	if err != nil {
		return nil, err
	}
	raw, err := sliceABI(data, offset+32, length)
	if err != nil {
		return nil, err
	}
	paddedLength := ((length + 31) / 32) * 32
	padding, err := sliceABI(data, offset+32+length, paddedLength-length)
	if err != nil {
		return nil, err
	}
	if !isZeroed(padding) {
		return nil, fmt.Errorf("invalid dynamic bytes padding")
	}
	return append([]byte(nil), raw...), nil
}

func decodeOffset(data []byte, offset int, base int, headEnd int) (int, error) {
	word, err := readWord(data, offset)
	if err != nil {
		return 0, err
	}
	relative, err := wordToInt(word)
	if err != nil {
		return 0, err
	}
	if relative%32 != 0 {
		return 0, fmt.Errorf("invalid offset %d", relative)
	}
	target := base + relative
	if target < headEnd || target > len(data) {
		return 0, fmt.Errorf("invalid offset %d", relative)
	}
	return target, nil
}

func readWord(data []byte, offset int) ([]byte, error) {
	return sliceABI(data, offset, 32)
}

func sliceABI(data []byte, offset int, length int) ([]byte, error) {
	if offset < 0 || length < 0 || offset > len(data) || offset+length > len(data) {
		return nil, fmt.Errorf("short ABI data")
	}
	return data[offset : offset+length], nil
}

func wordToInt(word []byte) (int, error) {
	value := new(big.Int).SetBytes(word)
	if !value.IsInt64() {
		return 0, fmt.Errorf("offset out of range")
	}
	parsed := value.Int64()
	if parsed < 0 {
		return 0, fmt.Errorf("offset out of range")
	}
	return int(parsed), nil
}

func decodeSignedWord(word []byte) *big.Int {
	value := new(big.Int).SetBytes(word)
	if len(word) > 0 && word[0]&0x80 != 0 {
		return value.Sub(value, twoTo256)
	}
	return value
}

func fitsSignedWidth(value *big.Int, size int) bool {
	if size == 0 {
		size = 256
	}

	limit := new(big.Int).Lsh(big.NewInt(1), uint(size-1))
	min := new(big.Int).Neg(limit)
	max := new(big.Int).Sub(new(big.Int).Set(limit), big.NewInt(1))
	return value.Cmp(min) >= 0 && value.Cmp(max) <= 0
}

func renderDecodedValue(typ Type, value any, nested bool) (string, error) {
	switch typ.Kind {
	case TypeKindAddress:
		text, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("invalid address value")
		}
		if nested {
			return marshalJSON(text)
		}
		return text, nil
	case TypeKindBool:
		flag, ok := value.(bool)
		if !ok {
			return "", fmt.Errorf("invalid bool value")
		}
		if flag {
			return "true", nil
		}
		return "false", nil
	case TypeKindString:
		text, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("invalid string value")
		}
		if nested {
			return marshalJSON(text)
		}
		return text, nil
	case TypeKindBytes:
		raw, ok := value.([]byte)
		if !ok {
			return "", fmt.Errorf("invalid bytes value")
		}
		text := eth.EncodeHexBytes(raw)
		if nested {
			return marshalJSON(text)
		}
		return text, nil
	case TypeKindUint, TypeKindInt:
		number, ok := value.(*big.Int)
		if !ok {
			return "", fmt.Errorf("invalid integer value")
		}
		return number.String(), nil
	case TypeKindArray, TypeKindFixedArray:
		items, ok := value.([]any)
		if !ok {
			return "", fmt.Errorf("invalid array value")
		}
		if typ.Elem == nil {
			return "", fmt.Errorf("invalid ABI type %q", typ.String())
		}
		parts := make([]string, 0, len(items))
		for _, item := range items {
			rendered, err := renderDecodedValue(*typ.Elem, item, true)
			if err != nil {
				return "", err
			}
			parts = append(parts, rendered)
		}
		return joinJSONArray(parts), nil
	case TypeKindTuple:
		items, ok := value.([]any)
		if !ok {
			return "", fmt.Errorf("invalid tuple value")
		}
		if len(items) != len(typ.Components) {
			return "", fmt.Errorf("tuple expects %d values, got %d", len(typ.Components), len(items))
		}
		parts := make([]string, 0, len(items))
		for i, item := range items {
			rendered, err := renderDecodedValue(typ.Components[i].Type, item, true)
			if err != nil {
				return "", err
			}
			parts = append(parts, rendered)
		}
		return joinJSONArray(parts), nil
	default:
		return "", fmt.Errorf("unsupported ABI type %q", typ.String())
	}
}

func joinJSONArray(parts []string) string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	for i, part := range parts {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(part)
	}
	buf.WriteByte(']')
	return buf.String()
}

func marshalJSON(value string) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func isZeroed(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}
