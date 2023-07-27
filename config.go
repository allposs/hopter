package hopter

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// Customize 用户自定义配置
type Customize map[string]any

// CustomizeInterface 自定义配置接口
type CustomizeInterface interface {
	Get(str string) any
}

// Get 用户获取配置信息
func (c *Customize) Get(str string) any {
	prefix := strings.Split(str, ".")
	getValue := getConfigValue(*c, prefix, 0)
	if getValue != nil {
		return getValue
	}
	return nil
}

// getConfigValue 递归读取用户配置文件
func getConfigValue(c Customize, prefix []string, index int) any {
	key := prefix[index]
	if v, ok := c[key]; ok {
		if index == len(prefix)-1 {
			//到了最后一个
			return v
		}
		index = index + 1
		if mv, ok := v.(Customize); ok {
			//值必须是Config类型
			return getConfigValue(mv, prefix, index)
		}
	}
	return nil
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port           string `yaml:"port"`
	IP             string `yaml:"ip"`
	ReadTimeout    int    `yaml:"readTimeout"`
	WriteTimeout   int    `yaml:"writeTimeout"`
	IdleTimeout    int    `yaml:"idleTimeout"`
	MaxHeaderBytes int    `yaml:"maxHeaderBytes"`
	SessionKey     string `yaml:"sessionKey"`
}

// config 配置文件
type config struct {
	server *ServerConfig
	log    *LogConfig
	config *Customize
}

// ConfigInterface web配置接口
type ConfigInterface interface {
	Server() *ServerConfig
	Log() *LogConfig
	Customize() CustomizeInterface
	Config() ConfigInterface
}

// NewConfig 新配置
func NewConfig() ConfigInterface {
	return &config{server: &ServerConfig{Port: "8080", IP: "0.0.0.0"}, log: &LogConfig{}}
}

func (c *config) Server() *ServerConfig {
	return c.server
}

func (c *config) Log() *LogConfig {
	return c.log
}

func (c *config) Customize() CustomizeInterface {
	return c.config
}

func (c *config) Config() ConfigInterface {
	if b := loadConfigFile(); b != nil {
		err := yaml.Unmarshal(b, c)
		if err != nil {
			Panic("系统初始化异常:服务器解析配置文件异常，%v", err)
		}
	}
	return c
}
