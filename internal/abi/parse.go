package abiutil

import (
	"encoding/json"
	"fmt"
	"os"
)

type rawABIEntry struct {
	Type            string        `json:"type"`
	Name            string        `json:"name,omitempty"`
	Inputs          []rawArgument `json:"inputs,omitempty"`
	Outputs         []rawArgument `json:"outputs,omitempty"`
	Anonymous       bool          `json:"anonymous,omitempty"`
	StateMutability string        `json:"stateMutability,omitempty"`
}

type rawArgument struct {
	Name       string        `json:"name,omitempty"`
	Type       string        `json:"type"`
	Indexed    bool          `json:"indexed,omitempty"`
	Components []rawArgument `json:"components,omitempty"`
}

func LoadFile(path string) (ABI, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ABI{}, fmt.Errorf("read abi file: %w", err)
	}

	parsed, err := Load(data)
	if err != nil {
		return ABI{}, err
	}
	return parsed, nil
}

func Load(data []byte) (ABI, error) {
	var entries []rawABIEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return ABI{}, fmt.Errorf("parse abi file: %w", err)
	}

	contractABI := ABI{
		Methods:           map[string]Method{},
		Events:            map[string]Event{},
		methodOverloads:   map[string][]Method{},
		eventOverloads:    map[string][]Event{},
		methodBySignature: map[string]Method{},
		eventBySignature:  map[string]Event{},
	}

	for _, entry := range entries {
		switch entry.Type {
		case "function":
			method, err := parseMethod(entry)
			if err != nil {
				return ABI{}, err
			}
			methodSignature := signature(method.Name, method.Inputs)
			if _, ok := contractABI.methodBySignature[methodSignature]; ok {
				return ABI{}, fmt.Errorf("duplicate method signature %q", methodSignature)
			}
			contractABI.methodOverloads[method.Name] = append(contractABI.methodOverloads[method.Name], method)
			contractABI.methodBySignature[methodSignature] = method
		case "event":
			event, err := parseEvent(entry)
			if err != nil {
				return ABI{}, err
			}
			eventSignature := signature(event.Name, event.Inputs)
			if _, ok := contractABI.eventBySignature[eventSignature]; ok {
				return ABI{}, fmt.Errorf("duplicate event signature %q", eventSignature)
			}
			contractABI.eventOverloads[event.Name] = append(contractABI.eventOverloads[event.Name], event)
			contractABI.eventBySignature[eventSignature] = event
		default:
			continue
		}
	}

	for name, overloads := range contractABI.methodOverloads {
		if len(overloads) == 1 {
			contractABI.Methods[name] = overloads[0]
		}
	}
	for name, overloads := range contractABI.eventOverloads {
		if len(overloads) == 1 {
			contractABI.Events[name] = overloads[0]
		}
	}

	return contractABI, nil
}

func parseMethod(entry rawABIEntry) (Method, error) {
	inputs, err := parseArguments(entry.Inputs)
	if err != nil {
		return Method{}, err
	}
	outputs, err := parseArguments(entry.Outputs)
	if err != nil {
		return Method{}, err
	}
	return Method{Name: entry.Name, Inputs: inputs, Outputs: outputs}, nil
}

func parseEvent(entry rawABIEntry) (Event, error) {
	inputs, err := parseArguments(entry.Inputs)
	if err != nil {
		return Event{}, err
	}
	return Event{Name: entry.Name, Inputs: inputs, Anonymous: entry.Anonymous}, nil
}

func parseArguments(entries []rawArgument) ([]Argument, error) {
	args := make([]Argument, 0, len(entries))
	for _, entry := range entries {
		arg, err := parseArgument(entry)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}
	return args, nil
}

func parseArgument(entry rawArgument) (Argument, error) {
	typ, err := parseTypeString(entry.Type, entry.Components)
	if err != nil {
		return Argument{}, err
	}
	return Argument{
		Name:    entry.Name,
		Type:    typ,
		Indexed: entry.Indexed,
	}, nil
}
