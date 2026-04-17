package eth

import "testing"

func TestMethodSelector(t *testing.T) {
	got := MethodSelector("balanceOf(address)")
	want := "70a08231"
	if got != want {
		t.Fatalf("MethodSelector() = %q, want %q", got, want)
	}
}

func TestMethodSelectorTransfer(t *testing.T) {
	got := MethodSelector("transfer(address,uint256)")
	want := "a9059cbb"
	if got != want {
		t.Fatalf("MethodSelector() = %q, want %q", got, want)
	}
}

func TestEventTopic(t *testing.T) {
	got := EventTopic("Transfer(address,address,uint256)")
	want := "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	if got != want {
		t.Fatalf("EventTopic() = %q, want %q", got, want)
	}
}
