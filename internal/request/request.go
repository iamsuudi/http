package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

const bufferSize = 8

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

type Request struct {
	RequestLine RequestLine
	Headers     map[string]string
	Body        []byte

	initialized bool
	done        bool
}

func (r *Request) parse(data []byte) (int, error) {
	if r.done {
		return 0, fmt.Errorf("trying to read data in a done state")
	}
	if !r.initialized {
		return 0, fmt.Errorf("trying to read data in an uninitialized state")
	}

	// Parse request line if not parsed yet
	if r.RequestLine.Method == "" {
		requestLine, n, err := parseRequestLine(data)
		if n == 0 && err == nil {
			return 0, nil
		} else if err != nil {
			return 0, err
		}
		r.RequestLine = requestLine
		return n, nil
	}

	// Look for end of headers (\r\n\r\n)
	// endOfHeadersIndex := bytes.Index(data, []byte("\r\n\r\n"))
	// if endOfHeadersIndex == -1 {
	// 	// Headers not complete yet
	// 	return 0, nil
	// }

	// We've found the complete request (no body expected for now)
	r.done = true
	return 0, nil
	// return endOfHeadersIndex + 4, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	readToIndex := 0
	buff := make([]byte, bufferSize, bufferSize)

	var request = &Request{
		initialized: true,
	}

	for request.done == false {
		// If buffer is full, grow it to twice its current size
		if readToIndex >= len(buff) {
			newBuff := make([]byte, len(buff)*2, len(buff)*2)
			copy(newBuff, buff)
			buff = newBuff
		}

		n, err := reader.Read(buff[readToIndex:])
		if err != nil {
			// If err is io.EOF, we're done reading
			if errors.Is(err, io.EOF) {
				request.done = true
				break
			} else {
				return nil, fmt.Errorf("failed to read request: %w", err)
			}
		}
		readToIndex += n

		// Call r.parse
		parsedN, err := request.parse(buff[:readToIndex])
		if err != nil {
			return nil, fmt.Errorf("failed to parse request: %w", err)
		} else if parsedN != 0 {
			// Shift unparsed data to the beginning of buffer
			copy(buff, buff[parsedN:])
			readToIndex -= parsedN
		}
	}

	return request, nil
}

func parseRequestLine(str []byte) (RequestLine, int, error) {
	line, _, ok := bytes.Cut(str, []byte("\r\n"))
	if !ok {
		return RequestLine{}, 0, nil
	}
	fields := strings.Fields(string(line))
	if len(fields) != 3 {
		return RequestLine{}, 0, fmt.Errorf("invalid request line: %s", line)
	}

	// Method should be all capital letters
	if fields[0] != strings.ToUpper(fields[0]) {
		return RequestLine{}, 0, fmt.Errorf("invalid request line: method should be all capital letters: %s", line)
	}

	// Http version should be in the format "HTTP/x.y" and for now it must be HTTP/1.1
	if !strings.HasPrefix(fields[2], "HTTP/1.1") {
		return RequestLine{}, 0, fmt.Errorf("invalid request line: http version should be in the format 'HTTP/x.y': %s", line)
	}

	// Extract version number from HTTP/1.1
	_, httpVersion, _ := strings.Cut(fields[2], "/")

	return RequestLine{
		Method:        fields[0],
		RequestTarget: fields[1],
		HttpVersion:   httpVersion,
	}, len(line) + 2, nil
}
