// Package test 提供测试支持工具和辅助函数
package test

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

// MockGenerator 提供用于测试的数据生成功能
type MockGenerator struct {
	rand *rand.Rand
}

// NewMockGenerator 创建一个新的模拟数据生成器
func NewMockGenerator() *MockGenerator {
	source := rand.NewSource(time.Now().UnixNano())
	return &MockGenerator{
		rand: rand.New(source),
	}
}

// WithSeed 使用固定种子初始化随机数生成器，用于生成可重现的数据
func (g *MockGenerator) WithSeed(seed int64) *MockGenerator {
	source := rand.NewSource(seed)
	g.rand = rand.New(source)
	return g
}

// RandomInt 生成指定范围内的随机整数
func (g *MockGenerator) RandomInt(min, max int) int {
	return min + g.rand.Intn(max-min+1)
}

// RandomFloat 生成指定范围内的随机浮点数
func (g *MockGenerator) RandomFloat(min, max float64) float64 {
	return min + g.rand.Float64()*(max-min)
}

// RandomBool 生成随机布尔值
func (g *MockGenerator) RandomBool() bool {
	return g.rand.Intn(2) == 1
}

// RandomString 生成指定长度的随机字符串
func (g *MockGenerator) RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[g.rand.Intn(len(charset))]
	}
	return string(result)
}

// RandomEmail 生成随机邮箱
func (g *MockGenerator) RandomEmail() string {
	domains := []string{"example.com", "test.com", "mock.org", "fakemail.net"}
	username := g.RandomString(8)
	domain := domains[g.rand.Intn(len(domains))]
	return fmt.Sprintf("%s@%s", strings.ToLower(username), domain)
}

// RandomName 生成随机姓名
func (g *MockGenerator) RandomName() string {
	firstNames := []string{"张", "李", "王", "赵", "刘", "陈", "杨", "黄", "周", "吴"}
	lastNames := []string{"伟", "芳", "娜", "秀英", "敏", "静", "丽", "强", "磊", "洋"}
	firstName := firstNames[g.rand.Intn(len(firstNames))]
	lastName := lastNames[g.rand.Intn(len(lastNames))]
	return firstName + lastName
}

// RandomDate 生成指定范围内的随机日期
func (g *MockGenerator) RandomDate(startYear, endYear int) time.Time {
	min := time.Date(startYear, 1, 1, 0, 0, 0, 0, time.Local).Unix()
	max := time.Date(endYear, 12, 31, 23, 59, 59, 0, time.Local).Unix()
	delta := max - min

	sec := g.rand.Int63n(delta) + min
	return time.Unix(sec, 0)
}

// RandomPhoneNumber 生成随机手机号
func (g *MockGenerator) RandomPhoneNumber() string {
	prefixes := []string{"138", "139", "150", "151", "152", "156", "158", "159", "182", "183"}
	prefix := prefixes[g.rand.Intn(len(prefixes))]
	return prefix + fmt.Sprintf("%08d", g.rand.Intn(100000000))
}

// RandomElement 从切片中随机选择一个元素
func (g *MockGenerator) RandomElement(slice interface{}) interface{} {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		panic("RandomElement只接受切片类型")
	}

	length := v.Len()
	if length == 0 {
		return nil
	}

	return v.Index(g.rand.Intn(length)).Interface()
}

// RandomSubset 从切片中随机选择n个元素
func (g *MockGenerator) RandomSubset(slice interface{}, n int) interface{} {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		panic("RandomSubset只接受切片类型")
	}

	length := v.Len()
	if n > length {
		n = length
	}

	// 创建索引数组
	indices := make([]int, length)
	for i := range indices {
		indices[i] = i
	}

	// 洗牌算法
	for i := length - 1; i > 0; i-- {
		j := g.rand.Intn(i + 1)
		indices[i], indices[j] = indices[j], indices[i]
	}

	// 创建结果切片
	resultType := reflect.SliceOf(v.Type().Elem())
	result := reflect.MakeSlice(resultType, n, n)

	// 填充结果
	for i := 0; i < n; i++ {
		result.Index(i).Set(v.Index(indices[i]))
	}

	return result.Interface()
}

// FillStruct 使用随机数据填充结构体
func (g *MockGenerator) FillStruct(s interface{}) {
	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		panic("FillStruct只接受结构体指针")
	}

	g.fillStructValue(v.Elem())
}

// fillStructValue 填充结构体字段
func (g *MockGenerator) fillStructValue(v reflect.Value) {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		// 如果字段不可设置，跳过
		if !field.CanSet() {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			fieldName := strings.ToLower(t.Field(i).Name)
			if strings.Contains(fieldName, "email") {
				field.SetString(g.RandomEmail())
			} else if strings.Contains(fieldName, "name") {
				field.SetString(g.RandomName())
			} else if strings.Contains(fieldName, "phone") {
				field.SetString(g.RandomPhoneNumber())
			} else {
				field.SetString(g.RandomString(10))
			}

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			field.SetInt(int64(g.RandomInt(1, 1000)))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			field.SetUint(uint64(g.RandomInt(1, 1000)))

		case reflect.Float32, reflect.Float64:
			field.SetFloat(g.RandomFloat(1.0, 1000.0))

		case reflect.Bool:
			field.SetBool(g.RandomBool())

		case reflect.Struct:
			if field.Type() == reflect.TypeOf(time.Time{}) {
				field.Set(reflect.ValueOf(g.RandomDate(2000, 2023)))
			} else {
				g.fillStructValue(field)
			}

		case reflect.Ptr:
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			g.fillStructValue(field.Elem())

		case reflect.Slice:
			// 对于简单类型的切片，创建1-5个元素
			elemType := field.Type().Elem()
			length := g.RandomInt(1, 5)
			slice := reflect.MakeSlice(field.Type(), length, length)

			for j := 0; j < length; j++ {
				g.fillValue(slice.Index(j), elemType)
			}

			field.Set(slice)
		}
	}
}

// fillValue 根据类型填充值
func (g *MockGenerator) fillValue(v reflect.Value, t reflect.Type) {
	switch t.Kind() {
	case reflect.String:
		v.SetString(g.RandomString(10))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(int64(g.RandomInt(1, 1000)))

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(uint64(g.RandomInt(1, 1000)))

	case reflect.Float32, reflect.Float64:
		v.SetFloat(g.RandomFloat(1.0, 1000.0))

	case reflect.Bool:
		v.SetBool(g.RandomBool())

	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			v.Set(reflect.ValueOf(g.RandomDate(2000, 2023)))
		} else {
			g.fillStructValue(v)
		}

	case reflect.Ptr:
		ptr := reflect.New(t.Elem())
		g.fillValue(ptr.Elem(), t.Elem())
		v.Set(ptr)

	case reflect.Slice:
		// 生成嵌套切片时，限制元素数量避免无限递归
		elemType := t.Elem()
		length := g.RandomInt(0, 3)
		slice := reflect.MakeSlice(t, length, length)

		for i := 0; i < length; i++ {
			g.fillValue(slice.Index(i), elemType)
		}

		v.Set(slice)
	}
}
