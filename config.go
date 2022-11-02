package hopter

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 用户自定义配置
type Config map[string]any

// Get 用户获取配置信息
func (c *Config) Get(str string) interface{} {
	prefix := strings.Split(str, ".")
	getValue := getConfigValue(*c, prefix, 0)
	if getValue != nil {
		return getValue
	}
	return nil
}

// getConfigValue 递归读取用户配置文件
func getConfigValue(c Config, prefix []string, index int) interface{} {
	key := prefix[index]
	if v, ok := c[key]; ok {
		if index == len(prefix)-1 { //到了最后一个
			return v
		} else {
			index = index + 1
			if mv, ok := v.(Config); ok {
				//值必须是Config类型
				return getConfigValue(mv, prefix, index)
			}
		}
	}
	return nil
}

// serverConfig 服务器配置
type serverConfig struct {
	Port           string `yaml:"port"`
	IP             string `yaml:"ip"`
	ReadTimeout    int    `yaml:"readTimeout"`
	WriteTimeout   int    `yaml:"writeTimeout"`
	IdleTimeout    int    `yaml:"idleTimeout"`
	MaxHeaderBytes int    `yaml:"maxHeaderBytes"`
}

// config 配置文件
type config struct {
	Server *serverConfig
	Logs   *option
	Config *Config
}

// newConfig 新的配置文件
func newConfig() *config {
	return &config{Server: &serverConfig{Port: "8080", IP: "0.0.0.0"}, Logs: &option{}}
}

// initConfig 初始化配置文件
func initConfig() *config {
	conf := newConfig()
	if b := loadConfigFile(); b != nil {
		err := yaml.Unmarshal(b, conf)
		if err != nil {
			Panic("系统初始化异常:服务器解析配置文件异常，%v", err)
		}
	}
	return conf
}
