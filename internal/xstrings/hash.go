package xstrings

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashStrings(strs ...string) string {
	h := sha256.New()
	for _, s := range strs {
		h.Write([]byte(s))
	}
	return hex.EncodeToString(h.Sum(nil))
}
