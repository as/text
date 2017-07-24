package text

import (
	"image"
)

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
	SetOrigin(int64)
	Fill()
}

type Editor interface {
	Buffer
	Selector
}

type Plane interface {
	Bounds() image.Rectangle
	Size() image.Point
}

type Win interface {
	Editor
	Scroller
}
