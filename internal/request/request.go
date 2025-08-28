package request

import (
	"bytes"
	"fmt"
	"httpGo/internal/headers"
	"io"
	"slices"
	"strconv"
	"strings"
)

type parserState string

const (
	StateInit   parserState = "init"
	StateHeader parserState = "header"
	StateDone   parserState = "done"
	StateError  parserState = "error"
	StateBody   parserState = "body"
	bufferSize              = 8
)

var (
	ErrorRequestInErrorState = fmt.Errorf("request in error state")
	METHODS                  = []string{"GET", "POST"}
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       parserState
}

func (r *Request) done() bool {
	return r.state == StateDone
}

func newRequest() *Request {
	return &Request{
		Headers: headers.NewHeaders(),
		Body:    []byte{},
		state:   StateInit,
	}
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func IsUpper(s string) bool {
	upper := strings.ToUpper(s)
	return upper == s
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, 8)
	bufIdx := 0

	for !request.done() {
		// Fill up free buffer space
		numRead, err := reader.Read(buf[bufIdx:]) // numRead = 20, buf = "GET /coffee HTTP/1.1"
		if err != nil && err != io.EOF {
			// WARNING: How to handle these errors???
			return nil, err
		}

		if numRead == 0 && err == io.EOF {
			if !request.done() {
				if request.state == StateInit && len(request.RequestLine.Method) == 0 {
					return nil, io.EOF
				}
				return nil, fmt.Errorf("unexpected EOF: request not complete")
			}
			break
		}

		if len(buf) == bufIdx {
			buf = slices.Grow(buf, len(buf))
			buf = buf[:cap(buf)]
		}

		bufIdx += numRead // bufIdx = 20

		// Parse consumes some amount of buf (from beginning buf to bufIdx)
		readN, err := request.parse(buf[:bufIdx]) // readN = 20
		if err != nil {
			return nil, err
		}

		// Move the stuff parse() did not consume to beginning of buffer
		// Up to the bufIdx is consumed by the parser, so shift the stuff after
		// what is read up to the bufIdx, to the beginning of the buffer
		copy(buf, buf[readN:bufIdx])
		// bufIdx should be moved back by the amount that the buffer "shrinks"
		bufIdx -= readN
	}

	return request, nil
}

func ParseRequestLine(request []byte) (*RequestLine, int, error) {
	idx := bytes.Index(request, []byte("\r\n"))
	if idx == -1 {
		return nil, 0, nil
	}

	// Up to and including the CRLF should be "consumed"
	read := idx + len("\r\n")

	rql := request[:idx] // Request line is up to up the CRLF
	splitRql := bytes.Split(rql, []byte(" "))
	if len(splitRql) != 3 {
		return nil, -1, fmt.Errorf("invalid number of HTTP request line components")
	}
	if !IsUpper(string(splitRql[0])) ||
		!slices.Contains(METHODS, string(splitRql[0])) ||
		!bytes.Equal(splitRql[2], []byte("HTTP/1.1")) {
		return nil, -1, fmt.Errorf("HTTP request line content is invalid")
	}

	version := bytes.Split(splitRql[2], []byte("/"))[1]

	return &RequestLine{
		HttpVersion:   string(version),
		RequestTarget: string(splitRql[1]),
		Method:        string(splitRql[0]),
	}, read, nil
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
	var err error

outer:
	for {
		switch r.state {
		case StateError:
			return 0, ErrorRequestInErrorState
		case StateInit:
			rl, n, err := ParseRequestLine(data[read:])
			if err != nil {
				r.state = StateError
				return 0, err
			}
			// unable to find \r\n, then return(0, nil) and
			// RequestFromReader will be able to keep reading into its 1024
			// byte buffer and giving it to us (as long as len(request line) < 1024
			// we should be able to parse it out)
			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n

			r.state = StateHeader
		case StateHeader:
			for !r.done() {
				n, done, err := r.Headers.ParseSingle(data[read:])
				if err != nil {
					r.state = StateError
					return 0, err
				}

				if n == 0 {
					break outer
				}

				read += n
				if done {
					r.state = StateBody
					break
				}
			}
		case StateBody:
			val, found := r.Headers.Get("Content-Length")
			if !found {
				r.state = StateDone
				break outer
			}

			contentLength, err := strconv.Atoi(val)
			if err != nil {
				r.state = StateError
				return 0, err
			}

			consumed := 0
			if len(r.Body) < contentLength {
				diff := contentLength - len(r.Body)
				consumed = min(diff, len(data[read:]))
				r.Body = append(r.Body, data[read:read+consumed]...)
			}

			read += consumed

			if len(r.Body) == contentLength {
				r.state = StateDone
			}

			break outer
		case StateDone:
			break outer
		default:
			return 0, fmt.Errorf("error: unknown state")
		}
	}

	return read, err
}
