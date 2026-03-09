package httpserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatch(t *testing.T) {
	r := &Router{}
	r.Get("/api/user/:id", func(ctx *Context) {})

	method := "GET"
	path := "/api/user/21"

	handler, params := r.Match(method, path)

	assert.Equal(t, "21", params["id"])
	assert.NotNil(t, handler)
}

func TestNotMatch(t *testing.T) {
	r := &Router{}
	r.Get("/api/user/:id", func(ctx *Context) {})

	method := "GET"
	path := "/api/users/21"

	handler, params := r.Match(method, path)

	assert.Nil(t, params)
	assert.Nil(t, handler)
}

func TestMatchWithoutParams(t *testing.T) {
	r := &Router{}
	ctx := &Context{}

	var called bool

	r.Get("/api/queue", func(ctx *Context) { called = true})
	r.Post("/api/user", func(ctx *Context) {})
	r.Delete("/api/queue", func(ctx *Context) {})

	method := "GET"
	path := "/api/queue"

	handler, params := r.Match(method, path)
	handler(ctx)

	assert.True(t, called)
	assert.NotNil(t, handler)
	assert.Empty(t, params)
}

func TestAddRouters(t *testing.T) {
	r := &Router{}

	r.Get("/path", func(ctx *Context) {})

	assert.Len(t, r.Routes, 1)
	assert.Equal(t, "GET", r.Routes[0].Method)
	assert.Equal(t, "/path", r.Routes[0].Pattern)
	assert.NotNil(t, r.Routes[0].Handler)

	
	r.Post("/path", func(ctx *Context) {})

	assert.Len(t, r.Routes, 2)
	assert.Equal(t, "POST", r.Routes[1].Method)
	assert.Equal(t, "/path", r.Routes[1].Pattern)
	assert.NotNil(t, r.Routes[1].Handler)


	r.Delete("/path", func(ctx *Context) {})

	assert.Len(t, r.Routes, 3)
	assert.Equal(t, "DELETE", r.Routes[2].Method)
	assert.Equal(t, "/path", r.Routes[2].Pattern)
	assert.NotNil(t, r.Routes[2].Handler)


	r.Patch("/path", func(ctx *Context) {})
	
	assert.Len(t, r.Routes, 4)
	assert.Equal(t, "PATCH", r.Routes[3].Method)
	assert.Equal(t, "/path", r.Routes[3].Pattern)
	assert.NotNil(t, r.Routes[3].Handler)

		
	r.Put("/path", func(ctx *Context) {})

	assert.Len(t, r.Routes, 5)
	assert.Equal(t, "PUT", r.Routes[4].Method)
	assert.Equal(t, "/path", r.Routes[4].Pattern)
	assert.NotNil(t, r.Routes[4].Handler)
}