package abiutil

import (
	"strings"
	"testing"
)

const encodeABIJSON = `[
  {
    "type": "function",
    "name": "setName",
    "inputs": [{ "name": "value", "type": "string" }],
    "outputs": []
  },
  {
    "type": "function",
    "name": "setBlob",
    "inputs": [{ "name": "value", "type": "bytes" }],
    "outputs": []
  },
  {
    "type": "function",
    "name": "setValues",
    "inputs": [{ "name": "value", "type": "uint256[]" }],
    "outputs": []
  },
  {
    "type": "function",
    "name": "setPair",
    "inputs": [
      {
        "name": "pair",
        "type": "tuple",
        "components": [
          { "name": "token", "type": "address" },
          { "name": "amount", "type": "uint256" }
        ]
      }
    ],
    "outputs": []
  },
  {
    "type": "function",
    "name": "quote",
    "inputs": [
      {
        "name": "route",
        "type": "tuple[]",
        "components": [
          { "name": "token", "type": "address" },
          { "name": "weights", "type": "uint256[2]" }
        ]
      }
    ],
    "outputs": []
  },
  {
    "type": "function",
    "name": "setPairName",
    "inputs": [
      {
        "name": "pair",
        "type": "tuple",
        "components": [
          { "name": "amount", "type": "uint256" },
          { "name": "label", "type": "string" }
        ]
      }
    ],
    "outputs": []
  }
]`

func TestPackMethodBalanceOfVector(t *testing.T) {
	contractABI, err := LoadFile("../../testdata/abi/erc20.json")
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	data, err := PackMethod(contractABI, "balanceOf", []string{"0x1111111111111111111111111111111111111111"})
	if err != nil {
		t.Fatalf("PackMethod returned error: %v", err)
	}

	want := "0x70a082310000000000000000000000001111111111111111111111111111111111111111"
	if data != want {
		t.Fatalf("unexpected calldata:\nwant %s\ngot  %s", want, data)
	}
}

func TestPackMethodStringVector(t *testing.T) {
	contractABI := loadEncodeABI(t)

	data, err := PackMethod(contractABI, "setName", []string{"hello"})
	if err != nil {
		t.Fatalf("PackMethod returned error: %v", err)
	}

	want := "0xc47f00270000000000000000000000000000000000000000000000000000000000000020" +
		"0000000000000000000000000000000000000000000000000000000000000005" +
		"68656c6c6f000000000000000000000000000000000000000000000000000000"
	if data != want {
		t.Fatalf("unexpected calldata:\nwant %s\ngot  %s", want, data)
	}
}

func TestPackMethodStringPreservesTopLevelWhitespace(t *testing.T) {
	contractABI := loadEncodeABI(t)

	data, err := PackMethod(contractABI, "setName", []string{" hello "})
	if err != nil {
		t.Fatalf("PackMethod returned error: %v", err)
	}

	want := "0xc47f00270000000000000000000000000000000000000000000000000000000000000020" +
		"0000000000000000000000000000000000000000000000000000000000000007" +
		"2068656c6c6f2000000000000000000000000000000000000000000000000000"
	if data != want {
		t.Fatalf("unexpected calldata:\nwant %s\ngot  %s", want, data)
	}
}

func TestPackMethodBytesVector(t *testing.T) {
	contractABI := loadEncodeABI(t)

	data, err := PackMethod(contractABI, "setBlob", []string{"0x1234"})
	if err != nil {
		t.Fatalf("PackMethod returned error: %v", err)
	}

	want := "0xdd7d5edb0000000000000000000000000000000000000000000000000000000000000020" +
		"0000000000000000000000000000000000000000000000000000000000000002" +
		"1234000000000000000000000000000000000000000000000000000000000000"
	if data != want {
		t.Fatalf("unexpected calldata:\nwant %s\ngot  %s", want, data)
	}
}

func TestPackMethodUintArrayVector(t *testing.T) {
	contractABI := loadEncodeABI(t)

	data, err := PackMethod(contractABI, "setValues", []string{"[1,2,3]"})
	if err != nil {
		t.Fatalf("PackMethod returned error: %v", err)
	}

	want := "0x6da2c5d00000000000000000000000000000000000000000000000000000000000000020" +
		"0000000000000000000000000000000000000000000000000000000000000003" +
		"0000000000000000000000000000000000000000000000000000000000000001" +
		"0000000000000000000000000000000000000000000000000000000000000002" +
		"0000000000000000000000000000000000000000000000000000000000000003"
	if data != want {
		t.Fatalf("unexpected calldata:\nwant %s\ngot  %s", want, data)
	}
}

func TestPackMethodTupleVector(t *testing.T) {
	contractABI := loadEncodeABI(t)

	data, err := PackMethod(contractABI, "setPair", []string{`["0x1111111111111111111111111111111111111111",42]`})
	if err != nil {
		t.Fatalf("PackMethod returned error: %v", err)
	}

	want := "0x0b22dfe40000000000000000000000001111111111111111111111111111111111111111" +
		"000000000000000000000000000000000000000000000000000000000000002a"
	if data != want {
		t.Fatalf("unexpected calldata:\nwant %s\ngot  %s", want, data)
	}
}

func TestPackMethodTupleArrayVector(t *testing.T) {
	contractABI := loadEncodeABI(t)

	data, err := PackMethod(contractABI, "quote", []string{`[["0x1111111111111111111111111111111111111111",[1,2]],["0x2222222222222222222222222222222222222222",[3,4]]]`})
	if err != nil {
		t.Fatalf("PackMethod returned error: %v", err)
	}

	want := "0x6f1ad6ba0000000000000000000000000000000000000000000000000000000000000020" +
		"0000000000000000000000000000000000000000000000000000000000000002" +
		"0000000000000000000000001111111111111111111111111111111111111111" +
		"0000000000000000000000000000000000000000000000000000000000000001" +
		"0000000000000000000000000000000000000000000000000000000000000002" +
		"0000000000000000000000002222222222222222222222222222222222222222" +
		"0000000000000000000000000000000000000000000000000000000000000003" +
		"0000000000000000000000000000000000000000000000000000000000000004"
	if data != want {
		t.Fatalf("unexpected calldata:\nwant %s\ngot  %s", want, data)
	}
}

func TestPackMethodDynamicTupleVector(t *testing.T) {
	contractABI := loadEncodeABI(t)

	data, err := PackMethod(contractABI, "setPairName", []string{`[42,"hello"]`})
	if err != nil {
		t.Fatalf("PackMethod returned error: %v", err)
	}

	want := "0x1a15f7450000000000000000000000000000000000000000000000000000000000000020" +
		"000000000000000000000000000000000000000000000000000000000000002a" +
		"0000000000000000000000000000000000000000000000000000000000000040" +
		"0000000000000000000000000000000000000000000000000000000000000005" +
		"68656c6c6f000000000000000000000000000000000000000000000000000000"
	if data != want {
		t.Fatalf("unexpected calldata:\nwant %s\ngot  %s", want, data)
	}
}

func TestParseJSONLiteralRejectsTrailingGarbage(t *testing.T) {
	if _, err := parseJSONLiteral(`[1]junk`); err == nil {
		t.Fatal("expected parseJSONLiteral to reject trailing garbage")
	}
}

func TestPackMethodRejectsAmbiguousBareName(t *testing.T) {
	contractABI, err := Load([]byte(overloadedABIJSON))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if _, err := PackMethod(contractABI, "swap", []string{"42"}); err == nil || !strings.Contains(err.Error(), "ambiguous method") {
		t.Fatalf("expected ambiguous method error, got %v", err)
	}
}

func TestPackMethodResolvesExactOverloadSignature(t *testing.T) {
	contractABI, err := Load([]byte(overloadedABIJSON))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	data, err := PackMethod(contractABI, "swap(address)", []string{"0x1111111111111111111111111111111111111111"})
	if err != nil {
		t.Fatalf("PackMethod returned error: %v", err)
	}

	want := "0x03438dd00000000000000000000000001111111111111111111111111111111111111111"
	if data != want {
		t.Fatalf("unexpected calldata:\nwant %s\ngot  %s", want, data)
	}
}

func loadEncodeABI(t *testing.T) ABI {
	t.Helper()

	contractABI, err := Load([]byte(encodeABIJSON))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	return contractABI
}
