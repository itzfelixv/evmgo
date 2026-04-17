package abiutil

import (
	"fmt"
	"strings"
	"testing"
)

const compositeABIJSON = `[
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
    "outputs": [
      {
        "name": "",
        "type": "tuple",
        "components": [
          { "name": "ok", "type": "bool" },
          { "name": "data", "type": "bytes" }
        ]
      }
    ]
  },
  {
    "type": "event",
    "name": "Quoted",
    "inputs": [
      {
        "name": "route",
        "type": "tuple[]",
        "components": [
          { "name": "token", "type": "address" },
          { "name": "weights", "type": "uint256[2]" }
        ]
      }
    ]
  }
]`

const overloadedABIJSON = `[
  {
    "type": "function",
    "name": "swap",
    "inputs": [{ "name": "amount", "type": "uint256" }],
    "outputs": [{ "name": "ok", "type": "bool" }]
  },
  {
    "type": "function",
    "name": "swap",
    "inputs": [{ "name": "token", "type": "address" }],
    "outputs": [{ "name": "ok", "type": "bool" }]
  },
  {
    "type": "function",
    "name": "quote",
    "inputs": [{ "name": "amount", "type": "uint256" }],
    "outputs": [{ "name": "amountOut", "type": "uint256" }]
  },
  {
    "type": "event",
    "name": "Filled",
    "inputs": [{ "name": "amount", "type": "uint256", "indexed": true }]
  },
  {
    "type": "event",
    "name": "Filled",
    "inputs": [{ "name": "token", "type": "address", "indexed": true }]
  },
  {
    "type": "event",
    "name": "Quoted",
    "inputs": [{ "name": "amount", "type": "uint256" }]
  },
  {
    "type": "event",
    "name": "AnonymousFilled",
    "anonymous": true,
    "inputs": [{ "name": "maker", "type": "address", "indexed": true }]
  }
]`

func TestLoadFileParsesMethodAndEvent(t *testing.T) {
	contractABI, err := LoadFile("../../testdata/abi/erc20.json")
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}
	if _, ok := contractABI.Methods["balanceOf"]; !ok {
		t.Fatal("expected balanceOf method")
	}
	if _, ok := contractABI.Events["Transfer"]; !ok {
		t.Fatal("expected Transfer event")
	}
}

func TestLoadParsesCompositeTypes(t *testing.T) {
	contractABI, err := Load([]byte(compositeABIJSON))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	method, ok := contractABI.Methods["quote"]
	if !ok {
		t.Fatal("expected quote method")
	}
	if got := method.Inputs[0].Type.String(); got != "(address,uint256[2])[]" {
		t.Fatalf("quote input type = %q, want %q", got, "(address,uint256[2])[]")
	}
	if got := method.Outputs[0].Type.String(); got != "(bool,bytes)" {
		t.Fatalf("quote output type = %q, want %q", got, "(bool,bytes)")
	}

	event, ok := contractABI.Events["Quoted"]
	if !ok {
		t.Fatal("expected Quoted event")
	}
	if got := event.Inputs[0].Type.String(); got != "(address,uint256[2])[]" {
		t.Fatalf("Quoted input type = %q, want %q", got, "(address,uint256[2])[]")
	}
}

func TestLoadRejectsInvalidABIType(t *testing.T) {
	cases := []string{"bytes33", "uint7", "int264"}

	for _, typ := range cases {
		t.Run(typ, func(t *testing.T) {
			data := fmt.Appendf(nil, `[
  {
    "type": "function",
    "name": "broken",
    "inputs": [{ "name": "value", "type": %q }],
    "outputs": []
  }
]`, typ)
			if _, err := Load(data); err == nil {
				t.Fatalf("expected Load to reject ABI type %q", typ)
			}
		})
	}
}

func TestLoadAllowsOverloadsAndPreservesAnonymousMetadata(t *testing.T) {
	contractABI, err := Load([]byte(overloadedABIJSON))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if _, ok := contractABI.Methods["quote"]; !ok {
		t.Fatal("expected unique quote method lookup")
	}
	if _, ok := contractABI.Events["Quoted"]; !ok {
		t.Fatal("expected unique Quoted event lookup")
	}

	if _, err := contractABI.lookupMethod("swap"); err == nil || !strings.Contains(err.Error(), "ambiguous method") {
		t.Fatalf("expected ambiguous method lookup error, got %v", err)
	}
	if _, err := contractABI.lookupEvent("Filled"); err == nil || !strings.Contains(err.Error(), "ambiguous event") {
		t.Fatalf("expected ambiguous event lookup error, got %v", err)
	}

	event, err := contractABI.lookupEvent("AnonymousFilled(address)")
	if err != nil {
		t.Fatalf("lookupEvent returned error: %v", err)
	}
	if !event.Anonymous {
		t.Fatal("expected anonymous event metadata to be preserved")
	}
}
