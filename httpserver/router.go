package httpserver

import (
	"strings"
)

type Route struct {
	Method  string
	Pattern string
	Handler HandlerFunc
}

type Router struct {
	Routes []Route
}


type HandlerFunc func(ctx *Context)

func (r *Router) Add(method string, pattern string, handler HandlerFunc) {
	route := Route{
		Method:  method,
		Pattern: pattern,
		Handler: handler,
	}

	r.Routes = append(r.Routes, route)
}

func (r *Router) Get(pattern string, handler HandlerFunc) {
	r.Add("GET", pattern, handler)
}

func (r *Router) Post(pattern string, handler HandlerFunc) {
	r.Add("POST", pattern, handler)
}

func (r *Router) Delete(pattern string, handler HandlerFunc) {
	r.Add("DELETE", pattern, handler)
}

func (r *Router) Patch(pattern string, handler HandlerFunc) {
	r.Add("PATCH", pattern, handler)
}

func (r *Router) Put(pattern string, handler HandlerFunc) {
	r.Add("PUT", pattern, handler)
}


func (r *Router) Match(method string, path string) (HandlerFunc, map[string]string) {
	for _, route := range r.Routes {
		if route.Method != method {
			continue
		}

		patternSegm := strings.Split(route.Pattern, "/")

		pathWithoutQuery, _, _ := strings.Cut(path, "?") 
		pathSegm := strings.Split(pathWithoutQuery, "/")

		if len(pathSegm) != len(patternSegm) {
			continue
		}

		params := make(map[string]string)

		matched := true
		
		for i := range len(patternSegm) {
			if strings.HasPrefix(patternSegm[i], ":") {
				params[patternSegm[i][1:]] = pathSegm[i] 
			} else {
				if pathSegm[i] != patternSegm[i] {
					matched = false
					break
				}
			}
		}
		
		if matched {
			return route.Handler, params
		}

	}

	return nil, nil
}