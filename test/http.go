// Package test 提供测试支持工具和辅助函数
package test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zzliekkas/flow"
)

// HTTPClient 是用于测试HTTP请求的客户端
type HTTPClient struct {
	Engine  *flow.Engine
	T       *testing.T
	BaseURL string
	Headers map[string]string
}

// NewHTTPClient 创建一个新的HTTP测试客户端
func NewHTTPClient(t *testing.T, engine *flow.Engine) *HTTPClient {
	return &HTTPClient{
		Engine:  engine,
		T:       t,
		BaseURL: "",
		Headers: make(map[string]string),
	}
}

// WithBaseURL 设置基础URL
func (c *HTTPClient) WithBaseURL(baseURL string) *HTTPClient {
	c.BaseURL = baseURL
	return c
}

// WithHeader 添加请求头
func (c *HTTPClient) WithHeader(key, value string) *HTTPClient {
	c.Headers[key] = value
	return c
}

// WithHeaders 批量添加请求头
func (c *HTTPClient) WithHeaders(headers map[string]string) *HTTPClient {
	for k, v := range headers {
		c.Headers[k] = v
	}
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
			reqBody = strings.NewReader(b)
		case []byte:
			reqBody = bytes.NewReader(b)
		case url.Values:
			reqBody = strings.NewReader(b.Encode())
		default:
			jsonBytes, err := json.Marshal(req.Body)
			require.NoError(c.T, err, "无法序列化请求体")
			reqBody = bytes.NewReader(jsonBytes)
		}
	}

	path := req.Path
	if req.Query != nil && len(req.Query) > 0 {
		path = path + "?" + req.Query.Encode()
	}

	httpReq := httptest.NewRequest(req.Method, c.BaseURL+path, reqBody)

	// 设置默认Content-Type
	if _, hasContentType := req.Headers["Content-Type"]; !hasContentType {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// 设置自定义头部
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// 设置通用头部
	for key, value := range c.Headers {
		if _, exists := req.Headers[key]; !exists {
			httpReq.Header.Set(key, value)
		}
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
		Method:  http.MethodGet,
		Path:    path,
		Headers: headers,
		Query:   query,
	})
}

// POST 发送POST请求
func (c *HTTPClient) POST(path string, body interface{}, headers map[string]string) *Response {
	return c.Do(Request{
		Method:  http.MethodPost,
		Path:    path,
		Headers: headers,
		Body:    body,
	})
}

// PUT 发送PUT请求
func (c *HTTPClient) PUT(path string, body interface{}, headers map[string]string) *Response {
	return c.Do(Request{
		Method:  http.MethodPut,
		Path:    path,
		Headers: headers,
		Body:    body,
	})
}

// DELETE 发送DELETE请求
func (c *HTTPClient) DELETE(path string, headers map[string]string) *Response {
	return c.Do(Request{
		Method:  http.MethodDelete,
		Path:    path,
		Headers: headers,
	})
}

// PATCH 发送PATCH请求
func (c *HTTPClient) PATCH(path string, body interface{}, headers map[string]string) *Response {
	return c.Do(Request{
		Method:  http.MethodPatch,
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

// AssertStatus 断言HTTP状态码
func (r *Response) AssertStatus(t *testing.T, status int) *Response {
	assert.Equal(t, status, r.StatusCode, "HTTP状态码不匹配")
	return r
}

// AssertJSON 断言JSON响应
func (r *Response) AssertJSON(t *testing.T, expected interface{}) *Response {
	var actual interface{}
	err := json.Unmarshal(r.Body, &actual)

	require.NoError(t, err, "响应不是有效的JSON")
	assert.Equal(t, expected, actual, "响应内容不匹配")
	return r
}

// AssertJSONContains 断言JSON响应中包含指定字段
func (r *Response) AssertJSONContains(t *testing.T, key string, value interface{}) *Response {
	var actual map[string]interface{}
	err := json.Unmarshal(r.Body, &actual)

	require.NoError(t, err, "响应不是有效的JSON")
	assert.Contains(t, actual, key, "响应中不包含指定键")
	assert.Equal(t, value, actual[key], "响应中指定键的值不匹配")
	return r
}

// AssertHeaderContains 断言Header包含指定值
func (r *Response) AssertHeaderContains(t *testing.T, key, value string) *Response {
	assert.Contains(t, r.Header.Get(key), value, "头部不包含指定值")
	return r
}

// AssertHeader 断言Header等于指定值
func (r *Response) AssertHeader(t *testing.T, key, value string) *Response {
	assert.Equal(t, value, r.Header.Get(key), "头部值不匹配")
	return r
}

// UploadFile 上传文件测试辅助函数
func (c *HTTPClient) UploadFile(path, fieldName, filePath string, additionalFields map[string]string) *Response {
	// 打开文件
	file, err := os.Open(filePath)
	require.NoError(c.T, err, "无法打开文件")
	defer file.Close()

	// 创建multipart表单
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
	require.NoError(c.T, err, "无法创建表单文件")

	// 复制文件内容
	_, err = io.Copy(part, file)
	require.NoError(c.T, err, "无法写入文件内容")

	// 添加其他表单字段
	for key, value := range additionalFields {
		err = writer.WriteField(key, value)
		require.NoError(c.T, err, "无法添加表单字段")
	}

	err = writer.Close()
	require.NoError(c.T, err, "无法关闭multipart写入器")

	// 创建并发送请求
	headers := map[string]string{
		"Content-Type": writer.FormDataContentType(),
	}

	return c.Do(Request{
		Method:  http.MethodPost,
		Path:    path,
		Headers: headers,
		Body:    body.Bytes(),
	})
}
