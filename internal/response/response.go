package response

import (
	"fmt"
	"httpGo/internal/headers"
	"io"
)

type (
	StatusCode  int
	WriterState string
	Writer      struct {
		writer      io.Writer
		WriterState WriterState
	}
)

const (
	StatusLine = "Status"
	Header     = "Header"
	Body       = "Body"
)

const (
	OK          StatusCode = 200
	BadRequest  StatusCode = 400
	ServerError StatusCode = 500
)

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer:      w,
		WriterState: StatusLine,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	var (
		err error
	)

	switch statusCode {
	case OK:
		_, err = w.writer.Write([]byte("HTTP/1.1 200 OK\r\n"))
	case BadRequest:
		_, err = w.writer.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
	case ServerError:
		_, err = w.writer.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
	default:
		_, err = w.writer.Write(fmt.Appendf([]byte{}, "HTTP/1.1 %d \r\n", statusCode))
	}

	if err != nil {
		return err
	}

	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")

	return headers
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	for key, val := range headers {
		_, err := w.writer.Write(fmt.Appendf([]byte{}, "%s: %s\r\n", key, val))
		if err != nil {
			return err
		}
	}

	_, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	if err != nil {
		return 0, err
	}

	return n, err
}
