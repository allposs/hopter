package hopter

// Middleware 中间件接口
type Middleware interface {
	// OnRequest Middleware 主体
	OnRequest(ctx *Context) error
	//OnInject 用于对象注入
	OnInject() any
}
