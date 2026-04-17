package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestMissingRPCUsesStructuredJSONError(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	err := executeWithEnv(stdout, stderr, func(string) string { return "" }, []string{"rpc", "eth_blockNumber", "--json"})
	if err == nil {
		t.Fatal("expected error")
	}

	want := "{\n  \"error\": \"rpc endpoint required: pass --rpc or set EVMGO_RPC\"\n}\n"
	if stderr.String() != want {
		t.Fatalf("unexpected stderr:\n%s", stderr.String())
	}
}

func TestBalanceAliasHelp(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{"bal", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "Read the native balance") {
		t.Fatalf("unexpected help output:\n%s", stdout.String())
	}
}
