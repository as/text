package text

import (
	"image"
)

type Sender interface{
	Send(interface{})
	SendFirst(interface{})
}

type Projector interface {
	PointOf(q int64) (pt image.Point)
	IndexOf(pt image.Point) (q int64)
}

type Ruler interface {
	Measure(s string) int
	Height() int
}
type Scroller interface {
	Origin() int64
	SetOrigin(int64, bool)
	Fill()
	Scroll(int)
}

type Editor interface {
	Buffer
	Selector
}

type Sweeper interface{
	Bounds() image.Rectangle
	Projector
	Scroller
	Selector
}

type Plane interface {
	Bounds() image.Rectangle
	Size() image.Point
}

type Win interface {
	Editor
	Scroller
	Plane
	Mark()
}
