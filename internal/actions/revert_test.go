package actions

import (
	"encoding/json"
	"testing"

	"github.com/itzfelixv/evmgo/internal/rpc"
)

func TestDecodeRevertReasonFromRPCData(t *testing.T) {
	err := &rpc.RPCError{
		Code:    3,
		Message: "execution reverted",
		Data:    json.RawMessage(`"0x08c379a000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000004626f6f6d00000000000000000000000000000000000000000000000000000000"`),
	}

	got := decodeRevertSection(err)
	if got.Decode.Status != "decoded" {
		t.Fatalf("unexpected status: %q", got.Decode.Status)
	}
	if got.Decode.Reason != "boom" {
		t.Fatalf("unexpected reason: %q", got.Decode.Reason)
	}
}

func TestDecodeRevertReasonFallsBackToMessage(t *testing.T) {
	err := &rpc.RPCError{
		Code:    3,
		Message: "execution reverted: ERC20: insufficient allowance",
	}

	got := decodeRevertSection(err)
	if got.Decode.Status != "decoded" {
		t.Fatalf("unexpected status: %q", got.Decode.Status)
	}
	if got.Decode.Reason != "ERC20: insufficient allowance" {
		t.Fatalf("unexpected reason: %q", got.Decode.Reason)
	}
}

func TestDecodeRevertReasonUnavailable(t *testing.T) {
	err := &rpc.RPCError{
		Code:    3,
		Message: "execution reverted",
	}

	got := decodeRevertSection(err)
	if got.Decode.Status != "unavailable" {
		t.Fatalf("unexpected status: %q", got.Decode.Status)
	}
	if got.Decode.Error != "no revert data" {
		t.Fatalf("unexpected error: %q", got.Decode.Error)
	}
}

func TestDecodeRevertReasonMalformedOversizedLengthIsUnavailable(t *testing.T) {
	err := &rpc.RPCError{
		Code:    3,
		Message: "execution reverted",
		Data:    json.RawMessage(`"0x08c379a00000000000000000000000000000000000000000000000000000000000000020ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"`),
	}

	got := decodeRevertSection(err)
	if got.Decode.Status != "unavailable" {
		t.Fatalf("unexpected status: %q", got.Decode.Status)
	}
	if got.Decode.Error != "no revert data" {
		t.Fatalf("unexpected error: %q", got.Decode.Error)
	}
}
