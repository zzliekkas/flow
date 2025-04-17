package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gin-gonic/gin"

	"github.com/zzliekkas/flow"
)

// TestCase 定义测试用例结构
type TestCase struct {
	Name           string
	Method         string
	Path           string
	Headers        map[string]string
	Body           interface{}
	Query          url.Values
	ExpectedStatus int
	ExpectedBody   interface{}
	Setup          func(t *testing.T)
	Teardown       func(t *testing.T)
	Handler        flow.HandlerFunc
	Middleware     []flow.HandlerFunc
}

// TestSuite 定义测试套件结构
type TestSuite struct {
	Engine      *flow.Engine
	BaseURL     string
	DefaultPath string
	T           *testing.T
}

// NewTestSuite 创建一个新的测试套件
func NewTestSuite(t *testing.T) *TestSuite {
	// 设置测试模式
	gin.SetMode(gin.TestMode)
	engine := flow.New()

	return &TestSuite{
		Engine:      engine,
		BaseURL:     "",
		DefaultPath: "/",
		T:           t,
	}
}

// Request 执行HTTP请求
func (ts *TestSuite) Request(method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var reqBody io.Reader

	if body != nil {
		switch b := body.(type) {
		case string:
			reqBody = strings.NewReader(b)
		case []byte:
			reqBody = bytes.NewReader(b)
		case url.Values:
			reqBody = strings.NewReader(b.Encode())
		default:
			jsonBytes, err := json.Marshal(body)
			require.NoError(ts.T, err, "无法序列化请求体")
			reqBody = bytes.NewReader(jsonBytes)
		}
	}

	req := httptest.NewRequest(method, ts.BaseURL+path, reqBody)

	// 设置默认Content-Type
	if headers == nil || headers["Content-Type"] == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// 设置自定义头部
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	ts.Engine.ServeHTTP(w, req)

	return w
}

// AssertStatus 断言HTTP状态码
func (ts *TestSuite) AssertStatus(w *httptest.ResponseRecorder, status int) {
	assert.Equal(ts.T, status, w.Code, "HTTP状态码不匹配")
}

// AssertJSON 断言JSON响应
func (ts *TestSuite) AssertJSON(w *httptest.ResponseRecorder, expected interface{}) {
	var actual interface{}
	err := json.Unmarshal(w.Body.Bytes(), &actual)

	require.NoError(ts.T, err, "响应不是有效的JSON")
	assert.Equal(ts.T, expected, actual, "响应内容不匹配")
}

// AssertJSONContains 断言JSON响应中包含指定字段
func (ts *TestSuite) AssertJSONContains(w *httptest.ResponseRecorder, key string, value interface{}) {
	var actual map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &actual)

	require.NoError(ts.T, err, "响应不是有效的JSON")
	assert.Contains(ts.T, actual, key, "响应中不包含指定键")
	assert.Equal(ts.T, value, actual[key], "响应中指定键的值不匹配")
}

// AssertHeaderContains 断言Header包含指定值
func (ts *TestSuite) AssertHeaderContains(w *httptest.ResponseRecorder, key, value string) {
	assert.Contains(ts.T, w.Header().Get(key), value, "头部不包含指定值")
}

// AssertHeader 断言Header等于指定值
func (ts *TestSuite) AssertHeader(w *httptest.ResponseRecorder, key, value string) {
	assert.Equal(ts.T, value, w.Header().Get(key), "头部值不匹配")
}

// RunTestCase 运行单个测试用例
func (ts *TestSuite) RunTestCase(tc TestCase) {
	if tc.Setup != nil {
		tc.Setup(ts.T)
	}

	defer func() {
		if tc.Teardown != nil {
			tc.Teardown(ts.T)
		}
	}()

	path := tc.Path
	if path == "" {
		path = ts.DefaultPath
	}

	if tc.Query != nil {
		path = fmt.Sprintf("%s?%s", path, tc.Query.Encode())
	}

	w := ts.Request(tc.Method, path, tc.Body, tc.Headers)

	if tc.ExpectedStatus > 0 {
		ts.AssertStatus(w, tc.ExpectedStatus)
	}

	if tc.ExpectedBody != nil {
		ts.AssertJSON(w, tc.ExpectedBody)
	}
}

// RunTestCases 运行多个测试用例
func (ts *TestSuite) RunTestCases(cases []TestCase) {
	for _, tc := range cases {
		ts.T.Run(tc.Name, func(t *testing.T) {
			ts.RunTestCase(tc)
		})
	}
}

// UploadFile 上传文件测试辅助函数
func (ts *TestSuite) UploadFile(path, fieldName, filePath string, additionalFields map[string]string) *httptest.ResponseRecorder {
	// 打开文件
	file, err := os.Open(filePath)
	require.NoError(ts.T, err, "无法打开文件")
	defer file.Close()

	// 创建multipart表单
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
	require.NoError(ts.T, err, "无法创建表单文件")

	// 复制文件内容
	_, err = io.Copy(part, file)
	require.NoError(ts.T, err, "无法写入文件内容")

	// 添加其他表单字段
	for key, value := range additionalFields {
		err = writer.WriteField(key, value)
		require.NoError(ts.T, err, "无法添加表单字段")
	}

	err = writer.Close()
	require.NoError(ts.T, err, "无法关闭multipart写入器")

	// 创建请求
	req := httptest.NewRequest("POST", ts.BaseURL+path, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	ts.Engine.ServeHTTP(w, req)

	return w
}
