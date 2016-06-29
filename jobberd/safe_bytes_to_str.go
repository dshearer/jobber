package main

import (
    "encoding/base64"
	"unicode/utf8"
)

func SafeBytesToStr(output []byte) (outputStr string, isBase64 bool) {
	if utf8.Valid(output) {
		return string(output), false
	} else {
		encoded := base64.StdEncoding.EncodeToString(output)
		return encoded, true
	}
}

