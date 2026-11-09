// Package identifier validates identifiers without database-driver dependencies.
package identifier

import (
	"encoding/hex"
	"strings"
)

func ValidUUID(value string) bool {
	if len(value) == 36 {
		if value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
			return false
		}
		value = strings.ReplaceAll(value, "-", "")
	}
	if len(value) != 32 {
		return false
	}
	var decoded [16]byte
	_, err := hex.Decode(decoded[:], []byte(value))
	return err == nil
}
