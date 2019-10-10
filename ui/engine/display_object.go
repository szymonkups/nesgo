package engine

type Displayable interface {
	GetChildren() []Displayable
	Draw(e *UIEngine) error
}
