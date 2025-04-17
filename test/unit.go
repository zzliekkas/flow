// Package test 提供测试支持工具和辅助函数
package test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// UnitTest 提供单元测试支持
type UnitTest struct {
	T *testing.T
}

// NewUnitTest 创建一个新的单元测试助手
func NewUnitTest(t *testing.T) *UnitTest {
	return &UnitTest{T: t}
}

// GetFunctionName 获取函数的名称（用于测试函数名称）
func (u *UnitTest) GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// GetTestName 从测试函数获取测试名称
func (u *UnitTest) GetTestName() string {
	pc, _, _, _ := runtime.Caller(1)
	fullName := runtime.FuncForPC(pc).Name()
	parts := strings.Split(fullName, ".")
	return parts[len(parts)-1]
}

// ToJSON 将对象转换为JSON字符串
func (u *UnitTest) ToJSON(obj interface{}) string {
	bytes, err := json.Marshal(obj)
	require.NoError(u.T, err, "无法转换对象为JSON")
	return string(bytes)
}

// FromJSON 将JSON字符串转换为对象
func (u *UnitTest) FromJSON(jsonStr string, obj interface{}) {
	err := json.Unmarshal([]byte(jsonStr), obj)
	require.NoError(u.T, err, "无法解析JSON字符串")
}

// AssertJSONEqual 断言两个JSON字符串相等
func (u *UnitTest) AssertJSONEqual(expected, actual string) {
	var expectedObj, actualObj interface{}

	err := json.Unmarshal([]byte(expected), &expectedObj)
	require.NoError(u.T, err, "无法解析预期的JSON字符串")

	err = json.Unmarshal([]byte(actual), &actualObj)
	require.NoError(u.T, err, "无法解析实际的JSON字符串")

	assert.Equal(u.T, expectedObj, actualObj, "JSON内容不相等")
}

// CallMethod 调用对象的方法并返回结果
func (u *UnitTest) CallMethod(obj interface{}, methodName string, args ...interface{}) []interface{} {
	objValue := reflect.ValueOf(obj)
	method := objValue.MethodByName(methodName)

	if !method.IsValid() {
		u.T.Fatalf("方法 %s 不存在于对象上", methodName)
	}

	methodType := method.Type()
	numIn := methodType.NumIn()
	numArgs := len(args)

	if numIn != numArgs {
		u.T.Fatalf("方法 %s 需要 %d 个参数，但提供了 %d 个", methodName, numIn, numArgs)
	}

	in := make([]reflect.Value, numArgs)
	for i := 0; i < numArgs; i++ {
		argValue := reflect.ValueOf(args[i])
		paramType := methodType.In(i)

		if !argValue.Type().AssignableTo(paramType) {
			u.T.Fatalf("参数 #%d 类型不匹配: 需要 %v，提供了 %v", i, paramType, argValue.Type())
		}

		in[i] = argValue
	}

	out := method.Call(in)
	result := make([]interface{}, len(out))

	for i, v := range out {
		result[i] = v.Interface()
	}

	return result
}

// RunTestCase 运行单个测试用例
func (u *UnitTest) RunTestCase(name string, fn func(t *testing.T)) {
	u.T.Run(name, fn)
}

// RunTestCases 运行多个测试用例
func (u *UnitTest) RunTestCases(testCases map[string]func(t *testing.T)) {
	for name, fn := range testCases {
		u.RunTestCase(name, fn)
	}
}

// Mock 是单元测试中的模拟对象基类
type Mock struct {
	Calls map[string][][]interface{}
}

// NewMock 创建一个新的模拟对象
func NewMock() *Mock {
	return &Mock{
		Calls: make(map[string][][]interface{}),
	}
}

// RecordCall 记录对模拟对象的方法调用
func (m *Mock) RecordCall(methodName string, args ...interface{}) {
	if m.Calls[methodName] == nil {
		m.Calls[methodName] = make([][]interface{}, 0)
	}

	m.Calls[methodName] = append(m.Calls[methodName], args)
}

// GetCalls 获取对特定方法的所有调用
func (m *Mock) GetCalls(methodName string) [][]interface{} {
	return m.Calls[methodName]
}

// CallCount 获取特定方法被调用的次数
func (m *Mock) CallCount(methodName string) int {
	calls := m.GetCalls(methodName)
	if calls == nil {
		return 0
	}
	return len(calls)
}

// WasCalled 判断特定方法是否被调用过
func (m *Mock) WasCalled(methodName string) bool {
	return m.CallCount(methodName) > 0
}

// MockAssertions 提供对模拟对象的断言
type MockAssertions struct {
	T    *testing.T
	Mock *Mock
}

// NewMockAssertions 创建新的模拟断言助手
func NewMockAssertions(t *testing.T, mock *Mock) *MockAssertions {
	return &MockAssertions{
		T:    t,
		Mock: mock,
	}
}

// AssertCalled 断言方法已被调用
func (a *MockAssertions) AssertCalled(methodName string, args ...interface{}) bool {
	calls := a.Mock.GetCalls(methodName)

	if len(calls) == 0 {
		assert.Fail(a.T, fmt.Sprintf("预期方法 '%s' 被调用，但它没有被调用", methodName))
		return false
	}

	if len(args) == 0 {
		return true
	}

	for _, call := range calls {
		if len(call) != len(args) {
			continue
		}

		match := true
		for i, arg := range args {
			if !reflect.DeepEqual(arg, call[i]) {
				match = false
				break
			}
		}

		if match {
			return true
		}
	}

	assert.Fail(a.T, fmt.Sprintf("预期方法 '%s' 使用参数 %v 被调用，但未找到匹配的调用", methodName, args))
	return false
}

// AssertNotCalled 断言方法未被调用
func (a *MockAssertions) AssertNotCalled(methodName string) bool {
	calls := a.Mock.GetCalls(methodName)

	if len(calls) > 0 {
		assert.Fail(a.T, fmt.Sprintf("预期方法 '%s' 不被调用，但它被调用了 %d 次", methodName, len(calls)))
		return false
	}

	return true
}

// AssertCalledTimes 断言方法被调用特定次数
func (a *MockAssertions) AssertCalledTimes(methodName string, times int) bool {
	callCount := a.Mock.CallCount(methodName)

	if callCount != times {
		assert.Fail(a.T, fmt.Sprintf("预期方法 '%s' 被调用 %d 次，但它被调用了 %d 次", methodName, times, callCount))
		return false
	}

	return true
}
