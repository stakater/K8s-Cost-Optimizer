package utils

import (
	"crypto/sha256"
	"fmt"
)

func AsSha256(o interface{}) (string, error) {
	h := sha256.New()
	_, err := h.Write([]byte(fmt.Sprintf("%v", o)))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
