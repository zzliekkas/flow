package providers

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zzliekkas/flow/config"
)

// Kd100Service 封装快递100相关API
type Kd100Service struct {
	cfg config.Kd100Config
}

// NewKd100Service 创建服务实例
func NewKd100Service(cfg config.Kd100Config) *Kd100Service {
	return &Kd100Service{cfg: cfg}
}

// Sign 生成快递100签名
func (s *Kd100Service) Sign(param string) string {
	// param+key+customer+salt（部分接口需要salt）
	str := param + s.cfg.Key + s.cfg.Customer
	if s.cfg.Salt != "" {
		str += s.cfg.Salt
	}
	h := md5.New()
	h.Write([]byte(str))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

// QueryTrack 查询物流轨迹
func (s *Kd100Service) QueryTrack(com, num string) (string, error) {
	paramMap := map[string]string{
		"com": com,
		"num": num,
	}
	paramBytes, _ := json.Marshal(paramMap)
	param := string(paramBytes)

	postData := map[string]string{
		"customer": s.cfg.Customer,
		"param":    param,
		"sign":     s.Sign(param),
	}
	postBytes, _ := json.Marshal(postData)

	url := s.cfg.BaseURL + "/poll/query.do"
	client := &http.Client{Timeout: time.Duration(s.cfg.Timeout) * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(postBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
}
