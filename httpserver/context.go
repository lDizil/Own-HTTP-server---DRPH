package httpserver

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Context struct {
	Method          string
	Path            string
	Params          map[string]string
	Query           map[string]string
	Headers         map[string]string
	responseHeaders map[string]string
	Body            []byte
	writer          io.Writer
	status          int
	RemoteAddr       string
}

func NewContext(method, path string, query map[string]string, headers map[string]string, body []byte, writer io.Writer) *Context {
	return &Context{
		Method:          method,
		Path:            path,
		Params:          make(map[string]string),
		Query:           query,
		Headers:         headers,
		responseHeaders: make(map[string]string),
		Body:            body,
		writer:          writer,
		status:          200,
		RemoteAddr: "Неизвестный адрес",
	}
}

func (c *Context) Bytes(code int, body []byte) {
	if code != 0 {
		c.status = code
	}

	var encodings []string

	if val, ok := c.Headers["Accept-Encoding"]; ok {
		encodings = strings.Split(val, ", ")
	}

	for _, encode := range encodings {
		if encode == "gzip" {
			var buf bytes.Buffer

			gzipWriter := gzip.NewWriter(&buf)

			_, _ = gzipWriter.Write(body)

			gzipWriter.Close()

			body = buf.Bytes()

			c.responseHeaders["Content-Encoding"] = "gzip"
		}
	}

	if _, ok := c.responseHeaders["Content-Type"]; !ok {
		c.SetResponseHeader("Content-Type", "text/plain; charset=utf-8")
	}

	n := len(body)

	c.SetResponseHeader("Content-Length", strconv.Itoa(n))

	sliceOfHeadersLine := []string{}

	for answHeader, val := range c.responseHeaders {
		sliceOfHeadersLine = append(sliceOfHeadersLine, fmt.Sprintf("%s: %s", answHeader, val))
	}

	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", code, StatusText(code))
	headersLine := strings.Join(sliceOfHeadersLine, "\r\n") + "\r\n\r\n"

	c.writer.Write([]byte(statusLine))
	c.writer.Write([]byte(headersLine))
	c.writer.Write(body)
}

func (c *Context) Text(code int, body string) {
	c.responseHeaders["Content-Type"] = "text/plain; charset=utf-8"
	c.Bytes(code, []byte(body))
}

func (c *Context) JSON(code int, v any) {
	data, err := json.Marshal(v)

	if err != nil {
		c.Bytes(500, []byte("internal error"))
		return
	}

	c.responseHeaders["Content-Type"] = "application/json"

	c.Bytes(code, data)
}

func (c *Context) File(code int, data []byte) {
	c.responseHeaders["Content-Type"] = "application/octet-stream"
	c.Bytes(code, data)
}

func (c *Context) Status(code int) {
	c.status = code
}

func (c *Context) SetResponseHeader(key, value string) {
	c.responseHeaders[key] = value
}

func (c *Context) StatusCode() int {
	return c.status
}
