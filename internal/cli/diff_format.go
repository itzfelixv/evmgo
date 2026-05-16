package cli

import (
	"fmt"

	"github.com/itzfelixv/evmgo/internal/actions"
)

func renderStateDiffText(result actions.StateDiffResult, showAll bool) []string {
	if !result.Changed && !showAll {
		return []string{fmt.Sprintf("no changes detected between blocks %s and %s", result.FromBlock, result.ToBlock)}
	}

	lines := []string{
		"from: " + result.FromBlock,
		"to:   " + result.ToBlock,
	}

	if showAll || result.Account.Balance.Changed {
		lines = appendStringDiff(lines, "balance", result.Account.Balance, showAll)
	}
	if showAll || result.Account.Nonce.Changed {
		lines = appendStringDiff(lines, "nonce", result.Account.Nonce, showAll)
	}
	if showAll || result.Account.Code.Changed {
		lines = appendCodeDiff(lines, result.Account.Code, showAll)
	}
	lines = appendStorageDiffs(lines, result.Storage, showAll)

	return lines
}

func appendStringDiff(lines []string, label string, diff actions.StringDiff, showChanged bool) []string {
	lines = append(lines,
		"",
		label+":",
		"  old: "+diff.Old,
		"  new: "+diff.New,
	)
	if showChanged {
		lines = append(lines, "  changed: "+yesNo(diff.Changed))
	}
	return lines
}

func appendCodeDiff(lines []string, diff actions.CodeDiff, showChanged bool) []string {
	lines = append(lines,
		"",
		"code:",
		"  old_hash: "+diff.OldHash,
		"  new_hash: "+diff.NewHash,
	)
	if showChanged {
		lines = append(lines, "  changed: "+yesNo(diff.Changed))
	}
	return lines
}

func appendStorageDiffs(lines []string, storage []actions.StorageDiffEntry, showChanged bool) []string {
	start := len(lines)
	for _, entry := range storage {
		if !showChanged && !entry.Changed {
			continue
		}
		if len(lines) == start {
			lines = append(lines, "", "storage:")
		}
		lines = append(lines,
			"  "+entry.Slot+":",
			"    old: "+entry.Old,
			"    new: "+entry.New,
		)
		if showChanged {
			lines = append(lines, "    changed: "+yesNo(entry.Changed))
		}
	}
	return lines
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
