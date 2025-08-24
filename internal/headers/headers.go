package headers

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func isValidToken(s string) bool {
	for _, runeVal := range s {
		switch {
		case unicode.IsLetter(runeVal):
			fallthrough
		case unicode.IsDigit(runeVal):
			continue
		}

		switch runeVal {
		case '!', '#', '$', '%', '&', '\'', '*',
			'+', '-', '.', '^', '_', '`', '|', '~':
			continue
		}

		return false
	}

	return true
}

func (h Headers) Set(key string, val string) error {
	newKey := strings.ToLower(key)
	if !isValidToken(newKey) {
		return fmt.Errorf("invalid header token")
	}

	if oldVal, ok := h[newKey]; ok {
		newVal := fmt.Sprintf("%s, %s", oldVal, val)
		h[newKey] = newVal
	} else {
		h[newKey] = val
	}

	return nil
}

func (h Headers) Get(key string) string {
	newKey := strings.ToLower(key)
	if val, ok := h[newKey]; ok {
		return val
	}
	return ""
}

func (h *Headers) ParseSingle(data []byte) (n int, done bool, err error) {
	read := 0
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return read, false, nil
	}

	if idx == 0 {
		return 2, true, nil
	}

	firstColonIdx := bytes.Index(data, []byte(":"))
	if data[firstColonIdx-1] == ' ' {
		return read, true, fmt.Errorf("cannot have space after header name, before the colon")
	}

	headerKey := bytes.TrimSpace(data[:firstColonIdx])
	headerVal := bytes.TrimSpace(data[firstColonIdx+1 : idx])

	// fmt.Printf("key: '%s', val '%s'\n", headerKey, headerVal)
	err = h.Set(string(headerKey), string(headerVal))
	if err != nil {
		return read, true, err
	}

	read += idx + 2
	return read, false, nil
}
