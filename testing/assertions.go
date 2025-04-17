package testing

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertStatus 断言HTTP状态码
func AssertStatus(t *testing.T, response *Response, expectedCode int) {
	assert.Equal(t, expectedCode, response.StatusCode, "HTTP状态码不匹配，期望：%d，实际：%d", expectedCode, response.StatusCode)
}

// AssertOK 断言HTTP状态码为200 OK
func AssertOK(t *testing.T, response *Response) {
	AssertStatus(t, response, http.StatusOK)
}

// AssertCreated 断言HTTP状态码为201 Created
func AssertCreated(t *testing.T, response *Response) {
	AssertStatus(t, response, http.StatusCreated)
}

// AssertNoContent 断言HTTP状态码为204 No Content
func AssertNoContent(t *testing.T, response *Response) {
	AssertStatus(t, response, http.StatusNoContent)
}

// AssertBadRequest 断言HTTP状态码为400 Bad Request
func AssertBadRequest(t *testing.T, response *Response) {
	AssertStatus(t, response, http.StatusBadRequest)
}

// AssertUnauthorized 断言HTTP状态码为401 Unauthorized
func AssertUnauthorized(t *testing.T, response *Response) {
	AssertStatus(t, response, http.StatusUnauthorized)
}

// AssertForbidden 断言HTTP状态码为403 Forbidden
func AssertForbidden(t *testing.T, response *Response) {
	AssertStatus(t, response, http.StatusForbidden)
}

// AssertNotFound 断言HTTP状态码为404 Not Found
func AssertNotFound(t *testing.T, response *Response) {
	AssertStatus(t, response, http.StatusNotFound)
}

// AssertHeader 断言HTTP头部值
func AssertHeader(t *testing.T, response *Response, key, expectedValue string) {
	assert.Equal(t, expectedValue, response.Header.Get(key), "HTTP头部值不匹配，键：%s，期望：%s，实际：%s", key, expectedValue, response.Header.Get(key))
}

// AssertJSON 断言JSON响应体
func AssertJSON(t *testing.T, response *Response, expectedJSON interface{}) {
	var expected, actual interface{}

	// 处理期望值
	switch v := expectedJSON.(type) {
	case string:
		require.NoError(t, json.Unmarshal([]byte(v), &expected), "无法解析期望的JSON字符串")
	case []byte:
		require.NoError(t, json.Unmarshal(v, &expected), "无法解析期望的JSON字节")
	default:
		expected = expectedJSON
	}

	// 解析实际响应
	require.NoError(t, json.Unmarshal(response.Body, &actual), "无法解析响应体")

	// 执行比较
	assert.Equal(t, expected, actual, "JSON响应不匹配")
}

// AssertJSONPath 断言JSON响应体中特定路径的值
func AssertJSONPath(t *testing.T, response *Response, path string, expectedValue interface{}) {
	var jsonData map[string]interface{}
	require.NoError(t, json.Unmarshal(response.Body, &jsonData), "无法解析响应体")

	// 简单路径查找 (只支持一级路径)
	value, exists := jsonData[path]
	assert.True(t, exists, "JSON路径 '%s' 不存在", path)
	assert.Equal(t, expectedValue, value, "JSON路径 '%s' 的值不匹配", path)
}

// AssertBodyContains 断言响应体包含特定字符串
func AssertBodyContains(t *testing.T, response *Response, substring string) {
	assert.Contains(t, string(response.Body), substring, "响应体不包含期望的子字符串")
}

// AssertBodyEquals 断言响应体等于特定字符串
func AssertBodyEquals(t *testing.T, response *Response, expected string) {
	assert.Equal(t, expected, string(response.Body), "响应体不匹配")
}

// AssertJSONSchema 断言JSON响应体符合指定的JSON Schema(简化版，仅检查必填字段)
func AssertJSONSchema(t *testing.T, response *Response, requiredFields []string) {
	var jsonData map[string]interface{}
	require.NoError(t, json.Unmarshal(response.Body, &jsonData), "无法解析响应体")

	for _, field := range requiredFields {
		_, exists := jsonData[field]
		assert.True(t, exists, "必填字段 '%s' 不存在", field)
	}
}
