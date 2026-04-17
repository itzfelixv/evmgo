package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCallCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if req.Method != "eth_call" {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if len(req.Params) != 2 {
			t.Fatalf("unexpected param count: %d", len(req.Params))
		}

		wantPayload := `{"data":"0x70a082310000000000000000000000001111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222"}`
		if string(req.Params[0]) != wantPayload {
			t.Fatalf("unexpected call payload:\nwant %s\ngot  %s", wantPayload, req.Params[0])
		}
		if string(req.Params[1]) != `"latest"` {
			t.Fatalf("unexpected block param: %s", req.Params[1])
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x000000000000000000000000000000000000000000000000000000000000000a"}`))
	}))
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{
		"--rpc", server.URL,
		"call",
		"--to", "2222222222222222222222222222222222222222",
		"--abi", "../../testdata/abi/erc20.json",
		"--method", "balanceOf",
		"--args", "0x1111111111111111111111111111111111111111",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	want := "to: 0x2222222222222222222222222222222222222222\nmethod: balanceOf\ndata: 0x70a082310000000000000000000000001111111111111111111111111111111111111111\nblock: latest\nraw: 0x000000000000000000000000000000000000000000000000000000000000000a\ndecoded: 10\n"
	if stdout.String() != want {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestCallCommandRejectsInvalidToAddress(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{
		"call",
		"--to", "0x1234",
		"--abi", "../../testdata/abi/erc20.json",
		"--method", "balanceOf",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected invalid address error")
	}

	want := `invalid --to address "0x1234"`
	if err.Error() != want {
		t.Fatalf("unexpected error:\nwant %s\ngot  %v", want, err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}
}
