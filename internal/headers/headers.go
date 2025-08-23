package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func isValidToken(s string) bool {
	for index, runeValue := range s {
	}
}

func (h Headers) Set(key string, val string) error {
	newKey := strings.ToLower(key)
	h[newKey] = val

	return nil
}

func (h Headers) Get(key string) string {
	return ""
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	read := 0
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return read, false, nil
	}

	if idx == 0 {
		return read, true, nil
	}

	firstColonIdx := bytes.Index(data, []byte(":"))
	if data[firstColonIdx-1] == ' ' {
		return read, true, fmt.Errorf("cannot have space after header name, before the colon")
	}

	headerKey := bytes.TrimSpace(data[:firstColonIdx])
	headerVal := bytes.TrimSpace(data[firstColonIdx+1 : idx])
	err = h.Set(string(headerKey), string(headerVal))
	if err != nil {
		return read, true, err
	}

	read += idx + 2
	return read, false, nil
}
