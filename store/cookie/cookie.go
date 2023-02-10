package cookie

import (
	web "github.com/allposs/hopter"
	"github.com/gorilla/sessions"
)

// Store 存储接口
type Store interface {
	web.Store
}

// NewStore 创建新的存储
func NewStore(keyPairs ...[]byte) Store {
	return &store{sessions.NewCookieStore(keyPairs...)}
}

// store store结构体
type store struct {
	*sessions.CookieStore
}

// Options 参数设置
func (c *store) Options(options web.Options) {
	c.CookieStore.Options = options.ToGorillaOptions()
}
