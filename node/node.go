package node

import (
	"image"

	"github.com/as/shiny/screen"
)

type Node struct {
	Sp, size, pad image.Point
	dirty         bool
}

func (n Node) Size() image.Point {
	return n.Size()
}
func (n Node) Pad() image.Point {
	return n.Sp.Add(n.Size())
}

type Device struct {
	scr    screen.Screen
	events screen.Window
	b      screen.Buffer
}
