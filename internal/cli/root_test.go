package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRootHelpIncludesGlobalFlags(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	help := stdout.String()
	for _, needle := range []string{"--rpc", "--json", "evmgo"} {
		if !strings.Contains(help, needle) {
			t.Fatalf("help output missing %q\n%s", needle, help)
		}
	}
}

func TestRootRegistersStateCommands(t *testing.T) {
	cmd := NewRootCmd(new(bytes.Buffer), new(bytes.Buffer), func(string) string { return "" })

	for _, name := range []string{"balance", "code", "storage"} {
		got, _, err := cmd.Find([]string{name})
		if err != nil {
			t.Fatalf("expected to find %q: %v", name, err)
		}
		if got.Name() != name {
			t.Fatalf("expected command %q, got %q", name, got.Name())
		}
	}

	alias, _, err := cmd.Find([]string{"bal"})
	if err != nil {
		t.Fatalf("expected to find alias bal: %v", err)
	}
	if alias.Name() != "balance" {
		t.Fatalf("expected bal alias to resolve to balance, got %q", alias.Name())
	}
}

func TestRootCommandWithoutArgsShowsHelp(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	help := stdout.String()
	for _, needle := range []string{"CLI for read-only EVM interactions", "--rpc", "--json", "evmgo"} {
		if !strings.Contains(help, needle) {
			t.Fatalf("help output missing %q\n%s", needle, help)
		}
	}
}

func TestRootCommandRejectsUnexpectedArgs(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{"garbage"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unexpected positional args")
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	got := stderr.String()
	for _, needle := range []string{"unknown command", "garbage", "evmgo"} {
		if !strings.Contains(got, needle) {
			t.Fatalf("stderr output missing %q\n%s", needle, got)
		}
	}
}

func TestExecuteWritesPlainErrorOnce(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"evmgo", "garbage"}
	defer func() {
		os.Args = oldArgs
	}()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	err := Execute(stdout, stderr)
	if err == nil {
		t.Fatal("expected Execute to return an error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	got := stderr.String()
	want := "Error: unknown command \"garbage\" for \"evmgo\"\n"
	if got != want {
		t.Fatalf("unexpected stderr output:\n%s", got)
	}
}

func TestExecuteWritesJSONError(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"evmgo", "--json", "garbage"}
	defer func() {
		os.Args = oldArgs
	}()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	err := Execute(stdout, stderr)
	if err == nil {
		t.Fatal("expected Execute to return an error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	got := stderr.String()
	want := "{\n  \"error\": \"unknown command \\\"garbage\\\" for \\\"evmgo\\\"\"\n}\n"
	if got != want {
		t.Fatalf("unexpected stderr output:\n%s", got)
	}
}
