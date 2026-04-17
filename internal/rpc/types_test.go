package rpc

import "testing"

func TestParseBlockRefHexSelector(t *testing.T) {
	ref, err := ParseBlockRef("0x2a")
	if err != nil {
		t.Fatalf("ParseBlockRef returned error: %v", err)
	}

	if got := ref.RPCArg(); got != "0x2a" {
		t.Fatalf("expected 0x2a, got %q", got)
	}
}

func TestParseBlockRefRejectsLeadingZeroQuantity(t *testing.T) {
	if _, err := ParseBlockRef("0x01"); err == nil {
		t.Fatal("expected leading-zero quantity to be rejected")
	}
}

func TestParseBlockRefRejectsPaddedLeadingZeroQuantity(t *testing.T) {
	if _, err := ParseBlockRef("0x0001"); err == nil {
		t.Fatal("expected padded leading-zero quantity to be rejected")
	}
}
