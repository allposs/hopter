package hopter

// Interface web对外接口
type Interface interface {
	Build(w *Web)
	Init()
}
