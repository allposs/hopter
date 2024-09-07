package hopter

// Service web对外接口
type Service interface {
	Init()
	Handles(e *Engine)
}

// Mount 挂载接口
func (e *Engine) Mount(group string, class ...Service) *Engine {
	e.group = e.engine.Group(group)
	for _, v := range class {
		e.beanFactory.Inject(v)
		e.Beans(v)
	}
	for _, v := range class {
		v.Init()
	}
	for _, v := range class {
		v.Handles(e)
	}
	return e
}
