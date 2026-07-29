package token

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
)

func GenerateVerificationToken() (string, error) {
	r := make([]byte, 35)
	_, err := rand.Read(r)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(r), nil
}

func GenerateAccountNumber() (string, error) {
	digits := "0123456789"
	r := make([]byte, 10)
	for i := range r {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		r[i] = digits[n.Int64()]
	}
	return string(r), nil

}
