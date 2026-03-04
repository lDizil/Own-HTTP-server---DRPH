package main

import "io"

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
}

func NewContext() *Context {
	return &Context{}
}
func (c *Context) Text(code int, body string) {
	if code != 0 {
		c.status = code
	}

	if header, ok := c.Headers["Content-Type"]; ok {
		c.SetHeader(header, c.Headers[header])
	} else {
		c.SetHeader("Content-Type", "text/plain; charset=utf-8")
	}

	n := len()
}

func (c *Context) Status(code int) {
	c.status = code
}

func (c *Context) SetHeader(key, value string) {
	c.responseHeaders[key] = value
}