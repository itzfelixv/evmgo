package actions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itzfelixv/evmgo/internal/rpc"
)

func TestCallRawParsesJSONParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			Params json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		if string(req.Params) != `[{"foo":9007199254740993},"bar",9007199254740993]` {
			t.Fatalf("unexpected params: %s", req.Params)
		}

		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":9007199254740993}`))
	}))
	defer server.Close()

	client := rpc.NewClient(server.URL)

	result, err := CallRaw(context.Background(), client, "eth_chainId", []string{`{"foo":9007199254740993}`, `"bar"`, `9007199254740993`})
	if err != nil {
		t.Fatalf("CallRaw returned error: %v", err)
	}

	num, ok := result.(json.Number)
	if !ok {
		t.Fatalf("expected json.Number result, got %#v", result)
	}
	if got, want := num.String(), "9007199254740993"; got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestParseRawParams(t *testing.T) {
	params, err := ParseRawParams([]string{`{"foo":9007199254740993}`, `"bar"`, `9007199254740993`, `not-json`})
	if err != nil {
		t.Fatalf("ParseRawParams returned error: %v", err)
	}

	if got, want := len(params), 4; got != want {
		t.Fatalf("expected %d params, got %d", want, got)
	}

	obj, ok := params[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first param to be a map, got %#v", params[0])
	}
	foo, ok := obj["foo"].(json.Number)
	if !ok {
		t.Fatalf("expected parsed JSON number, got %#v", obj["foo"])
	}
	if got, want := foo.String(), "9007199254740993"; got != want {
		t.Fatalf("expected parsed JSON number %s, got %s", want, got)
	}

	if got, want := params[1], "bar"; got != want {
		t.Fatalf("expected second param %q, got %#v", want, got)
	}

	num, ok := params[2].(json.Number)
	if !ok {
		t.Fatalf("expected third param json.Number, got %#v", params[2])
	}
	if got, want := num.String(), "9007199254740993"; got != want {
		t.Fatalf("expected third param %s, got %s", want, got)
	}

	if got, want := params[3], "not-json"; got != want {
		t.Fatalf("expected fallback string %q, got %#v", want, got)
	}
}

func TestParseRawParamsFallsBackOnTrailingTokens(t *testing.T) {
	params, err := ParseRawParams([]string{
		`truefalse`,
		`9007199254740993junk`,
		`{"a":1} trailing`,
	})
	if err != nil {
		t.Fatalf("ParseRawParams returned error: %v", err)
	}

	if got, want := len(params), 3; got != want {
		t.Fatalf("expected %d params, got %d", want, got)
	}

	for i, want := range []string{
		`truefalse`,
		`9007199254740993junk`,
		`{"a":1} trailing`,
	} {
		if got, ok := params[i].(string); !ok || got != want {
			t.Fatalf("expected fallback string %q at %d, got %#v", want, i, params[i])
		}
	}
}
