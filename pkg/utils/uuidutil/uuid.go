package uuidutil

import (
	"encoding/base64"
	"encoding/hex"
	"github.com/google/uuid"
	"strings"
)

var escaper = strings.NewReplacer("9", "99", "-", "90", "_", "91")

func UUID() string {
	id := uuid.New()
	return hex.EncodeToString(id[:])
}

// ShortUUID refer to https://stackoverflow.com/questions/37934162/output-uuid-in-go-as-a-short-string
func ShortUUID() string{
	id := uuid.New()
	return escaper.Replace(base64.RawURLEncoding.EncodeToString(id[:]))
}
