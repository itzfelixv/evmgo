package abiutil

import (
	"strings"
	"testing"
)

func TestMethodSignatureBalanceOf(t *testing.T) {
	contractABI, err := LoadFile("../../testdata/abi/erc20.json")
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	got, err := MethodSignature(contractABI, "balanceOf")
	if err != nil {
		t.Fatalf("MethodSignature returned error: %v", err)
	}
	if got != "balanceOf(address)" {
		t.Fatalf("unexpected signature: %q", got)
	}
}

func TestEventSignatureTransfer(t *testing.T) {
	contractABI, err := LoadFile("../../testdata/abi/erc20.json")
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	got, err := EventSignature(contractABI, "Transfer")
	if err != nil {
		t.Fatalf("EventSignature returned error: %v", err)
	}
	if got != "Transfer(address,address,uint256)" {
		t.Fatalf("unexpected signature: %q", got)
	}
}

func TestMethodSelectorBalanceOf(t *testing.T) {
	contractABI, err := LoadFile("../../testdata/abi/erc20.json")
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	got, err := MethodSelector(contractABI, "balanceOf")
	if err != nil {
		t.Fatalf("MethodSelector returned error: %v", err)
	}
	if got != "0x70a08231" {
		t.Fatalf("unexpected selector: %q", got)
	}
}

func TestMethodSelectorComposite(t *testing.T) {
	contractABI, err := Load([]byte(compositeABIJSON))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	got, err := MethodSelector(contractABI, "quote")
	if err != nil {
		t.Fatalf("MethodSelector returned error: %v", err)
	}
	if got != "0x6f1ad6ba" {
		t.Fatalf("unexpected selector: %q", got)
	}
}

func TestEventTopicTransfer(t *testing.T) {
	contractABI, err := LoadFile("../../testdata/abi/erc20.json")
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	got, err := EventTopic(contractABI, "Transfer")
	if err != nil {
		t.Fatalf("EventTopic returned error: %v", err)
	}
	want := "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	if got != want {
		t.Fatalf("unexpected topic: %q", got)
	}
}

func TestMethodSignatureComposite(t *testing.T) {
	contractABI, err := Load([]byte(compositeABIJSON))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	got, err := MethodSignature(contractABI, "quote")
	if err != nil {
		t.Fatalf("MethodSignature returned error: %v", err)
	}
	if got != "quote((address,uint256[2])[])" {
		t.Fatalf("unexpected signature: %q", got)
	}
}

func TestEventSignatureComposite(t *testing.T) {
	contractABI, err := Load([]byte(compositeABIJSON))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	got, err := EventSignature(contractABI, "Quoted")
	if err != nil {
		t.Fatalf("EventSignature returned error: %v", err)
	}
	if got != "Quoted((address,uint256[2])[])" {
		t.Fatalf("unexpected signature: %q", got)
	}
}

func TestMethodSignatureUniqueMethodInOverloadedABI(t *testing.T) {
	contractABI, err := Load([]byte(overloadedABIJSON))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	got, err := MethodSignature(contractABI, "quote")
	if err != nil {
		t.Fatalf("MethodSignature returned error: %v", err)
	}
	if got != "quote(uint256)" {
		t.Fatalf("unexpected signature: %q", got)
	}
}

func TestMethodSignatureRejectsAmbiguousBareName(t *testing.T) {
	contractABI, err := Load([]byte(overloadedABIJSON))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if _, err := MethodSignature(contractABI, "swap"); err == nil || !strings.Contains(err.Error(), "ambiguous method") {
		t.Fatalf("expected ambiguous method error, got %v", err)
	}
}

func TestMethodSignatureResolvesExactOverloadSignature(t *testing.T) {
	contractABI, err := Load([]byte(overloadedABIJSON))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	got, err := MethodSignature(contractABI, "swap(address)")
	if err != nil {
		t.Fatalf("MethodSignature returned error: %v", err)
	}
	if got != "swap(address)" {
		t.Fatalf("unexpected signature: %q", got)
	}
}

func TestEventSignatureUniqueEventInOverloadedABI(t *testing.T) {
	contractABI, err := Load([]byte(overloadedABIJSON))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	got, err := EventSignature(contractABI, "Quoted")
	if err != nil {
		t.Fatalf("EventSignature returned error: %v", err)
	}
	if got != "Quoted(uint256)" {
		t.Fatalf("unexpected signature: %q", got)
	}
}

func TestEventSignatureRejectsAmbiguousBareName(t *testing.T) {
	contractABI, err := Load([]byte(overloadedABIJSON))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if _, err := EventSignature(contractABI, "Filled"); err == nil || !strings.Contains(err.Error(), "ambiguous event") {
		t.Fatalf("expected ambiguous event error, got %v", err)
	}
}

func TestEventSignatureResolvesExactOverloadSignature(t *testing.T) {
	contractABI, err := Load([]byte(overloadedABIJSON))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	got, err := EventSignature(contractABI, "Filled(address)")
	if err != nil {
		t.Fatalf("EventSignature returned error: %v", err)
	}
	if got != "Filled(address)" {
		t.Fatalf("unexpected signature: %q", got)
	}
}

func TestEventTopicRejectsAnonymousEvent(t *testing.T) {
	contractABI, err := Load([]byte(overloadedABIJSON))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if _, err := EventTopic(contractABI, "AnonymousFilled"); err == nil || !strings.Contains(err.Error(), "anonymous event") {
		t.Fatalf("expected anonymous event error, got %v", err)
	}
}
