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
}

func(g *Group) Get(pattern string, handler HandlerFunc) {
	g.routers.Get(g.prefix + pattern, handler)
}

func(g *Group) Post(pattern string, handler HandlerFunc) {
	g.routers.Post(g.prefix + pattern, handler)
}

func(g *Group) Patch(pattern string, handler HandlerFunc) {
	g.routers.Patch(g.prefix + pattern, handler)
}

func(g *Group) Put(pattern string, handler HandlerFunc) {
	g.routers.Put(g.prefix + pattern, handler)
}
func(g *Group) Delete(pattern string, handler HandlerFunc) {
	g.routers.Delete(g.prefix + pattern, handler)
}

func(g *Group) Group(prefix string) *Group {
	return &Group{
		routers: g.routers,
		prefix: g.prefix + prefix,
	}
}