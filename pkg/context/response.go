package context

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ResponseType 定义响应类型
type ResponseType int

const (
	// ResponseTypeJSON JSON响应类型
	ResponseTypeJSON ResponseType = iota
	// ResponseTypeXML XML响应类型
	ResponseTypeXML
	// ResponseTypeHTML HTML响应类型
	ResponseTypeHTML
	// ResponseTypeText 文本响应类型
	ResponseTypeText
)

// Response 标准响应结构
type Response struct {
	Code    int         `json:"code"`              // 状态码
	Message string      `json:"message,omitempty"` // 消息
	Data    interface{} `json:"data,omitempty"`    // 数据
}

// JSON 返回JSON响应
func (c *Context) JSON(code int, data interface{}) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(code)
	json.NewEncoder(c.Writer).Encode(data)
}

// Success 返回成功响应
func (c *Context) Success(data interface{}) {
	resp := Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	}
	c.JSON(http.StatusOK, resp)
}

// Error 返回错误响应
func (c *Context) Error(code int, message string) {
	resp := Response{
		Code:    code,
		Message: message,
	}
	c.JSON(code, resp)
}

// BadRequest 返回400错误
func (c *Context) BadRequest(message string) {
	c.Error(http.StatusBadRequest, message)
}

// Unauthorized 返回401错误
func (c *Context) Unauthorized(message string) {
	c.Error(http.StatusUnauthorized, message)
}

// Forbidden 返回403错误
func (c *Context) Forbidden(message string) {
	c.Error(http.StatusForbidden, message)
}

// NotFound 返回404错误
func (c *Context) NotFound(message string) {
	c.Error(http.StatusNotFound, message)
}

// InternalServerError 返回500错误
func (c *Context) InternalServerError(message string) {
	c.Error(http.StatusInternalServerError, message)
}

// String 返回文本响应
func (c *Context) String(code int, format string, values ...interface{}) {
	c.Writer.Header().Set("Content-Type", "text/plain")
	c.Writer.WriteHeader(code)
	if len(values) > 0 {
		format = fmt.Sprintf(format, values...)
	}
	c.Writer.Write([]byte(format))
}

// HTML 返回HTML响应
func (c *Context) HTML(code int, html string) {
	c.Writer.Header().Set("Content-Type", "text/html")
	c.Writer.WriteHeader(code)
	c.Writer.Write([]byte(html))
}

// File 返回文件下载
func (c *Context) File(filepath string) {
	http.ServeFile(c.Writer, c.Request, filepath)
}

// Redirect 重定向
func (c *Context) Redirect(code int, location string) {
	http.Redirect(c.Writer, c.Request, location, code)
}
