package abiutil

import "testing"

func TestTypeStringCanonical(t *testing.T) {
	tuple := Type{
		Kind: TypeKindTuple,
		Components: []Argument{
			{Type: Type{Kind: TypeKindAddress}},
			{Type: Type{Kind: TypeKindUint, Size: 256}},
		},
	}

	cases := []struct {
		name string
		typ  Type
		want string
	}{
		{name: "address", typ: Type{Kind: TypeKindAddress}, want: "address"},
		{name: "uint256", typ: Type{Kind: TypeKindUint, Size: 256}, want: "uint256"},
		{name: "bytes4", typ: Type{Kind: TypeKindBytes, Size: 4}, want: "bytes4"},
		{name: "dynamic array", typ: Type{Kind: TypeKindArray, Elem: &Type{Kind: TypeKindAddress}}, want: "address[]"},
		{name: "fixed array", typ: Type{Kind: TypeKindFixedArray, Elem: &Type{Kind: TypeKindUint, Size: 256}, Length: 2}, want: "uint256[2]"},
		{name: "tuple array", typ: Type{Kind: TypeKindArray, Elem: &tuple}, want: "(address,uint256)[]"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.typ.String(); got != tc.want {
				t.Fatalf("Type.String() = %q, want %q", got, tc.want)
			}
		})
	}
}
