package config

import "testing"

func TestResolveRPCURLUsesFlagFirst(t *testing.T) {
	flags := GlobalFlags{RPC: "https://rpc.example"}

	got, err := ResolveRPCURL(flags, func(string) string { return "https://env.example" })
	if err != nil {
		t.Fatalf("ResolveRPCURL returned error: %v", err)
	}
	if got != "https://rpc.example" {
		t.Fatalf("expected flag value, got %q", got)
	}
}

func TestResolveRPCURLFallsBackToEnv(t *testing.T) {
	got, err := ResolveRPCURL(GlobalFlags{}, func(key string) string {
		if key == "EVMGO_RPC" {
			return "https://env.example"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("ResolveRPCURL returned error: %v", err)
	}
	if got != "https://env.example" {
		t.Fatalf("expected env value, got %q", got)
	}
}

func TestResolveRPCURLRequiresValue(t *testing.T) {
	_, err := ResolveRPCURL(GlobalFlags{}, func(string) string { return "" })
	if err == nil {
		t.Fatal("expected missing endpoint error")
	}
	if err.Error() != "rpc endpoint required: pass --rpc or set EVMGO_RPC" {
		t.Fatalf("unexpected error: %v", err)
	}
}
