package eth

import "testing"

func TestNormalizeHash(t *testing.T) {
	got, err := NormalizeHash(" 0x0123456789abcdef0123456789ABCDEF0123456789abcdef0123456789ABCDEF ")
	if err != nil {
		t.Fatalf("NormalizeHash() error = %v", err)
	}

	want := "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if got != want {
		t.Fatalf("NormalizeHash() = %q, want %q", got, want)
	}
}

func TestNormalizeHashRejectsShortHash(t *testing.T) {
	_, err := NormalizeHash("0x1234")
	if err == nil {
		t.Fatal("NormalizeHash() error = nil, want error")
	}
}

func TestNormalizeHashAcceptsBareHash(t *testing.T) {
	got, err := NormalizeHash("0123456789abcdef0123456789ABCDEF0123456789abcdef0123456789ABCDEF")
	if err != nil {
		t.Fatalf("NormalizeHash() error = %v", err)
	}

	want := "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if got != want {
		t.Fatalf("NormalizeHash() = %q, want %q", got, want)
	}
}

func TestNormalizeHashAcceptsUppercasePrefix(t *testing.T) {
	got, err := NormalizeHash("0X0123456789abcdef0123456789ABCDEF0123456789abcdef0123456789ABCDEF")
	if err != nil {
		t.Fatalf("NormalizeHash() error = %v", err)
	}

	want := "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if got != want {
		t.Fatalf("NormalizeHash() = %q, want %q", got, want)
	}
}

func TestIsHash(t *testing.T) {
	if !IsHash("0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef") {
		t.Fatal("IsHash() = false, want true")
	}

	if IsHash("0x1234") {
		t.Fatal("IsHash() = true, want false")
	}
}
