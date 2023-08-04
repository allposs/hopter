package hopter

// Interface web对外接口
type Interface interface {
	Windows(w *Web)
	Init()
}
