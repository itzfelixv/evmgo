package abiutil

import (
	"strings"
	"testing"
)

const decodeABIJSON = `[
  {
    "type": "function",
    "name": "isReady",
    "inputs": [],
    "outputs": [{ "name": "value", "type": "bool" }]
  },
  {
    "type": "function",
    "name": "name",
    "inputs": [],
    "outputs": [{ "name": "value", "type": "string" }]
  },
  {
    "type": "function",
    "name": "blob",
    "inputs": [],
    "outputs": [{ "name": "value", "type": "bytes" }]
  },
  {
    "type": "function",
    "name": "fixedBlob",
    "inputs": [],
    "outputs": [{ "name": "value", "type": "bytes4" }]
  },
  {
    "type": "function",
    "name": "values",
    "inputs": [],
    "outputs": [{ "name": "value", "type": "uint256[]" }]
  },
  {
    "type": "function",
    "name": "pair",
    "inputs": [],
    "outputs": [
      {
        "name": "value",
        "type": "tuple",
        "components": [
          { "name": "token", "type": "address" },
          { "name": "amount", "type": "uint256" }
        ]
      }
    ]
  },
  {
    "type": "function",
    "name": "pairName",
    "inputs": [],
    "outputs": [
      {
        "name": "value",
        "type": "tuple",
        "components": [
          { "name": "amount", "type": "uint256" },
          { "name": "label", "type": "string" }
        ]
      }
    ]
  },
  {
    "type": "function",
    "name": "summary",
    "inputs": [],
    "outputs": [
      { "name": "ok", "type": "bool" },
      {
        "name": "pairs",
        "type": "tuple[]",
        "components": [
          { "name": "token", "type": "address" },
          { "name": "amount", "type": "uint256" }
        ]
      },
      { "name": "data", "type": "bytes" }
    ]
  }
]`

func TestDecodeMethodOutputUint256(t *testing.T) {
	contractABI, err := LoadFile("../../testdata/abi/erc20.json")
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	raw := "0x000000000000000000000000000000000000000000000000000000000000002a"
	got, err := DecodeMethodOutput(contractABI, "balanceOf", raw)
	if err != nil {
		t.Fatalf("DecodeMethodOutput returned error: %v", err)
	}
	if len(got) != 1 || got[0] != "42" {
		t.Fatalf("unexpected decoded output: %#v", got)
	}
}

func TestDecodeMethodOutputBool(t *testing.T) {
	contractABI := loadDecodeABI(t)

	raw := "0x0000000000000000000000000000000000000000000000000000000000000001"
	got, err := DecodeMethodOutput(contractABI, "isReady", raw)
	if err != nil {
		t.Fatalf("DecodeMethodOutput returned error: %v", err)
	}
	if len(got) != 1 || got[0] != "true" {
		t.Fatalf("unexpected decoded output: %#v", got)
	}
}

func TestDecodeMethodOutputString(t *testing.T) {
	contractABI := loadDecodeABI(t)

	raw := "0x0000000000000000000000000000000000000000000000000000000000000020" +
		"0000000000000000000000000000000000000000000000000000000000000005" +
		"68656c6c6f000000000000000000000000000000000000000000000000000000"
	got, err := DecodeMethodOutput(contractABI, "name", raw)
	if err != nil {
		t.Fatalf("DecodeMethodOutput returned error: %v", err)
	}
	if len(got) != 1 || got[0] != "hello" {
		t.Fatalf("unexpected decoded output: %#v", got)
	}
}

func TestDecodeMethodOutputBytes(t *testing.T) {
	contractABI := loadDecodeABI(t)

	raw := "0x0000000000000000000000000000000000000000000000000000000000000020" +
		"0000000000000000000000000000000000000000000000000000000000000002" +
		"1234000000000000000000000000000000000000000000000000000000000000"
	got, err := DecodeMethodOutput(contractABI, "blob", raw)
	if err != nil {
		t.Fatalf("DecodeMethodOutput returned error: %v", err)
	}
	if len(got) != 1 || got[0] != "0x1234" {
		t.Fatalf("unexpected decoded output: %#v", got)
	}
}

func TestDecodeMethodOutputArray(t *testing.T) {
	contractABI := loadDecodeABI(t)

	raw := "0x0000000000000000000000000000000000000000000000000000000000000020" +
		"0000000000000000000000000000000000000000000000000000000000000003" +
		"0000000000000000000000000000000000000000000000000000000000000001" +
		"0000000000000000000000000000000000000000000000000000000000000002" +
		"0000000000000000000000000000000000000000000000000000000000000003"
	got, err := DecodeMethodOutput(contractABI, "values", raw)
	if err != nil {
		t.Fatalf("DecodeMethodOutput returned error: %v", err)
	}
	if len(got) != 1 || got[0] != `[1,2,3]` {
		t.Fatalf("unexpected decoded output: %#v", got)
	}
}

func TestDecodeMethodOutputTuple(t *testing.T) {
	contractABI := loadDecodeABI(t)

	raw := "0x0000000000000000000000001111111111111111111111111111111111111111" +
		"000000000000000000000000000000000000000000000000000000000000002a"
	got, err := DecodeMethodOutput(contractABI, "pair", raw)
	if err != nil {
		t.Fatalf("DecodeMethodOutput returned error: %v", err)
	}
	if len(got) != 1 || got[0] != `["0x1111111111111111111111111111111111111111",42]` {
		t.Fatalf("unexpected decoded output: %#v", got)
	}
}

func TestDecodeMethodOutputDynamicTuple(t *testing.T) {
	contractABI := loadDecodeABI(t)

	raw := "0x0000000000000000000000000000000000000000000000000000000000000020" +
		"000000000000000000000000000000000000000000000000000000000000002a" +
		"0000000000000000000000000000000000000000000000000000000000000040" +
		"0000000000000000000000000000000000000000000000000000000000000005" +
		"68656c6c6f000000000000000000000000000000000000000000000000000000"
	got, err := DecodeMethodOutput(contractABI, "pairName", raw)
	if err != nil {
		t.Fatalf("DecodeMethodOutput returned error: %v", err)
	}
	if len(got) != 1 || got[0] != `[42,"hello"]` {
		t.Fatalf("unexpected decoded output: %#v", got)
	}
}

func TestDecodeMethodOutputMixedValues(t *testing.T) {
	contractABI := loadDecodeABI(t)

	raw := "0x0000000000000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000000000000000060" +
		"0000000000000000000000000000000000000000000000000000000000000100" +
		"0000000000000000000000000000000000000000000000000000000000000002" +
		"0000000000000000000000001111111111111111111111111111111111111111" +
		"000000000000000000000000000000000000000000000000000000000000002a" +
		"0000000000000000000000002222222222222222222222222222222222222222" +
		"0000000000000000000000000000000000000000000000000000000000000063" +
		"0000000000000000000000000000000000000000000000000000000000000004" +
		"deadbeef00000000000000000000000000000000000000000000000000000000"
	got, err := DecodeMethodOutput(contractABI, "summary", raw)
	if err != nil {
		t.Fatalf("DecodeMethodOutput returned error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("unexpected decoded output length: %#v", got)
	}
	if got[0] != "false" {
		t.Fatalf("unexpected bool output: %q", got[0])
	}
	if got[1] != `[["0x1111111111111111111111111111111111111111",42],["0x2222222222222222222222222222222222222222",99]]` {
		t.Fatalf("unexpected tuple array output: %q", got[1])
	}
	if got[2] != "0xdeadbeef" {
		t.Fatalf("unexpected bytes output: %q", got[2])
	}
}

func TestDecodeMethodOutputRejectsDynamicOffsetAliasingIntoHead(t *testing.T) {
	contractABI := loadDecodeABI(t)

	raw := "0x0000000000000000000000000000000000000000000000000000000000000020" +
		"0000000000000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000000000000000000"
	if _, err := DecodeMethodOutput(contractABI, "pairName", raw); err == nil {
		t.Fatal("expected DecodeMethodOutput to reject dynamic offset aliasing into head")
	}
}

func TestDecodeMethodOutputRejectsAddressPadding(t *testing.T) {
	contractABI := loadDecodeABI(t)

	raw := "0x0100000000000000000000001111111111111111111111111111111111111111" +
		"000000000000000000000000000000000000000000000000000000000000002a"
	if _, err := DecodeMethodOutput(contractABI, "pair", raw); err == nil {
		t.Fatal("expected DecodeMethodOutput to reject non-zero address padding")
	}
}

func TestDecodeMethodOutputRejectsFixedBytesPadding(t *testing.T) {
	contractABI := loadDecodeABI(t)

	raw := "0x1234567801000000000000000000000000000000000000000000000000000000"
	if _, err := DecodeMethodOutput(contractABI, "fixedBlob", raw); err == nil {
		t.Fatal("expected DecodeMethodOutput to reject non-zero fixed-bytes padding")
	}
}

func TestDecodeMethodOutputRejectsDynamicBytesPadding(t *testing.T) {
	contractABI := loadDecodeABI(t)

	raw := "0x0000000000000000000000000000000000000000000000000000000000000020" +
		"0000000000000000000000000000000000000000000000000000000000000002" +
		"1234010000000000000000000000000000000000000000000000000000000000"
	if _, err := DecodeMethodOutput(contractABI, "blob", raw); err == nil {
		t.Fatal("expected DecodeMethodOutput to reject non-zero dynamic bytes padding")
	}
}

func TestDecodeMethodOutputRejectsMisalignedDynamicOffset(t *testing.T) {
	contractABI := loadDecodeABI(t)

	raw := "0x0000000000000000000000000000000000000000000000000000000000000021" +
		"00" +
		"0000000000000000000000000000000000000000000000000000000000000002" +
		"1234000000000000000000000000000000000000000000000000000000000000"
	if _, err := DecodeMethodOutput(contractABI, "blob", raw); err == nil {
		t.Fatal("expected DecodeMethodOutput to reject misaligned dynamic offset")
	}
}

func TestDecodeMethodOutputRejectsStringPadding(t *testing.T) {
	contractABI := loadDecodeABI(t)

	raw := "0x0000000000000000000000000000000000000000000000000000000000000020" +
		"0000000000000000000000000000000000000000000000000000000000000005" +
		"68656c6c6f010000000000000000000000000000000000000000000000000000"
	if _, err := DecodeMethodOutput(contractABI, "name", raw); err == nil {
		t.Fatal("expected DecodeMethodOutput to reject non-zero string padding")
	}
}

func TestDecodeMethodOutputRejectsAmbiguousBareName(t *testing.T) {
	contractABI, err := Load([]byte(overloadedABIJSON))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	raw := "0x0000000000000000000000000000000000000000000000000000000000000001"
	if _, err := DecodeMethodOutput(contractABI, "swap", raw); err == nil || !strings.Contains(err.Error(), "ambiguous method") {
		t.Fatalf("expected ambiguous method error, got %v", err)
	}
}

func TestDecodeMethodOutputResolvesExactOverloadSignature(t *testing.T) {
	contractABI, err := Load([]byte(overloadedABIJSON))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	raw := "0x0000000000000000000000000000000000000000000000000000000000000001"
	got, err := DecodeMethodOutput(contractABI, "swap(address)", raw)
	if err != nil {
		t.Fatalf("DecodeMethodOutput returned error: %v", err)
	}
	if len(got) != 1 || got[0] != "true" {
		t.Fatalf("unexpected decoded output: %#v", got)
	}
}

func loadDecodeABI(t *testing.T) ABI {
	t.Helper()

	contractABI, err := Load([]byte(decodeABIJSON))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	return contractABI
}
