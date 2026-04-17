package eth

import "testing"

func TestNormalizeAddress(t *testing.T) {
	got, err := NormalizeAddress(" 0x0123456789ABCDEF0123456789abcdef01234567 ")
	if err != nil {
		t.Fatalf("NormalizeAddress() error = %v", err)
	}

	want := "0x0123456789abcdef0123456789abcdef01234567"
	if got != want {
		t.Fatalf("NormalizeAddress() = %q, want %q", got, want)
	}
}

func TestNormalizeAddressRejectsBadLength(t *testing.T) {
	_, err := NormalizeAddress("0x1234")
	if err == nil {
		t.Fatal("NormalizeAddress() error = nil, want error")
	}
}

func TestIsAddress(t *testing.T) {
	if !IsAddress("0x0123456789abcdef0123456789abcdef01234567") {
		t.Fatal("IsAddress() = false, want true")
	}

	if IsAddress("0x1234") {
		t.Fatal("IsAddress() = true, want false")
	}
}
