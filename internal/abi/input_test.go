package abiutil

import (
	"errors"
	"testing"
)

const selectorCollisionABIJSON = `[
  {
    "type": "function",
    "name": "f29621",
    "inputs": [
      { "name": "account", "type": "address" }
    ],
    "outputs": []
  },
  {
    "type": "function",
    "name": "f43573",
    "inputs": [
      { "name": "amount", "type": "uint256" }
    ],
    "outputs": []
  }
]`

func TestDecodeMethodInput(t *testing.T) {
	contractABI, err := LoadFile("../../testdata/abi/erc20.json")
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	got, err := DecodeMethodInput(contractABI, "0xa9059cbb00000000000000000000000011111111111111111111111111111111111111110000000000000000000000000000000000000000000000000de0b6b3a7640000")
	if err != nil {
		t.Fatalf("DecodeMethodInput returned error: %v", err)
	}
	if got.Method != "transfer(address,uint256)" {
		t.Fatalf("unexpected method: %q", got.Method)
	}
	if len(got.Args) != 2 {
		t.Fatalf("unexpected arg count: %d", len(got.Args))
	}
	if got.Args[0].Name != "to" || got.Args[0].Type != "address" || got.Args[0].Value != "0x1111111111111111111111111111111111111111" {
		t.Fatalf("unexpected first arg: %#v", got.Args[0])
	}
	if got.Args[1].Name != "amount" || got.Args[1].Type != "uint256" || got.Args[1].Value != "1000000000000000000" {
		t.Fatalf("unexpected second arg: %#v", got.Args[1])
	}
}

func TestDecodeMethodInputReturnsSelectorNotFound(t *testing.T) {
	contractABI, err := LoadFile("../../testdata/abi/erc20.json")
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	_, err = DecodeMethodInput(contractABI, "0xffffffff0000000000000000000000001111111111111111111111111111111111111111")
	if err == nil || !errors.Is(err, ErrSelectorNotFound) {
		t.Fatalf("expected selector-not-found error, got %v", err)
	}
}

func TestDecodeMethodInputReturnsMismatch(t *testing.T) {
	contractABI, err := LoadFile("../../testdata/abi/erc20.json")
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	_, err = DecodeMethodInput(contractABI, "0xa9059cbb")
	if err == nil || !errors.Is(err, ErrInputDecodeMismatch) {
		t.Fatalf("expected ErrInputDecodeMismatch, got %v", err)
	}
}

func TestDecodeMethodInputReturnsAmbiguousSelector(t *testing.T) {
	contractABI, err := Load([]byte(selectorCollisionABIJSON))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	_, err = DecodeMethodInput(contractABI, "0xffb7764d0000000000000000000000001111111111111111111111111111111111111111")
	if err == nil || !errors.Is(err, ErrSelectorAmbiguous) {
		t.Fatalf("expected ErrSelectorAmbiguous, got %v", err)
	}
}
