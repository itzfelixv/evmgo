package cli

import "testing"

func TestValidateTransactionHashRejectsShortHash(t *testing.T) {
	err := validateTransactionHash("0x1234")
	if err == nil {
		t.Fatal("expected short hash to be rejected")
	}
}

func TestValidateTransactionHashAcceptsBareHash(t *testing.T) {
	err := validateTransactionHash("1111111111111111111111111111111111111111111111111111111111111111")
	if err != nil {
		t.Fatalf("expected bare hash to be accepted: %v", err)
	}
}

func TestValidateTransactionHashAcceptsUppercasePrefix(t *testing.T) {
	err := validateTransactionHash("0X1111111111111111111111111111111111111111111111111111111111111111")
	if err != nil {
		t.Fatalf("expected 0X-prefixed hash to be accepted: %v", err)
	}
}

func TestValidateTransactionHashRejectsWrappedWhitespace(t *testing.T) {
	err := validateTransactionHash(" 0x1111111111111111111111111111111111111111111111111111111111111111 ")
	if err == nil {
		t.Fatal("expected whitespace-wrapped hash to be rejected")
	}
}

func TestValidateStorageSlotAcceptsHexBytes(t *testing.T) {
	if err := validateStorageSlot("0x0102"); err != nil {
		t.Fatalf("expected hex bytes slot to be accepted: %v", err)
	}
}

func TestValidateStorageSlotRejectsWrappedWhitespace(t *testing.T) {
	err := validateStorageSlot(" 0x0102 ")
	if err == nil {
		t.Fatal("expected whitespace-wrapped slot to be rejected")
	}
}
