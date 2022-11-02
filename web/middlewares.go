package web

import "github.com/gin-gonic/gin"

// Middleware 中间件接口
type Middleware interface {
	// OnRequest Middleware 主体
	OnRequest(ctx *gin.Context) error
	//OnInject 用于对象注入
	OnInject() interface{}
}
