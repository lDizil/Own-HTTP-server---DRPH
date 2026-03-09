package httpserver

type HandlerFunc func(ctx *Context)

type MiddlewareFunc func(ctx *Context, next HandlerFunc)