package httpserver

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func startServer(t *testing.T) (s *Server, addr string) {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	addr = l.Addr().String()
	l.Close()

	s = NewServer()
	go s.Listen(addr)
	<-s.Ready()

	t.Cleanup(func() { s.Shutdown() })
	return s, "http://" + addr
}

func TestBasicGet(t *testing.T) {
	s, addr := startServer(t)

	s.Get("/ping", func(ctx *Context) {
		ctx.Text(200, "pong")
	})

	resp, err := http.Get(addr + "/ping")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, "pong", string(body))
}

func TestGetWithPathParams(t *testing.T) {
	s, addr := startServer(t)

	type User struct {
		Id    int    `json:"id"`
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email,omitempty"`
	}

	user42 := User{
		Id:   42,
		Name: "test42",
		Age:  42,
	}

	user21 := User{
		Id:   21,
		Name: "test21",
		Age:  21,
	}

	users := []User{user42, user21}

	s.Get("/user/:id", func(ctx *Context) {
		id, err := strconv.Atoi(ctx.Params["id"])

		assert.NoError(t, err)

		for _, user := range users {
			if user.Id == id {
				ctx.JSON(200, user)
				return
			}
		}

		ctx.JSON(404, "")

	})

	resp, err := http.Get(addr + "/user/42")
	require.NoError(t, err)
	defer resp.Body.Close()

	jsonUser42, err := json.Marshal(user42)
	assert.NoError(t, err)

	assert.Equal(t, 200, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	assert.Equal(t, jsonUser42, respBody)
}

func TestGetError404(t *testing.T) {
	s, addr := startServer(t)

	s.Get("/", func(ctx *Context) {})

	resp, err := http.Get(addr + "/unknownpath")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 404, resp.StatusCode)
	respBody, _ := io.ReadAll(resp.Body)
	assert.Len(t, respBody, 0)
}

func TestGlobalMiddleware(t *testing.T) {
	s, addr := startServer(t)

	s.Use(func(ctx *Context, next HandlerFunc) {
		newBody := "middleware test check " + string(ctx.Body)
		ctx.Body = []byte(newBody)
		next(ctx)
	})

	s.Post("/check", func(ctx *Context) {
		ctx.Text(200, string(ctx.Body))
	})

	sendingBody := "test text for middle"

	var sendBuff bytes.Buffer
	_, _ = sendBuff.Write([]byte(sendingBody))

	resp, err := http.Post(addr+"/check", "text/plain; charset=utf-8", &sendBuff)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "middleware test check test text for middle", string(respBody))
	assert.Equal(t, 200, resp.StatusCode)
}

func TestErrorToLargeBody(t *testing.T) {
	s, addr := startServer(t)

	sendingBody := make([]byte, 10<<23)
	var sendBuff bytes.Buffer
	_, _ = sendBuff.Write(sendingBody)

	s.Post("/", func(ctx *Context){}) 

	resp, err := http.Post(addr+"/", "text/plain; charset=utf-8", &sendBuff)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	
	assert.Equal(t, 413, resp.StatusCode)
	assert.Len(t, body, 0)
}

func TestErrorInvalidRequest(t *testing.T) {
	_, addr := startServer(t)

	addrIp, _ := strings.CutPrefix(addr, "http://")

	conn, err := net.Dial("tcp", addrIp)
	require.NoError(t, err)
	defer conn.Close()
	conn.Write([]byte("INVALID REQUEST\r\n\r\n"))

	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	response := string(buf[:n])

	assert.Contains(t, response, "400 Bad Request")
}