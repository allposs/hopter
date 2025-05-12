package hopter

import (
	"github.com/spf13/viper"
)

// config 配置
type config struct {
	*viper.Viper
}

func NewConfig(path, prefix string) *config {
	res := &config{viper.New()}
	if path != "" {
		res.SetConfigFile(path)
	} else {
		res.AddConfigPath("./config")
		res.SetConfigName("config")
		res.SetConfigType("yaml")
	}

	// 读取环境变量
	if prefix != "" {
		res.AutomaticEnv()
		// 环境变量前缀
		res.SetEnvPrefix(prefix)
	}
	// 设置默认值
	res.SetDefault("server.port", 8000)
	res.SetDefault("server.ip", "0.0.0.0")
	res.SetDefault("server.readTimeout", 30)
	res.SetDefault("server.writeTimeout", 30)
	res.SetDefault("server.idleTimeout", 30)
	res.SetDefault("server.maxHeaderBytes", 16384)
	res.SetDefault("server.sessionKey", sessionKeyPairs)
	res.SetDefault("log.level", "info")
	res.SetDefault("log.path", "./logs/server.log")
	res.SetDefault("log.type", "text")
	return res
}

// Get 用户获取配置信息
func (c config) Get(str string) any {
	return c.Viper.Get(str)
}

// // Set 设置配置参数
func (c *config) Set(str string, value any) Config {
	c.Viper.Set(str, value)
	return c
}

// Config 配置接口
type Config interface {
	Get(str string) any
	UnmarshalKey(str string, value any, opts ...viper.DecoderConfigOption) error
	ReadInConfig() Config
	Set(str string, value any) Config
}

// ReadInConfig 读取配置文件
func (c *config) ReadInConfig() Config {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			Warn("web服务启动异常:服务器解析配置文件异常，%v", err)
		}
	}
	return c
}

// UnmarshalKey 用于将配置文件中的特定key的值解析并映射到一个结构体（Struct）中
func (c *config) UnmarshalKey(str string, value any, opts ...viper.DecoderConfigOption) error {
	return c.Viper.UnmarshalKey(str, value, opts...)
}

// ginConfig 服务器配置
type ginConfig struct {
	ENV            string `yaml:"env"`
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
