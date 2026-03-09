package httpserver

import (
	"bytes"
	"compress/gzip"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytes(t *testing.T) {
	var buf bytes.Buffer

	ctx := NewContext("GET", "/", map[string]string{}, map[string]string{}, nil, &buf)

	ctx.Bytes(200, []byte{})	

	assert.Contains(t, buf.String(), "HTTP/1.1 200 OK")
	assert.Contains(t, buf.String(), "Content-Type: text/plain; charset=utf-8")

	responseBody := strings.Split(buf.String(), "\r\n\r\n")[1]
	assert.Len(t, responseBody, 0)
}

func TestBytesGzip(t *testing.T) {
	var buf bytes.Buffer

	ctx := NewContext("GET", "/", map[string]string{}, map[string]string{}, nil, &buf)

	ctx.Headers["Accept-Encoding"] = "gzip"

	body := "hello test data"
	ctx.Bytes(200, []byte(body))	

	var gzipBuf bytes.Buffer
	writer := gzip.NewWriter(&gzipBuf)
	writer.Write([]byte(body))
	writer.Close()

	assert.Contains(t, buf.String(), "HTTP/1.1 200 OK")
	assert.Contains(t, buf.String(), "Content-Type: text/plain; charset=utf-8")
	assert.Contains(t, buf.String(), "Content-Encoding: gzip")

	responseBody := strings.Split(buf.String(), "\r\n\r\n")[1]
	assert.Len(t, responseBody, len(gzipBuf.Bytes()))
}

func TestText(t *testing.T) {
	var buf bytes.Buffer

	ctx := NewContext("GET", "/", map[string]string{}, map[string]string{}, nil, &buf)

	ctx.Text(200, "test hello")

	assert.Contains(t, buf.String(), "HTTP/1.1 200 OK")
	assert.Contains(t, buf.String(), "Content-Type: text/plain; charset=utf-8")
	assert.Contains(t, buf.String(), "test hello")

	responseBody := strings.Split(buf.String(), "\r\n\r\n")[1]
	assert.NotNil(t, responseBody)
}

func TestJSON(t *testing.T) {
	var buf bytes.Buffer

	ctx := NewContext("GET", "/", map[string]string{}, map[string]string{}, nil, &buf)

	type User struct {
		Name string `json:"name"`
		Age int `json:"age"`
		Email string `json:"email,omitempty"`
	}

	user := User{
		Name: "testname",
		Age: 190,
	}

	ctx.JSON(200, user)
	
	assert.Contains(t, buf.String(), "HTTP/1.1 200 OK")
	assert.Contains(t, buf.String(), "Content-Type: application/json")
	assert.Contains(t, buf.String(), "\"name\":\"testname\",\"age\":190")

	responseBody := strings.Split(buf.String(), "\r\n\r\n")[1]
	assert.NotNil(t, responseBody)
}

func TestFile(t *testing.T) {
	var buf bytes.Buffer

	ctx := NewContext("GET", "/", map[string]string{}, map[string]string{}, nil, &buf)

	dataFromFile := "1 2 3 4 5 6 7 8"

	ctx.File(200, []byte(dataFromFile))
	
	assert.Contains(t, buf.String(), "HTTP/1.1 200 OK")
	assert.Contains(t, buf.String(), "Content-Type: application/octet-stream")
	assert.Contains(t, buf.String(), "1 2 3 4 5 6 7 8")

	responseBody := strings.Split(buf.String(), "\r\n\r\n")[1]
	assert.NotNil(t, responseBody)
}

func TestStatus(t *testing.T) {
	var buf bytes.Buffer

	ctx := NewContext("GET", "/", map[string]string{}, map[string]string{}, nil, &buf)

	ctx.Status(404)

	assert.Equal(t, ctx.status, 404)
}

func TestSetResponseHeader(t *testing.T) {
	var buf bytes.Buffer

	ctx := NewContext("GET", "/", map[string]string{}, map[string]string{}, nil, &buf)

	ctx.SetResponseHeader("Content-Encoding", "gzip")
	ctx.SetResponseHeader("Content-Type", "application/octet-stream")

	assert.Len(t, ctx.responseHeaders, 2)
	assert.Equal(t, "gzip", ctx.responseHeaders["Content-Encoding"])
	assert.Equal(t, "application/octet-stream", ctx.responseHeaders["Content-Type"])
}