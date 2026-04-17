package abiutil

import "testing"

func TestPackMethod(t *testing.T) {
	contractABI, err := LoadFile("../../testdata/abi/erc20.json")
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	data, err := PackMethod(contractABI, "balanceOf", []string{"0x1111111111111111111111111111111111111111"})
	if err != nil {
		t.Fatalf("PackMethod returned error: %v", err)
	}
	want := "0x70a082310000000000000000000000001111111111111111111111111111111111111111"
	if data != want {
		t.Fatalf("unexpected calldata:\nwant %s\ngot  %s", want, data)
	}
}
