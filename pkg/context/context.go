package context

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
)

const (
	UserBindKey = "bind_user"
)

// Context 封装请求上下文
type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
	store   map[string]interface{} // 用于存储上下文数据
}

// New 创建新的上下文
func New(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer:  w,
		Request: r,
		store:   make(map[string]interface{}),
	}
}

// Set 在上下文中存储值
func (c *Context) Set(key string, value interface{}) {
	c.store[key] = value
}

// Get 从上下文中获取值
func (c *Context) Get(key string) (interface{}, bool) {
	value, exists := c.store[key]
	return value, exists
}

// GetString 获取字符串值
func (c *Context) GetString(key string) string {
	if val, exists := c.store[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// BindUser 绑定用户信息
func (c *Context) BindUser(val interface{}) {
	c.Set(UserBindKey, val)
}

// GetBindUser 获取绑定的用户信息
func (c *Context) GetBindUser(recipient interface{}) error {
	val, exists := c.Get(UserBindKey)
	if !exists {
		return fmt.Errorf("not found bind user")
	}
	if reflect.ValueOf(recipient).Kind() != reflect.Ptr {
		return fmt.Errorf("recipient must be a pointer")
	}

	// 使用 json 序列化和反序列化进行深度拷贝
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, recipient)
}

// find 按优先级查找参数值
func (c *Context) find(key string) string {
	// 1. 查找URL路径参数
	// 2. 查找Query参数
	if val := c.Request.URL.Query().Get(key); val != "" {
		return val
	}
	// 3. 查找Header
	if val := c.Request.Header.Get(key); val != "" {
		return val
	}
	// 4. 查找POST表单
	if err := c.Request.ParseForm(); err == nil {
		if val := c.Request.PostForm.Get(key); val != "" {
			return val
		}
	}
	// 5. 查找存储的值
	if val := c.GetString(key); val != "" {
		return val
	}
	return ""
}

// FindString 查找字符串参数
func (c *Context) FindString(key string) string {
	return c.find(key)
}

// FindDefaultString 查找字符串参数,支持默认值
func (c *Context) FindDefaultString(key string, defaultValue string) string {
	val := c.find(key)
	if val != "" {
		return val
	}
	return defaultValue
}

// FindInt 查找整数参数
func (c *Context) FindInt(key string) int {
	val := c.find(key)
	if val == "" {
		return 0
	}
	i, _ := strconv.Atoi(val)
	return i
}

// FindDefaultInt 查找整数参数,支持默认值
func (c *Context) FindDefaultInt(key string, defaultValue int) int {
	val := c.FindInt(key)
	if val != 0 {
		return val
	}
	return defaultValue
}

// FindBool 查找布尔参数
func (c *Context) FindBool(key string) bool {
	val := c.find(key)
	if val == "" {
		return false
	}
	b, _ := strconv.ParseBool(val)
	return b
}

// FindDefaultBool 查找布尔参数,支持默认值
func (c *Context) FindDefaultBool(key string, defaultValue bool) bool {
	val := c.find(key)
	if val != "" {
		b, _ := strconv.ParseBool(val)
		return b
	}
	return defaultValue
}
