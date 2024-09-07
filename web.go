package hopter

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Engine web程序结构
type Engine struct {
	engine      *gin.Engine
	group       *gin.RouterGroup
	beanFactory *BeanFactory
	server      *http.Server
}

func init() {
	gin.SetMode(gin.ReleaseMode)
}

// New 创建web程序
func New(conf Config) *Engine {
	var this = &Engine{}
	logger, err := initLog(conf.Read())
	if err != nil {
		Fatal("web服务启动失败:初始化日志错误，%v", err)
	}
	switch strings.ToLower(Level) {
	case "debug":
		gin.SetMode(gin.DebugMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
	this.server = &http.Server{
		Addr:           "0.0.0.0:8080",
		Handler:        this.engine,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    30 * time.Second,
		MaxHeaderBytes: 16384,
	}
	this.engine = gin.Default()
	this.beanFactory = NewBeanFactory()
	this.beanFactory.set(&Endpoint{conf, logger})
	this.engine.Use(recovered())
	this.engine.Use(LogMiddleware())
	this.metric()
	return this
}

// Context gin的context封装
type Context struct {
	*gin.Context
}

type Message interface {
}

// HandlerFunc  处理函数
type HandlerFunc func(ctx *Context) Message

// Func 消息返回处理
func (h HandlerFunc) Func() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		h(&Context{ctx})
	}
}

// Handle  重载gin的handle方法
func (e *Engine) Handle(httpMethod, relativePath string, handler HandlerFunc) *Engine {
	e.group.Handle(httpMethod, relativePath, handler.Func())
	return e
}

// Run 运行Web程序
func (e *Engine) Run() {
	if bean := e.beanFactory.get(new(Endpoint)); bean != nil {
		if conf, ok := bean.(*Endpoint); ok {
			value := defaultGinConfig()
			if v := conf.Config().Get("server"); v != nil {
				if err := conf.Config().Unmarshal("server", value); err != nil {
					Fatal("web服务启动失败:获取启动参数异常,%v", err)
				}
			}
			e.server.Addr = fmt.Sprintf("%s:%s", value.IP, value.Port)
			e.server.ReadTimeout = time.Duration(value.ReadTimeout) * time.Second
			e.server.WriteTimeout = time.Duration(value.WriteTimeout) * time.Second
			e.server.IdleTimeout = time.Duration(value.IdleTimeout) * time.Second
			e.server.MaxHeaderBytes = value.MaxHeaderBytes
		}
	}
	e.server.Handler = e.engine
	if err := e.server.ListenAndServe(); err != nil {
		Fatal("web服务启动失败:服务器监听端口异常，%v", err)
	}
}

// Shutdown 关闭服务
func (e *Engine) Shutdown(ctx context.Context) error {
	return e.server.Shutdown(ctx)
}
