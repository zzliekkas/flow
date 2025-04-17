package testing

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zzliekkas/flow"
)

// HTTPClient 是测试HTTP请求的客户端
type HTTPClient struct {
	Engine  *flow.Engine
	T       *testing.T
	BaseURL string
}

// NewHTTPClient 创建一个新的HTTP测试客户端
func NewHTTPClient(t *testing.T, engine *flow.Engine) *HTTPClient {
	return &HTTPClient{
		Engine:  engine,
		T:       t,
		BaseURL: "",
	}
}

// WithBaseURL 设置基础URL
func (c *HTTPClient) WithBaseURL(baseURL string) *HTTPClient {
	c.BaseURL = baseURL
	return c
}

// Request 定义HTTP请求结构
type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    interface{}
	Query   url.Values
}

// Response 定义HTTP响应结构
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	Result     interface{}
	Raw        *httptest.ResponseRecorder
}

// Do 执行HTTP请求并返回响应
func (c *HTTPClient) Do(req Request) *Response {
	var reqBody io.Reader

	if req.Body != nil {
		switch b := req.Body.(type) {
		case string:
			reqBody = bytes.NewReader([]byte(b))
		case []byte:
			reqBody = bytes.NewReader(b)
		case url.Values:
			reqBody = bytes.NewReader([]byte(b.Encode()))
		default:
			jsonBytes, err := json.Marshal(req.Body)
			require.NoError(c.T, err, "无法序列化请求体")
			reqBody = bytes.NewReader(jsonBytes)
		}
	}

	path := req.Path
	if req.Query != nil {
		if len(req.Query) > 0 {
			path = path + "?" + req.Query.Encode()
		}
	}

	httpReq := httptest.NewRequest(req.Method, c.BaseURL+path, reqBody)

	// 设置默认Content-Type
	if req.Headers == nil || req.Headers["Content-Type"] == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// 设置自定义头部
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	c.Engine.ServeHTTP(w, httpReq)

	return &Response{
		StatusCode: w.Code,
		Header:     w.Header(),
		Body:       w.Body.Bytes(),
		Raw:        w,
	}
}

// GET 发送GET请求
func (c *HTTPClient) GET(path string, query url.Values, headers map[string]string) *Response {
	return c.Do(Request{
		Method:  "GET",
		Path:    path,
		Headers: headers,
		Query:   query,
	})
}

// POST 发送POST请求
func (c *HTTPClient) POST(path string, body interface{}, headers map[string]string) *Response {
	return c.Do(Request{
		Method:  "POST",
		Path:    path,
		Headers: headers,
		Body:    body,
	})
}

// PUT 发送PUT请求
func (c *HTTPClient) PUT(path string, body interface{}, headers map[string]string) *Response {
	return c.Do(Request{
		Method:  "PUT",
		Path:    path,
		Headers: headers,
		Body:    body,
	})
}

// DELETE 发送DELETE请求
func (c *HTTPClient) DELETE(path string, headers map[string]string) *Response {
	return c.Do(Request{
		Method:  "DELETE",
		Path:    path,
		Headers: headers,
	})
}

// PATCH 发送PATCH请求
func (c *HTTPClient) PATCH(path string, body interface{}, headers map[string]string) *Response {
	return c.Do(Request{
		Method:  "PATCH",
		Path:    path,
		Headers: headers,
		Body:    body,
	})
}

// BindJSON 将响应体绑定到JSON结构
func (r *Response) BindJSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// JSON 返回响应体解析为JSON的结果
func (r *Response) JSON() (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal(r.Body, &result)
	return result, err
}

// String 返回响应体的字符串表示
func (r *Response) String() string {
	return string(r.Body)
}
