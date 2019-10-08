package engine

type DisplayObject struct {
	Children []*DisplayObject
	Draw     func(e *UIEngine)
}
