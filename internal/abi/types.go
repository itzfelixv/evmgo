package abiutil

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type ABI struct {
	// Methods contains entries addressable by bare method name.
	// Overloaded methods are resolved through lookupMethod.
	Methods map[string]Method
	// Events contains entries addressable by bare event name.
	// Overloaded events are resolved through lookupEvent.
	Events  map[string]Event

	methodOverloads   map[string][]Method
	eventOverloads    map[string][]Event
	methodBySignature map[string]Method
	eventBySignature  map[string]Event
}

type Method struct {
	Name    string
	Inputs  []Argument
	Outputs []Argument
}

type Event struct {
	Name      string
	Inputs    []Argument
	Anonymous bool
}

type Argument struct {
	Name    string
	Type    Type
	Indexed bool
}

type TypeKind string

const (
	TypeKindAddress    TypeKind = "address"
	TypeKindBool       TypeKind = "bool"
	TypeKindString     TypeKind = "string"
	TypeKindBytes      TypeKind = "bytes"
	TypeKindInt        TypeKind = "int"
	TypeKindUint       TypeKind = "uint"
	TypeKindArray      TypeKind = "array"
	TypeKindFixedArray TypeKind = "fixedArray"
	TypeKindTuple      TypeKind = "tuple"
)

type Type struct {
	Kind       TypeKind
	Size       int
	Length     int
	Elem       *Type
	Components []Argument
}

func (t Type) String() string {
	switch t.Kind {
	case TypeKindAddress, TypeKindBool, TypeKindString:
		return string(t.Kind)
	case TypeKindBytes:
		if t.Size == 0 {
			return "bytes"
		}
		return fmt.Sprintf("bytes%d", t.Size)
	case TypeKindInt, TypeKindUint:
		if t.Size == 0 {
			return fmt.Sprintf("%s256", t.Kind)
		}
		return fmt.Sprintf("%s%d", t.Kind, t.Size)
	case TypeKindArray:
		if t.Elem == nil {
			return "[]"
		}
		return t.Elem.String() + "[]"
	case TypeKindFixedArray:
		if t.Elem == nil {
			return fmt.Sprintf("[%d]", t.Length)
		}
		return fmt.Sprintf("%s[%d]", t.Elem.String(), t.Length)
	case TypeKindTuple:
		parts := make([]string, 0, len(t.Components))
		for _, component := range t.Components {
			parts = append(parts, component.Type.String())
		}
		return "(" + strings.Join(parts, ",") + ")"
	default:
		if t.Kind == "" {
			return ""
		}
		return string(t.Kind)
	}
}

func (t Type) isArray() bool {
	return t.Kind == TypeKindArray || t.Kind == TypeKindFixedArray
}

func wrapArrayType(elem Type, fixedLength *int) Type {
	if fixedLength == nil {
		return Type{Kind: TypeKindArray, Elem: &elem}
	}
	return Type{Kind: TypeKindFixedArray, Elem: &elem, Length: *fixedLength}
}

func parseTypeString(raw string, components []rawArgument) (Type, error) {
	base, dims, err := splitTypeString(raw)
	if err != nil {
		return Type{}, err
	}

	parsed, err := parseBaseType(base, components)
	if err != nil {
		return Type{}, err
	}

	for i := len(dims) - 1; i >= 0; i-- {
		dim := dims[i]
		parsed = wrapArrayType(parsed, dim)
	}

	return parsed, nil
}

func splitTypeString(raw string) (string, []*int, error) {
	var dims []*int
	rest := raw

	for strings.HasSuffix(rest, "]") {
		start := strings.LastIndex(rest, "[")
		if start < 0 || start == len(rest)-1 {
			return "", nil, fmt.Errorf("invalid ABI type %q", raw)
		}

		token := rest[start+1 : len(rest)-1]
		rest = rest[:start]

		if token == "" {
			dims = append(dims, nil)
			continue
		}

		length, err := strconv.Atoi(token)
		if err != nil || length < 0 {
			return "", nil, fmt.Errorf("invalid ABI array length %q in type %q", token, raw)
		}

		dims = append(dims, &length)
	}

	return rest, dims, nil
}

func parseBaseType(raw string, components []rawArgument) (Type, error) {
	switch {
	case raw == "address":
		return Type{Kind: TypeKindAddress}, nil
	case raw == "bool":
		return Type{Kind: TypeKindBool}, nil
	case raw == "string":
		return Type{Kind: TypeKindString}, nil
	case raw == "bytes":
		return Type{Kind: TypeKindBytes}, nil
	case raw == "tuple":
		children := make([]Argument, 0, len(components))
		for _, component := range components {
			child, err := parseArgument(component)
			if err != nil {
				return Type{}, err
			}
			children = append(children, child)
		}
		return Type{Kind: TypeKindTuple, Components: children}, nil
	case strings.HasPrefix(raw, "bytes"):
		size, err := strconv.Atoi(strings.TrimPrefix(raw, "bytes"))
		if err != nil || size <= 0 || size > 32 {
			return Type{}, fmt.Errorf("invalid fixed bytes type %q", raw)
		}
		return Type{Kind: TypeKindBytes, Size: size}, nil
	case strings.HasPrefix(raw, "uint"):
		size := 256
		suffix := strings.TrimPrefix(raw, "uint")
		if suffix != "" {
			parsed, err := strconv.Atoi(suffix)
			if err != nil || !validIntegerWidth(parsed) {
				return Type{}, fmt.Errorf("invalid uint type %q", raw)
			}
			size = parsed
		}
		return Type{Kind: TypeKindUint, Size: size}, nil
	case strings.HasPrefix(raw, "int"):
		size := 256
		suffix := strings.TrimPrefix(raw, "int")
		if suffix != "" {
			parsed, err := strconv.Atoi(suffix)
			if err != nil || !validIntegerWidth(parsed) {
				return Type{}, fmt.Errorf("invalid int type %q", raw)
			}
			size = parsed
		}
		return Type{Kind: TypeKindInt, Size: size}, nil
	default:
		return Type{}, fmt.Errorf("unsupported ABI type %q", raw)
	}
}

func validIntegerWidth(size int) bool {
	return size > 0 && size <= 256 && size%8 == 0
}

func (abi ABI) lookupMethod(name string) (Method, error) {
	if method, ok := abi.Methods[name]; ok {
		return method, nil
	}
	if method, ok := abi.methodBySignature[name]; ok {
		return method, nil
	}

	overloads := abi.methodOverloads[name]
	if len(overloads) == 0 {
		return Method{}, fmt.Errorf("unknown method %q", name)
	}
	return Method{}, ambiguityError("method", name, methodSignatures(overloads))
}

func (abi ABI) lookupEvent(name string) (Event, error) {
	if event, ok := abi.Events[name]; ok {
		return event, nil
	}
	if event, ok := abi.eventBySignature[name]; ok {
		return event, nil
	}

	overloads := abi.eventOverloads[name]
	if len(overloads) == 0 {
		return Event{}, fmt.Errorf("unknown event %q", name)
	}
	return Event{}, ambiguityError("event", name, eventSignatures(overloads))
}

func ambiguityError(kind string, name string, overloads []string) error {
	sort.Strings(overloads)
	return fmt.Errorf("ambiguous %s %q; use a full signature: %s", kind, name, strings.Join(overloads, ", "))
}

func methodSignatures(methods []Method) []string {
	signatures := make([]string, 0, len(methods))
	for _, method := range methods {
		signatures = append(signatures, signature(method.Name, method.Inputs))
	}
	return signatures
}

func eventSignatures(events []Event) []string {
	signatures := make([]string, 0, len(events))
	for _, event := range events {
		signatures = append(signatures, signature(event.Name, event.Inputs))
	}
	return signatures
}
