package eth

import (
	"math/big"
	"testing"
)

func TestDecodeQuantity(t *testing.T) {
	got, err := DecodeQuantity("0x2a")
	if err != nil {
		t.Fatalf("DecodeQuantity() error = %v", err)
	}

	want := big.NewInt(42)
	if got.Cmp(want) != 0 {
		t.Fatalf("DecodeQuantity() = %v, want %v", got, want)
	}
}

func TestDecodeQuantityRejectsSignedInput(t *testing.T) {
	_, err := DecodeQuantity("0x-2a")
	if err == nil {
		t.Fatal("DecodeQuantity() error = nil, want error")
	}
}

func TestEncodeQuantity(t *testing.T) {
	got := EncodeQuantity(big.NewInt(42))
	want := "0x2a"
	if got != want {
		t.Fatalf("EncodeQuantity() = %q, want %q", got, want)
	}
}

func TestEncodeQuantityNil(t *testing.T) {
	got := EncodeQuantity(nil)
	want := "0x0"
	if got != want {
		t.Fatalf("EncodeQuantity() = %q, want %q", got, want)
	}
}

func TestEncodeQuantityPanicsOnNegativeValues(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("EncodeQuantity() did not panic for negative value")
		}
	}()

	EncodeQuantity(big.NewInt(-42))
}

func TestDecodeHexBytes(t *testing.T) {
	got, err := DecodeHexBytes("0x0102")
	if err != nil {
		t.Fatalf("DecodeHexBytes() error = %v", err)
	}

	want := []byte{0x01, 0x02}
	if len(got) != len(want) {
		t.Fatalf("DecodeHexBytes() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("DecodeHexBytes()[%d] = 0x%02x, want 0x%02x", i, got[i], want[i])
		}
	}
}

func TestDecodeHexBytesOddLength(t *testing.T) {
	got, err := DecodeHexBytes("0xabc")
	if err != nil {
		t.Fatalf("DecodeHexBytes() error = %v", err)
	}

	want := []byte{0x0a, 0xbc}
	if len(got) != len(want) {
		t.Fatalf("DecodeHexBytes() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("DecodeHexBytes()[%d] = 0x%02x, want 0x%02x", i, got[i], want[i])
		}
	}
}

func TestEncodeHexBytes(t *testing.T) {
	got := EncodeHexBytes([]byte{0x01, 0x02})
	want := "0x0102"
	if got != want {
		t.Fatalf("EncodeHexBytes() = %q, want %q", got, want)
	}
}
