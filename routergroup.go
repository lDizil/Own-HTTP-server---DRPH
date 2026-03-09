package httpserver

type RouterGroup interface {
    Get(pattern string, handler HandlerFunc)
    Post(pattern string, handler HandlerFunc)
	Patch(pattern string, handler HandlerFunc)
    Put(pattern string, handler HandlerFunc)
	Delete(pattern string, handler HandlerFunc)
	
}

type Group struct {
	routers RouterGroup
	prefix string
	middlewares []MiddlewareFunc
}

func(g *Group) Group(prefix string) *Group {
	newMiddlewares := make([]MiddlewareFunc, len(g.middlewares))
	copy(newMiddlewares, g.middlewares)
	return &Group{
		routers: g.routers,
		prefix: g.prefix + prefix,
		middlewares: newMiddlewares,
	}
}

func (g *Group) Use(mw ...MiddlewareFunc) {
	g.middlewares = append(g.middlewares, mw...)
}

func(g *Group) buildGroupChain(handler HandlerFunc) HandlerFunc {
	n := len(g.middlewares)

	if n == 0 {
		return handler
	}

	chain := handler

	for i := n - 1; i >= 0; i-- {
		next := chain
		mw := g.middlewares[i]

		chain = func(ctx *Context) {
			mw(ctx, next)
		}
	}

	return chain
}

func(g *Group) Get(pattern string, handler HandlerFunc) {
	wrapHandler := g.buildGroupChain(handler)
	g.routers.Get(g.prefix + pattern, wrapHandler)
}

func(g *Group) Post(pattern string, handler HandlerFunc) {
	wrapHandler := g.buildGroupChain(handler)
	g.routers.Post(g.prefix + pattern, wrapHandler)
}

func(g *Group) Patch(pattern string, handler HandlerFunc) {
	wrapHandler := g.buildGroupChain(handler)
	g.routers.Patch(g.prefix + pattern, wrapHandler)
}

func(g *Group) Put(pattern string, handler HandlerFunc) {
	wrapHandler := g.buildGroupChain(handler)
	g.routers.Put(g.prefix + pattern, wrapHandler)
}
func(g *Group) Delete(pattern string, handler HandlerFunc) {
	wrapHandler := g.buildGroupChain(handler)
	g.routers.Delete(g.prefix + pattern, wrapHandler)
}
