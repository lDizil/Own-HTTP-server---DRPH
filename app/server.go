package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

type server struct {
	router  *Router
	dirName string
}

func NewServer(router *Router, dirName string) *server {
	return &server{
		router:  router,
		dirName: dirName,
	}
}

func (s *server) Listen(addr string) error {
	l, err := net.Listen("tcp", addr)

	if err != nil {
		err = fmt.Errorf("Failed to bind to port %s", addr)
		return err
	}


	for {
		conn, err := l.Accept()

		if err != nil {
			err = fmt.Errorf("Error accepting connection: %s", err.Error())
			return err
		}

		go s.handleConn(conn)

	}
}

func (s *server) handleConn(conn net.Conn) {
	reader := bufio.NewReader(conn)

	for {
		headers := make(map[string]string)

		var method string
		var path string
		var body []byte

		pathLine, err := reader.ReadString('\n')
		
		
		if err != nil {
			conn.Close()
			break
		}

		pathElements := strings.Split(pathLine, " ")
		
		method, path = pathElements[0], pathElements[1]

		var query map[string]string

		if strings.Contains(path, "?") {
			rawQuery := strings.Split(path, "?")[1]
			pairs := strings.Split(rawQuery, "&")

			query = make(map[string]string, len(pairs))
			
			for _, pair := range pairs {
				kv := strings.SplitN(pair, "=", 2)
				if len(kv) == 2 {
					query[kv[0]] = kv[1]
				}
			}
	 	}

		for {

			headersLine, _ := reader.ReadString('\n')

			if headersLine == "\r\n" {
				break
			}

			reqPartsHeaders := strings.Split(headersLine, ": ")
			headers[reqPartsHeaders[0]] = strings.TrimSpace(reqPartsHeaders[1])

		}

		var shouldClose bool

		if val, ok := headers["Connection"]; ok {
			if val == "close" {
				shouldClose = true
			}
		}

		if lenStr, ok := headers["Content-Length"]; ok {
			length, _ := strconv.Atoi(strings.TrimSpace(lenStr))
			body = make([]byte, length)
			io.ReadFull(reader, body)
		}

		ctx := NewContext(method, path, query, headers, body, conn)

		handler, params := s.router.Match(method, path)

		if shouldClose {	
			ctx.responseHeaders["Connection"] = "close"
		}

		if handler == nil {
			ctx.Text(404, "")
		} else {
			ctx.Params = params
			handler(ctx)
		}

		if shouldClose {	
			conn.Close()
			break
		}
	}
}

func (s *server) Get(pattern string, handler HandlerFunc) {
	s.router.Get(pattern, handler)
}

func (s *server) Post(pattern string, handler HandlerFunc) {
	s.router.Post(pattern, handler)
}
