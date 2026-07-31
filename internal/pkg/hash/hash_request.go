package hash

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashRequest(path string, method string, body []byte) string {
	sh := sha256.New()
	sh.Write([]byte(path))
	sh.Write([]byte(method))
	sh.Write(body)
	return hex.EncodeToString(sh.Sum(nil))
}
