package router

import (
	"cakeduel-backend/internal/app"
	"net/http"
)

// groupType 路由组
type groupType struct {
	Prefix string
	Router *Router
}

// Router 路由
type Router struct {
	Routes map[string]map[string]http.Handler
	App    *app.App
}

// group 创建路由组
func (r *Router) group(prefix string) *groupType {
	return &groupType{
		Prefix: prefix,
		Router: r,
	}
}

// handle 注册路由
func (r *Router) handle(method, path string, handler http.Handler) {
	if r.Routes[path] == nil {
		r.Routes[path] = make(map[string]http.Handler)
	}
	r.Routes[path][method] = handler
}

// handleFunc 注册路由处理函数
func (r *Router) handleFunc(method, path string, handler http.HandlerFunc) {
	r.handle(method, path, handler)
}

// ServeHTTP 路由服务
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	method := req.Method
	// 路由匹配
	methods, pathExists := r.Routes[path]
	if pathExists {
		handler, methodExists := methods[method]
		if methodExists {
			handler.ServeHTTP(w, req)
			return
		}
	}
	// 未精确匹配时回退到根处理器(SPA/静态资源)
	if root, ok := r.Routes["/"]; ok {
		if handler, ok2 := root[req.Method]; ok2 {
			handler.ServeHTTP(w, req)
			return
		}
	}
	http.NotFound(w, req)
}

// groupHandleFunc 路由组注册处理函数
func (g *groupType) handleFunc(method, path string, handler http.HandlerFunc) {
	g.Router.handleFunc(method, g.Prefix+path, handler)
}
