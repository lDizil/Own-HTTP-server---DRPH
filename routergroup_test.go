package httpserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouterGroupBasic(t *testing.T) {
	s := NewServer()

	group := s.Group("/api")
	group.Use(func(ctx *Context, next HandlerFunc) {})

	assert.Len(t, s.router.Routes, 0)
	assert.Len(t, group.middlewares, 1)

	assert.Equal(t, group.routers, s)
	assert.Equal(t, "/api", group.prefix)

	v1 := group.Group("/v1")
	v1.Use(func(ctx *Context, next HandlerFunc) {})

	assert.Len(t, s.router.Routes, 0)
	assert.Len(t, v1.middlewares, 2)

	assert.Equal(t, v1.routers, s)
	assert.Equal(t, "/api/v1", v1.prefix)

	v1.Get("/user/:id", func(ctx *Context) {})

	assert.Len(t, s.router.Routes, 1)
}

func TestBuildGroupChain(t *testing.T) {
	group := &Group{}

	var calls []string

	group.Use(func(ctx *Context, next HandlerFunc) {
		calls = append(calls, "first")
		next(ctx)
	}, 
		func(ctx *Context, next HandlerFunc) {
			calls = append(calls, "second")
			next(ctx)
	})

	handler := func(ctx *Context) {
		calls = append(calls, "handler")
	}

	chain := group.buildGroupChain(handler)
	chain(&Context{})

	assert.Equal(t, []string{"first", "second", "handler"}, calls)
}


func TestBuildGroupChainWithoutMiddleware(t *testing.T) {
	group := &Group{}

	var calls bool

	handler := func(ctx *Context) {
		calls = true
	}

	chain := group.buildGroupChain(handler)
	chain(&Context{})

	assert.True(t, calls)
}

func TestUse(t *testing.T) {
	group := &Group{}
	
	assert.Len(t, group.middlewares, 0)

	group.Use(func(ctx *Context, next HandlerFunc) {
		next(ctx)
	}, 
		func(ctx *Context, next HandlerFunc) {
			next(ctx)
	})
	
	assert.Len(t, group.middlewares, 2)
}

func TestAddRoutersGroup(t *testing.T) {
	s := NewServer()
	test := s.Group("/test")

	test.Get("/path", func(ctx *Context) {})

	assert.Len(t, s.router.Routes, 1)
	assert.Equal(t, "GET", s.router.Routes[0].Method)
	assert.Equal(t, "/test/path", s.router.Routes[0].Pattern)
	assert.NotNil(t, s.router.Routes[0].Handler)

	
	test.Post("/path", func(ctx *Context) {})

	assert.Len(t, s.router.Routes, 2)
	assert.Equal(t, "POST", s.router.Routes[1].Method)
	assert.Equal(t, "/test/path", s.router.Routes[1].Pattern)
	assert.NotNil(t, s.router.Routes[1].Handler)


	test.Delete("/path", func(ctx *Context) {})

	assert.Len(t, s.router.Routes, 3)
	assert.Equal(t, "DELETE", s.router.Routes[2].Method)
	assert.Equal(t, "/test/path", s.router.Routes[2].Pattern)
	assert.NotNil(t, s.router.Routes[2].Handler)


	test.Patch("/path", func(ctx *Context) {})
	
	assert.Len(t, s.router.Routes, 4)
	assert.Equal(t, "PATCH", s.router.Routes[3].Method)
	assert.Equal(t, "/test/path", s.router.Routes[3].Pattern)
	assert.NotNil(t, s.router.Routes[3].Handler)

		
	test.Put("/path", func(ctx *Context) {})

	assert.Len(t, s.router.Routes, 5)
	assert.Equal(t, "PUT", s.router.Routes[4].Method)
	assert.Equal(t, "/test/path", s.router.Routes[4].Pattern)
	assert.NotNil(t, s.router.Routes[4].Handler)
}