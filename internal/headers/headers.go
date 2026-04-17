package headers

import (
	"bytes"
	"fmt"
	"regexp"
)

type Headers map[string]string

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	// Is this last header?
	// if the remaining data starts with \r\n, it's the end
	if bytes.HasPrefix(data, []byte("\r\n")) {
		return 0, true, nil
	}

	// extract the line (up to \r\n)
	line, _, found := bytes.Cut(data, []byte("\r\n"))
	if !found {
		return 0, false, fmt.Errorf("incomplete header line")
	}

	// extract key and value
	key, value, ok := bytes.Cut(line, []byte(":"))
	if !ok {
		return 0, false, fmt.Errorf("missing colon in header")
	}

	// validate key contains only valid token characters
	validToken := regexp.MustCompile("^[!#$%&'*+\\-.^_`|+~0-9A-Za-z]+$")
	if !validToken.Match(key) {
		return 0, false, fmt.Errorf("invalid character in header key: %s", key)
	}

	// validate value has trailing \r\n (already confirmed since we found \r\n)
	value = bytes.TrimSpace(value)
	h[string(bytes.ToLower(key))] = string(value)

	return len(line) + 2, false, nil
}

func NewHeaders() Headers {
	return make(Headers)
}
