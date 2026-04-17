package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogsCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if req.Method != "eth_getLogs" {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if len(req.Params) != 1 {
			t.Fatalf("unexpected param count: %d", len(req.Params))
		}

		wantFilter := `{"address":"0x2222222222222222222222222222222222222222","fromBlock":"0x1","toBlock":"0x2a","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"]}`
		if string(req.Params[0]) != wantFilter {
			t.Fatalf("unexpected log filter:\nwant %s\ngot  %s", wantFilter, req.Params[0])
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":[{"address":"0x2222222222222222222222222222222222222222","blockNumber":"0x2a","transactionHash":"0xaaa","data":"0x00","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"]}]}`))
	}))
	defer server.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{
		"--rpc", server.URL,
		"logs",
		"--address", "2222222222222222222222222222222222222222",
		"--abi", "../../testdata/abi/erc20.json",
		"--event", "Transfer",
		"--from-block", "1",
		"--to-block", "42",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	want := "block=0x2a tx=0xaaa address=0x2222222222222222222222222222222222222222 topics=1\n"
	if stdout.String() != want {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestLogsCommandRejectsInvalidAddress(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(stdout, stderr, func(string) string { return "" })
	cmd.SetArgs([]string{
		"logs",
		"--address", "0x1234",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected invalid address error")
	}

	want := `invalid --address value "0x1234"`
	if err.Error() != want {
		t.Fatalf("unexpected error:\nwant %s\ngot  %v", want, err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}
}
