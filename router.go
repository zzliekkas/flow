package flow

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RouterGroup 是Flow的路由组结构
type RouterGroup struct {
	RouterGroup gin.RouterGroup
	engine      *Engine
}

// wrapHandlers 将Flow的HandlerFunc切片转换为gin的HandlerFunc切片
func wrapHandlers(engine *Engine, handlers []HandlerFunc) []gin.HandlerFunc {
	ginHandlers := make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handler := handler // 创建局部变量避免闭包问题
		ginHandlers[i] = func(c *gin.Context) {
			flowContext := engine.NewContext(c)
			handler(flowContext)
		}
	}
	return ginHandlers
}

// NewContext 创建Flow上下文
func (e *Engine) NewContext(c *gin.Context) *Context {
	return &Context{
		Context: c,
		engine:  e,
	}
}

// Handle 注册处理函数到给定的HTTP方法和路径
func (e *Engine) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) {
	e.Engine.Handle(httpMethod, relativePath, wrapHandlers(e, handlers)...)
}

// GET 是对Handle("GET", path, handlers)的简便方法
func (e *Engine) GET(relativePath string, handlers ...HandlerFunc) {
	e.Handle(http.MethodGet, relativePath, handlers...)
}

// POST 是对Handle("POST", path, handlers)的简便方法
func (e *Engine) POST(relativePath string, handlers ...HandlerFunc) {
	e.Handle(http.MethodPost, relativePath, handlers...)
}

// PUT 是对Handle("PUT", path, handlers)的简便方法
func (e *Engine) PUT(relativePath string, handlers ...HandlerFunc) {
	e.Handle(http.MethodPut, relativePath, handlers...)
}

// DELETE 是对Handle("DELETE", path, handlers)的简便方法
func (e *Engine) DELETE(relativePath string, handlers ...HandlerFunc) {
	e.Handle(http.MethodDelete, relativePath, handlers...)
}

// PATCH 是对Handle("PATCH", path, handlers)的简便方法
func (e *Engine) PATCH(relativePath string, handlers ...HandlerFunc) {
	e.Handle(http.MethodPatch, relativePath, handlers...)
}

// Group 创建一个新的路由组
func (e *Engine) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	ginGroup := e.Engine.Group(relativePath, wrapHandlers(e, handlers)...)
	return &RouterGroup{
		RouterGroup: *ginGroup,
		engine:      e,
	}
}

// Use 添加全局中间件
func (e *Engine) Use(middleware ...HandlerFunc) *Engine {
	e.Engine.Use(wrapHandlers(e, middleware)...)
	return e
}

// Handle 在路由组中注册处理函数
func (g *RouterGroup) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) {
	g.RouterGroup.Handle(httpMethod, relativePath, wrapHandlers(g.engine, handlers)...)
}

// GET 是对Handle("GET", path, handlers)的简便方法
func (g *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) {
	g.Handle(http.MethodGet, relativePath, handlers...)
}

// POST 是对Handle("POST", path, handlers)的简便方法
func (g *RouterGroup) POST(relativePath string, handlers ...HandlerFunc) {
	g.Handle(http.MethodPost, relativePath, handlers...)
}

// PUT 是对Handle("PUT", path, handlers)的简便方法
func (g *RouterGroup) PUT(relativePath string, handlers ...HandlerFunc) {
	g.Handle(http.MethodPut, relativePath, handlers...)
}

// DELETE 是对Handle("DELETE", path, handlers)的简便方法
func (g *RouterGroup) DELETE(relativePath string, handlers ...HandlerFunc) {
	g.Handle(http.MethodDelete, relativePath, handlers...)
}

// PATCH 是对Handle("PATCH", path, handlers)的简便方法
func (g *RouterGroup) PATCH(relativePath string, handlers ...HandlerFunc) {
	g.Handle(http.MethodPatch, relativePath, handlers...)
}

// Group 创建一个子路由组
func (g *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	ginGroup := g.RouterGroup.Group(relativePath, wrapHandlers(g.engine, handlers)...)
	return &RouterGroup{
		RouterGroup: *ginGroup,
		engine:      g.engine,
	}
}

// Use 添加路由组中间件
func (g *RouterGroup) Use(middleware ...HandlerFunc) *RouterGroup {
	g.RouterGroup.Use(wrapHandlers(g.engine, middleware)...)
	return g
}
