package hopter

import (
	"github.com/gin-gonic/gin"
)

// MessageClass 消息类型
type MessageClass string

const (
	//JSONClass json类型
	JSONClass MessageClass = "json"
	//WebsocketClass websocket类型
	WebsocketClass MessageClass = "websocket"
)

// Message 返回消息接口
type Message interface {
	String() string
	IsClass() MessageClass
}

// errorHandler 错误处理中间件
func errorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if e := recover(); e != nil {
				switch msg := e.(type) {
				case Message:
					ctx.AbortWithStatusJSON(200, msg.String())
				default:
					ctx.AbortWithStatusJSON(401, gin.H{"服务器异常:%v,请联系管理人员": msg})
				}
			}
		}()
		ctx.Next()
	}
}

// Context gin的context封装
type Context struct {
	*gin.Context
}

// HandlerFunc  处理函数
type HandlerFunc func(ctx *Context) Message

// RespondTo 消息返回处理
func (h HandlerFunc) RespondTo() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		handlerFunc := h(&Context{ctx})
		switch handlerFunc.IsClass() {
		case WebsocketClass:
			return
		case JSONClass:
			ctx.Writer.Header().Set("Content-type", "application/json")
			if _, err := ctx.Writer.WriteString(handlerFunc.String()); err != nil {
				Panic("系统初始化异常:服务器写入消息异常，%v", err)
			}
			return
		default:
			return
		}
	}
}
