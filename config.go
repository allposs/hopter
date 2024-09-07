package hopter

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// DataNotFount 数据不存在
var DataNotFount = fmt.Errorf("data not fount")

// config 配置
type config map[string]any

// Get 用户获取配置信息
func (c *config) Get(str string) any {
	prefix := strings.Split(str, ".")
	getValue := getConfigValue(*c, prefix, 0)
	if getValue != nil {
		return getValue
	}
	return nil
}

// Set 设置配置参数
func (c *config) Set(str string, value any) Config {
	prefix := strings.Split(str, ".")
	config := set(*c, prefix, 0, value)
	c = &config
	return c
}

func set(c config, prefix []string, index int, value any) config {
	key := prefix[index]
	_, ok := c[key]
	if !ok {
		c[key] = map[string]any{}
	}
	if index == len(prefix)-1 {
		//到了最后一个
		c[key] = value
		return c
	}
	if conf, is := c[key].(config); is {
		c[key] = set(conf, prefix, index+1, value)
	}
	if conf, is := c[key].(map[string]any); is {
		c[key] = set(conf, prefix, index+1, value)
	}
	return c
}

// getConfigValue 递归读取用户配置文件
func getConfigValue(c config, prefix []string, index int) any {
	key := prefix[index]
	if v, ok := c[key]; ok {
		if index == len(prefix)-1 {
			//到了最后一个
			return v
		}
		index = index + 1
		if mv, ok := v.(config); ok {
			//值必须是Config类型
			return getConfigValue(mv, prefix, index)
		}
	}
	return nil
}

// Config 配置接口
type Config interface {
	Get(str string) any
	Unmarshal(str string, value any) error
	Read() Config
	Set(str string, value any) Config
}

// NewConfig 新配置
func NewConfig() Config {
	cfg := make(config, 0)
	cfg["server"] = map[string]any{"port": "8080", "ip": "0.0.0.0"}
	cfg["log"] = map[string]any{"logLevel": "info"}
	return &cfg
}

func (c *config) Read() Config {
	if b := loadConfigFile(); b != nil {
		err := yaml.Unmarshal(b, c)
		if err != nil {
			Warn("web服务启动异常:服务器解析配置文件异常，%v", err)
		}
	}
	return c
}

func (c *config) Unmarshal(str string, value any) error {
	if v := c.Get(str); v != nil {
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		return json.Unmarshal(b, value)
	}
	return DataNotFount
}

// ginConfig 服务器配置
type ginConfig struct {
	Port           string `yaml:"port"`
	IP             string `yaml:"ip"`
	ReadTimeout    int    `yaml:"readTimeout"`
	WriteTimeout   int    `yaml:"writeTimeout"`
	IdleTimeout    int    `yaml:"idleTimeout"`
	MaxHeaderBytes int    `yaml:"maxHeaderBytes"`
	SessionKey     string `yaml:"sessionKey"`
}

// defaultGinConfig 默认配置
func defaultGinConfig() *ginConfig {
	res := new(ginConfig)
	res.IP = "0.0.0.0"
	res.Port = "8080"
	res.ReadTimeout = 30
	res.WriteTimeout = 30
	res.IdleTimeout = 30
	res.MaxHeaderBytes = 16384
	res.SessionKey = sessionKeyPairs
	return res
}

// Endpoint 对外端点
type Endpoint struct {
	config Config
	logs   *Klogger
}

// Config 获取配置
func (e *Endpoint) Config() Config {
	return e.config
}

// Logs 获取日志
func (e *Endpoint) Logs() *Klogger {
	return e.logs
}
