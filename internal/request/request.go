package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/iamsuudi/httpfromtcp/internal/headers"
)

const bufferSize = 8

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte

	initialized          bool
	bodyCompleted        bool
	headerCompleted      bool
	requestLineCompleted bool
	bodyBytesRead        int
}

func (r *Request) Done() bool {
	return r.requestLineCompleted && r.headerCompleted && r.bodyCompleted
}

func (r *Request) InitializeBodyBuffer() error {
	v, ok := r.Headers.Get("content-length")
	if !ok {
		r.bodyCompleted = true
		return nil
	}

	contentLength, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return err
	}

	// Make sure body buffer is large enough
	r.Body = make([]byte, contentLength)
	if contentLength == 0 {
		r.bodyCompleted = true
		return nil
	}

	return nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.Done() {
		return 0, fmt.Errorf("trying to read data in a done state")
	}
	if !r.initialized {
		return 0, fmt.Errorf("trying to read data in an uninitialized state")
	}

	switch {
	case !r.requestLineCompleted:
		// Parse request line if not parsed yet
		requestLine, n, err := parseRequestLine(data)
		if n == 0 && err == nil {
			return 0, nil
		} else if err != nil {
			return 0, err
		}
		r.RequestLine = requestLine
		r.requestLineCompleted = true
		return n, nil
	case !r.headerCompleted:
		// Parse headers if not parsed yet or if headers are incomplete
		n, done, err := r.Headers.Parse(data)
		if done {
			r.headerCompleted = done
			return n, r.InitializeBodyBuffer()
		}
		return n, err
	case !r.bodyCompleted:
		// If adding this data would exceed content-length, throw error
		if r.bodyBytesRead+len(data) > len(r.Body) {
			return 0, fmt.Errorf("content exceeded specified content-length: %d > %d", r.bodyBytesRead+len(data), len(r.Body))
		}

		// Read content into body buffer
		n := copy(r.Body[r.bodyBytesRead:], data)
		r.bodyBytesRead += n

		// Check if body buffer is full and mark body completed
		if r.bodyBytesRead == len(r.Body) {
			r.bodyCompleted = true
		}
		return n, nil
	case !r.headerCompleted:
		// Parse headers if not parsed yet or if headers are incomplete
		n, done, err := r.Headers.Parse(data)
		if done {
			r.headerCompleted = done
		}
		return n, err
	}

	return 0, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	readToIndex := 0
	buff := make([]byte, bufferSize, bufferSize)

	var request = &Request{
		initialized: true,
		Headers:     headers.NewHeaders(),
	}

	for !request.Done() {
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
				// If content is less than content-length return error
				if request.bodyBytesRead < len(request.Body) {
					return nil, fmt.Errorf("EOF before reading full body: got %d, expected %d", request.bodyBytesRead, len(request.Body))
				}
				request.bodyCompleted = true
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
