package main

import (
	"crypto/sha256"
	"encoding/hex"
)

func calculateHash(record string) string {
	h := sha256.New()
	h.Write([]byte(record))
	hash := h.Sum(nil)
	return hex.EncodeToString(hash)
}
