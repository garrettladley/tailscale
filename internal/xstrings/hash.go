package xstrings

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashStrings(delimiter string, strs ...string) string {
	var (
		h       = sha256.New()
		lenStrs = len(strs)
	)
	for i, s := range strs {
		if i != lenStrs-1 {
			h.Write([]byte(s + delimiter))
		} else {
			h.Write([]byte(s))
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}
