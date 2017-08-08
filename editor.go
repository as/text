package text

import (
	"image"
)

type Buffer interface {
	Insert(p []byte, at int64) (n int)
	Delete(q0, q1 int64) (n int)
	Len() int64
	Bytes() []byte
}

type Selector interface {
	Select(q0, q1 int64)
	Dot() (q0, q1 int64)
}

type Editor interface {
	Buffer
	Selector
}

type Plane interface {
	Bounds() image.Rectangle
	Size() image.Point
}

type Projector interface {
	PointOf(q int64) (pt image.Point)
	IndexOf(pt image.Point) (q int64)
}

type Ruler interface {
	Measure(s string) int
	Height() int
}

type Win interface {
	Editor
	Scroller
	Plane
	Mark()
}

type Sweeper interface {
	Plane
	Projector
	Scroller
	Selector
}

type Scroller interface {
	Origin() int64
	SetOrigin(int64, bool)
	Fill()
	Scroll(int)
}

type Inverse struct {
	e interface{}
}

type History interface {
	Next() interface{}
	Prev() interface{}
	Event() interface{}
	Add(e interface{})
	Commit()
	Apply()
}

type Sender interface {
	Send(interface{})
	SendFirst(interface{})
}
