package random

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

func Hex(size int) (string, error) {
	if size <= 0 {
		return "", nil
	}
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

func String(alphabet string, length int) (string, error) {
	alphabet = strings.TrimSpace(alphabet)
	if length <= 0 {
		return "", nil
	}
	if alphabet == "" {
		return "", fmt.Errorf("alphabet is empty")
	}
	max := big.NewInt(int64(len(alphabet)))
	out := make([]byte, length)
	for idx := range out {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[idx] = alphabet[n.Int64()]
	}
	return string(out), nil
}
