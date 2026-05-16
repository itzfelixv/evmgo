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
		FromBlock: "0x64",
		ToBlock:   "0xc8",
		Changed:   true,
		Account: actions.AccountDiffResult{
			Balance: actions.StringDiff{Old: "0x1", New: "0x2", Changed: true},
			Nonce:   actions.StringDiff{Old: "0xa", New: "0xa", Changed: false},
			Code:    actions.CodeDiff{OldHash: "0xcode", NewHash: "0xcode", Changed: false},
		},
		Storage: []actions.StorageDiffEntry{
			{Slot: "0x0", Old: "0x00", New: "0x01", Changed: true},
			{Slot: "0x1", Old: "0x02", New: "0x02", Changed: false},
		},
	}

	got := renderStateDiffText(result, false)
	joined := strings.Join(got, "\n")
	for _, visible := range []string{"from: 0x64", "to:   0xc8", "balance:", "storage:", "0x0:"} {
		if !strings.Contains(joined, visible) {
			t.Fatalf("changed-only output missing %q:\n%s", visible, joined)
		}
	}
	for _, hidden := range []string{"nonce:", "code:", "0x1:", "changed:"} {
		if strings.Contains(joined, hidden) {
			t.Fatalf("changed-only output unexpectedly contains %q:\n%s", hidden, joined)
		}
	}
}

func TestRenderStateDiffTextShowAllIncludesUnchangedAndMarkers(t *testing.T) {
	result := actions.StateDiffResult{
		Address:   "0x1111111111111111111111111111111111111111",
		FromBlock: "0x64",
		ToBlock:   "0xc8",
		Changed:   true,
		Account: actions.AccountDiffResult{
			Balance: actions.StringDiff{Old: "0x1", New: "0x2", Changed: true},
			Nonce:   actions.StringDiff{Old: "0xa", New: "0xa", Changed: false},
			Code:    actions.CodeDiff{OldHash: "0xcode", NewHash: "0xcode", Changed: false},
		},
		Storage: []actions.StorageDiffEntry{
			{Slot: "0x0", Old: "0x00", New: "0x01", Changed: true},
			{Slot: "0x1", Old: "0x02", New: "0x02", Changed: false},
		},
	}

	got := renderStateDiffText(result, true)
	joined := strings.Join(got, "\n")
	for _, visible := range []string{"balance:", "changed: yes", "nonce:", "changed: no", "code:", "storage:"} {
		if !strings.Contains(joined, visible) {
			t.Fatalf("showAll output missing %q:\n%s", visible, joined)
		}
	}
}

func TestRenderStateDiffTextNoChangeChangedOnly(t *testing.T) {
	result := actions.StateDiffResult{
		Address:   "0x1111111111111111111111111111111111111111",
		FromBlock: "19000000",
		ToBlock:   "19000100",
		Changed:   false,
		Account: actions.AccountDiffResult{
			Balance: actions.StringDiff{Old: "0x1", New: "0x1", Changed: false},
			Nonce:   actions.StringDiff{Old: "0xa", New: "0xa", Changed: false},
			Code:    actions.CodeDiff{OldHash: "0xcode", NewHash: "0xcode", Changed: false},
		},
	}

	got := renderStateDiffText(result, false)
	want := []string{"no changes detected between blocks 19000000 and 19000100"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("renderStateDiffText() = %#v, want %#v", got, want)
	}
	if strings.Contains(strings.Join(got, "\n"), result.Address) {
		t.Fatalf("no-change output unexpectedly contains address:\n%s", strings.Join(got, "\n"))
	}
}
