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

	// key is case-insensitive has no trailing whitespaces followed by a colon
	// between colon and key there can't be any whitespace
	// then there is vavlue that has one or more trailing whitespaces both prefix and suffix
	matched, err := regexp.Match(`^(\w+):\s*(.+)\s*(\r\n)$`, data)

	if err != nil {
		return 0, false, err
	}
	if !matched {
		return 0, false, fmt.Errorf("Not matched")
	}

	// extract key and value from the match
	line, _, _ := bytes.Cut(data, []byte("\r\n"))
	key, value, ok := bytes.Cut(line, []byte(":"))
	if !ok {
		fmt.Println("Cut error", line)
		return 0, false, nil
	}
	value = bytes.TrimSpace(value)
	// key = bytes.ToLower(key)
	h[string(key)] = string(value)

	return len(line) + 2, false, nil
}

func NewHeaders() Headers {
	return make(Headers)
}
