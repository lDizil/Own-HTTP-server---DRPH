package httpserver

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type Server struct {
	router      *Router
	middlewares []MiddlewareFunc
	listener    net.Listener
	wg sync.WaitGroup
	closing atomic.Bool
}

func NewServer(router *Router) *Server {
	return &Server{
		router:      router,
		middlewares: []MiddlewareFunc{},
	}
}

func (s *Server) Listen(addr string) error {
	l, err := net.Listen("tcp", addr)

	if err != nil {
		err = fmt.Errorf("Failed to bind to port %s", addr)
		return err
	}

	s.listener = l

	for {
		conn, err := l.Accept()

		if err != nil && s.closing.Load() {
			fmt.Println("Сервер останавливается... Ожидание завершения обработки запросов")
			return nil
		} else if err != nil {
			err = fmt.Errorf("Error accepting connection: %s", err.Error())
			return err
		}

		s.wg.Add(1)
		go s.handleConn(conn)

	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer s.wg.Done()

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
		ctx.RemoteAddr = conn.RemoteAddr().String()

		handler, params := s.router.Match(method, path)

		if shouldClose {
			ctx.responseHeaders["Connection"] = "close"
		}

		if handler == nil {
			ctx.Text(404, "")
		} else {
			ctx.Params = params
			chain := s.buildChain(handler)
			chain(ctx)
		}

		if shouldClose {
			conn.Close()
			break
		}
	}
}

func (s *Server) Get(pattern string, handler HandlerFunc) {
	s.router.Get(pattern, handler)
}

func (s *Server) Post(pattern string, handler HandlerFunc) {
	s.router.Post(pattern, handler)
}

func (s *Server) Use(mw MiddlewareFunc) {
	s.middlewares = append(s.middlewares, mw)
}

func (s *Server) buildChain(handler HandlerFunc) HandlerFunc {
	chain := handler

	for i := len(s.middlewares) - 1; i >= 0; i-- {
		next := chain
		mw := s.middlewares[i]

		chain = func(ctx *Context) {
			mw(ctx, next)
		}
	}

	return chain
}

func (s *Server) Shutdown() {
	s.closing.Store(true)
	s.listener.Close()
	s.wg.Wait()
}