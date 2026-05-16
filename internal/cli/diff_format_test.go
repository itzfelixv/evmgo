package cli

import (
	"reflect"
	"strings"
	"testing"

	"github.com/itzfelixv/evmgo/internal/actions"
)

func TestRenderStateDiffTextChangedOnly(t *testing.T) {
	result := actions.StateDiffResult{
		Address:   "0x1111111111111111111111111111111111111111",
		FromBlock: "0x1",
		ToBlock:   "0x2",
		Changed:   true,
		Account: actions.AccountDiffResult{
			Balance: actions.StringDiff{Old: "0x1", New: "0x2", Changed: true},
			Nonce:   actions.StringDiff{Old: "0xa", New: "0xa", Changed: false},
			Code:    actions.CodeDiff{OldHash: "0xcode", NewHash: "0xcode", Changed: false},
		},
		Storage: []actions.StorageDiffEntry{
			{Slot: "0x00", Old: "0xabc", New: "0xdef", Changed: true},
			{Slot: "0x01", Old: "0x123", New: "0x123", Changed: false},
		},
	}

	got := renderStateDiffText(result, false)
	want := []string{
		"from: 0x1",
		"to:   0x2",
		"",
		"balance:",
		"  old: 0x1",
		"  new: 0x2",
		"",
		"storage:",
		"  slot: 0x00",
		"    old: 0xabc",
		"    new: 0xdef",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("renderStateDiffText() = %#v, want %#v", got, want)
	}
	joined := strings.Join(got, "\n")
	for _, hidden := range []string{"nonce:", "code:", "0x01", "changed:"} {
		if strings.Contains(joined, hidden) {
			t.Fatalf("changed-only output unexpectedly contains %q:\n%s", hidden, joined)
		}
	}
}

func TestRenderStateDiffTextShowAllIncludesUnchangedAndMarkers(t *testing.T) {
	result := actions.StateDiffResult{
		Address:   "0x1111111111111111111111111111111111111111",
		FromBlock: "0x1",
		ToBlock:   "0x2",
		Changed:   true,
		Account: actions.AccountDiffResult{
			Balance: actions.StringDiff{Old: "0x1", New: "0x2", Changed: true},
			Nonce:   actions.StringDiff{Old: "0xa", New: "0xa", Changed: false},
			Code:    actions.CodeDiff{OldHash: "0xcode", NewHash: "0xcode", Changed: false},
		},
		Storage: []actions.StorageDiffEntry{
			{Slot: "0x00", Old: "0xabc", New: "0xdef", Changed: true},
			{Slot: "0x01", Old: "0x123", New: "0x123", Changed: false},
		},
	}

	got := renderStateDiffText(result, true)
	want := []string{
		"from: 0x1",
		"to:   0x2",
		"",
		"balance:",
		"  old: 0x1",
		"  new: 0x2",
		"  changed: yes",
		"",
		"nonce:",
		"  old: 0xa",
		"  new: 0xa",
		"  changed: no",
		"",
		"code:",
		"  old_hash: 0xcode",
		"  new_hash: 0xcode",
		"  changed: no",
		"",
		"storage:",
		"  slot: 0x00",
		"    old: 0xabc",
		"    new: 0xdef",
		"    changed: yes",
		"  slot: 0x01",
		"    old: 0x123",
		"    new: 0x123",
		"    changed: no",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("renderStateDiffText() = %#v, want %#v", got, want)
	}
}

func TestRenderStateDiffTextNoChangeChangedOnly(t *testing.T) {
	result := actions.StateDiffResult{
		Address:   "0x1111111111111111111111111111111111111111",
		FromBlock: "0x1",
		ToBlock:   "0x2",
		Changed:   false,
		Account: actions.AccountDiffResult{
			Balance: actions.StringDiff{Old: "0x1", New: "0x1", Changed: false},
			Nonce:   actions.StringDiff{Old: "0xa", New: "0xa", Changed: false},
			Code:    actions.CodeDiff{OldHash: "0xcode", NewHash: "0xcode", Changed: false},
		},
	}

	got := renderStateDiffText(result, false)
	want := []string{"no changes detected between blocks 0x1 and 0x2"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("renderStateDiffText() = %#v, want %#v", got, want)
	}
	if strings.Contains(strings.Join(got, "\n"), result.Address) {
		t.Fatalf("no-change output unexpectedly contains address:\n%s", strings.Join(got, "\n"))
	}
}
