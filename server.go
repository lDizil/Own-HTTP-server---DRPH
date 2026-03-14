package httpserver

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/lDizil/Own-HTTP-server---DRPH/metrics"
)

type Server struct {
	router      *Router
	middlewares []MiddlewareFunc
	listener    net.Listener
	wg          sync.WaitGroup
	closing     atomic.Bool
	MaxBodySize int
	ready       chan struct{}
	conns       map[net.Conn]bool
	connsMu     sync.Mutex
}

func NewServer() *Server {
	return &Server{
		router:      &Router{},
		middlewares: []MiddlewareFunc{},
		MaxBodySize: 10 << 20,
		ready:       make(chan struct{}),
		conns:       make(map[net.Conn]bool),
	}
}

func (s *Server) Listen(addr string) error {
	l, err := net.Listen("tcp", addr)

	if err != nil {
		err = fmt.Errorf("Failed to bind to port %s", addr)
		return err
	}

	s.listener = l

	fmt.Printf("Сервер запущен на %s\n", addr)
	close(s.ready)

	for {
		conn, err := l.Accept()

		if err != nil && s.closing.Load() {
			fmt.Println("Сервер останавливается... Ожидание завершения обработки запросов")
			return nil
		} else if err != nil {
			err = fmt.Errorf("Error accepting connection: %s", err.Error())
			return err
		}

		metrics.HttpActiveConn.Inc()
		
		s.wg.Add(1)
		go s.handleConn(conn)

	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer s.wg.Done()
	defer metrics.HttpActiveConn.Dec()

	s.connsMu.Lock()
	s.conns[conn] = false
	s.connsMu.Unlock()

	defer func() {
		s.connsMu.Lock()
		delete(s.conns, conn)
		s.connsMu.Unlock()
	}()

	reader := bufio.NewReader(conn)

	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		headers := make(map[string]string)

		var method string
		var path string
		var body []byte

		var badRequest bool

		pathLine, err := reader.ReadString('\n')

		s.connsMu.Lock()
		s.conns[conn] = true
		s.connsMu.Unlock()

		conn.SetDeadline(time.Now().Add(10 * time.Second))

		if err != nil {
			conn.Close()
			break
		}

		pathElements := strings.Split(strings.TrimSpace(pathLine), " ")

		if len(pathElements) < 3 {
			conn.Write([]byte("HTTP/1.1 400 Bad Request\r\nContent-Length: 0\r\nConnection: close\r\n\r\n"))
			conn.Close()
			break
		}

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

			reqPartsHeaders := strings.SplitN(headersLine, ": ", 2)

			if len(reqPartsHeaders) < 2 {
				conn.Write([]byte("HTTP/1.1 400 Bad Request\r\nContent-Length: 0\r\nConnection: close\r\n\r\n"))
				conn.Close()
				badRequest = true
				break
			}

			headers[textproto.CanonicalMIMEHeaderKey(reqPartsHeaders[0])] = strings.TrimSpace(reqPartsHeaders[1])

		}

		if badRequest {
			break
		}

		var shouldClose bool

		if val, ok := headers["Connection"]; ok {
			if val == "close" {
				shouldClose = true
			}
		}

		if lenStr, ok := headers["Content-Length"]; ok {
			length, err := strconv.Atoi(strings.TrimSpace(lenStr))

			if err != nil || length < 0 {
				conn.Write([]byte("HTTP/1.1 400 Bad Request\r\nContent-Length: 0\r\nConnection: close\r\n\r\n"))
				conn.Close()
				break
			} else if length > s.MaxBodySize {
				conn.Write([]byte("HTTP/1.1 413 Content Too Large\r\nContent-Length: 0\r\nConnection: close\r\n\r\n"))
				conn.Close()
				break
			}

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

		s.connsMu.Lock()
		s.conns[conn] = false
		s.connsMu.Unlock()

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

func (s *Server) Put(pattern string, handler HandlerFunc) {
	s.router.Put(pattern, handler)
}

func (s *Server) Delete(pattern string, handler HandlerFunc) {
	s.router.Delete(pattern, handler)
}

func (s *Server) Patch(pattern string, handler HandlerFunc) {
	s.router.Patch(pattern, handler)
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

	s.connsMu.Lock()
	for conn, active := range s.conns {
		if !active {
			conn.Close()
		}
	}
	s.connsMu.Unlock()

	s.wg.Wait()
}

func (s *Server) Run(port string) {
	go s.Listen("0.0.0.0:" + port)

	<-s.Ready()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	s.Shutdown()

	fmt.Println("Сервер остановлен")
}

func (s *Server) Group(prefix string) *Group {
	return &Group{
		routers: s,
		prefix:  prefix,
	}
}

func (s *Server) Ready() <-chan struct{} {
	return s.ready
}
