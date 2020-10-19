package hash

import (
	"crypto/sha256"
	"encoding/hex"
)

// SHA256Hash returns the hexadecimal encoding of sha256 has of salt + input
func SHA256Hash(salt, input string) (string, error) {
	sum := sha256.Sum256([]byte(salt + input))
	return hex.EncodeToString(sum[:]), nil
}
