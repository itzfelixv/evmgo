package eth

import (
	"encoding/hex"

	"golang.org/x/crypto/sha3"
)

func Keccak256(data []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	_, _ = hash.Write(data)
	return hash.Sum(nil)
}

func MethodSelector(signature string) string {
	sum := Keccak256([]byte(signature))
	return hex.EncodeToString(sum[:4])
}

func EventTopic(signature string) string {
	return "0x" + hex.EncodeToString(Keccak256([]byte(signature)))
}
