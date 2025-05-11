package hopter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	ctx "github.com/gorilla/context"
	"github.com/gorilla/sessions"
)

// sessionKeyPairs session默认的KeyPairs
var sessionKeyPairs string = "yostar.com"

// Store 存储接口
type Store interface {
	sessions.Store
	Options(Options)
}

// Session Session方法接口
type Session interface {
	ID() string
	Get(key any) any
	Set(key any, val any)
	Delete(key any)
	Clear()
	AddFlash(value any, vars ...string)
	Flashes(vars ...string) []any
	Options(Options)
	Save() error
}

type session struct {
	name    string
	request *http.Request
	store   Store
	session *sessions.Session
	written bool
	writer  http.ResponseWriter
}

func (s *session) ID() string {
	return s.Session().ID
}

func (s *session) Get(key any) any {
	return s.Session().Values[key]
}

func (s *session) Set(key any, val any) {
	s.Session().Values[key] = val
	s.written = true
}

func (s *session) Delete(key any) {
	delete(s.Session().Values, key)
	s.written = true
}

func (s *session) Clear() {
	for key := range s.Session().Values {
		s.Delete(key)
	}
}

func (s *session) AddFlash(value any, vars ...string) {
	s.Session().AddFlash(value, vars...)
	s.written = true
}

func (s *session) Flashes(vars ...string) []any {
	s.written = true
	return s.Session().Flashes(vars...)
}

func (s *session) Options(options Options) {
	s.written = true
	s.Session().Options = options.ToGorillaOptions()
}

func (s *session) Save() error {
	if s.Written() {
		e := s.Session().Save(s.request, s.writer)
		if e == nil {
			s.written = false
		}
		return e
	}
	return nil
}

func (s *session) Session() *sessions.Session {
	if s.session == nil {
		var err error
		s.session, err = s.store.Get(s.request, s.name)
		if err != nil {
			Error("[sessions] ERROR! %s\n", err)
		}
	}
	return s.session
}

func (s *session) Written() bool {
	return s.written
}

// Session 获取Session
func (ctx *Context) Session(name string) Session {
	return ctx.MustGet("SessionStore").(map[string]Session)[name]
}

type Options struct {
	Path     string
	Domain   string
	MaxAge   int
	Secure   bool
	HTTPOnly bool
	SameSite http.SameSite
}

func (options Options) ToGorillaOptions() *sessions.Options {
	return &sessions.Options{
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HTTPOnly,
		SameSite: options.SameSite,
	}
}

// sessionsMany 复数session
func sessionsMany(store Store, names ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessions := make(map[string]Session, len(names))
		for _, name := range names {
			sessions[name] = &session{name, c.Request, store, nil, false, c.Writer}
		}
		c.Set("SessionStore", sessions)
		defer ctx.Clear(c.Request)
		c.Next()
	}
}

// SetSessionsStore Sessions存储
func (e *Engine) SetSessionsStore(store Store, names ...string) *Engine {
	e.engine.Use(sessionsMany(store, names...))
	return e
}
