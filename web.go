package hopter

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/context"
)

// Web web程序结构
type Web struct {
	*gin.Engine
	group       *gin.RouterGroup
	beanFactory *BeanFactory
}

func init() {
	gin.SetMode(gin.ReleaseMode)
}

// New 创建web程序
func New() *Web {
	//配置文件
	conf := initConfig()
	var this = &Web{}
	logger, err := initLog(conf.Logs)
	if err != nil {
		Fatal("系统初始化异常:初始化日志错误，%v", err)
	}
	switch strings.ToLower(conf.Logs.LogLevel) {
	case "debug":
		gin.SetMode(gin.DebugMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
	this.Engine = gin.Default()
	this.beanFactory = NewBeanFactory()
	this.beanFactory.setBean(conf.Server)
	this.beanFactory.setBean(conf.Config)
	this.beanFactory.setBean(logger)
	//错误处理
	this.Use(errorHandler())
	this.Use(LogMiddleware())
	this.metricMiddleware()
	this.Beans(this)
	return this
}

// sessionsMany 复数session
func sessionsMany(store Store, names ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessions := make(map[string]Session, len(names))
		for _, name := range names {
			sessions[name] = &session{name, c.Request, store, nil, false, c.Writer}
		}
		c.Set("SessionStore", sessions)
		defer context.Clear(c.Request)
		c.Next()
	}
}

// SetSessionsStore Sessions存储
func (w *Web) SetSessionsStore(store Store, names ...string) *Web {
	w.Use(sessionsMany(store, names...))
	return w
}

// Run 运行Web程序
func (w *Web) Run() {
	web := http.Server{
		Addr:           "0.0.0.0:8080",
		Handler:        w,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    30 * time.Second,
		MaxHeaderBytes: 16384,
	}
	if bean := w.beanFactory.GetBean(new(serverConfig)); bean != nil {
		if conf, ok := bean.(*serverConfig); ok {
			web.Addr = fmt.Sprintf("%s:%s", conf.IP, conf.Port)
			web.ReadTimeout = time.Duration(conf.ReadTimeout) * time.Second
			web.WriteTimeout = time.Duration(conf.WriteTimeout) * time.Second
			web.IdleTimeout = time.Duration(conf.IdleTimeout) * time.Second
			web.MaxHeaderBytes = conf.MaxHeaderBytes
		}
	}
	if err := web.ListenAndServe(); err != nil {
		Fatal("系统初始化异常:服务器监听端口异常，%v", err)
	}
}

// Attach 中间件加入
func (w *Web) Attach(m Middleware) *Web {
	w.Use(func(ctx *gin.Context) {
		w.beanFactory.Inject(m.OnInject())
		err := m.OnRequest(&Context{ctx})
		if err != nil {
			ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		} else {
			ctx.Next()
		}
	})
	return w
}

// Handle  重载gin的handle方法
func (w *Web) Handle(httpMethod, relativePath string, handler HandlerFunc) *Web {
	w.group.Handle(httpMethod, relativePath, handler.RespondTo())
	return w
}

// Beans Bean注册
func (w *Web) Beans(beans ...any) *Web {
	w.beanFactory.setBean(beans...)
	return w
}

// Mount 挂载接口
func (w *Web) Mount(group string, class ...Interface) *Web {
	w.group = w.Group(group)
	for _, v := range class {
		w.beanFactory.inject(v)
		//		v.Init()
		//		v.Build(w)
		w.Beans(v)
	}
	for _, v := range class {
		v.Init()
	}
	for _, v := range class {
		v.Build(w)
	}
	return w
}
