package headers

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

type Headers map[string]string

func (h Headers) Get(key string) (string, bool) {
	v, ok := h[strings.ToLower(key)]
	return v, ok
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	totalParsedBytes := 0

	for !done {
		n, done, err = h.parseSingle(data)
		if done {
			totalParsedBytes += 2
			break
		} else if n == 0 && err == nil {
			break
		} else if err != nil {
			break
		}
		totalParsedBytes += n
		data = data[n:]
	}

	return totalParsedBytes, done, err
}

func (h Headers) parseSingle(data []byte) (n int, done bool, err error) {

	// Is this last header?
	// if the remaining data starts with \r\n, it's the end
	if bytes.HasPrefix(data, []byte("\r\n")) {
		return 0, true, nil
	}

	// extract the line (up to \r\n)
	line, _, found := bytes.Cut(data, []byte("\r\n"))
	if !found {
		return 0, false, nil
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
	key = bytes.ToLower(key)
	v, ok := h[string(key)]
	if ok {
		h[string(key)] = v + ", " + string(value)
	} else {
		h[string(key)] = string(value)
	}

	return len(line) + 2, false, nil
}

func NewHeaders() Headers {
	return make(Headers)
}
