package hopter

// Service web对外接口
type Service interface {
	Init()
	Handles(w *Web)
}

// Mount 挂载接口
func (w *Web) Mount(group string, class ...Service) *Web {
	w.group = w.engine.Group(group)
	for _, v := range class {
		w.beanFactory.Inject(v)
		w.Beans(v)
	}
	for _, v := range class {
		v.Init()
	}
	for _, v := range class {
		v.Handles(w)
	}
	return w
}
