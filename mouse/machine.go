package mouse

import (
	"image"
	"time"

	"github.com/as/text"
	"golang.org/x/mobile/event/mouse"
)

// Machine is the conduit that state transitions happen
// though. It contains a Skink chan for input mouse events
// that drive the StateFns
type Machine struct {
	r         image.Rectangle
	Sink      chan mouse.Event
	down      mouse.Button
	first     mouse.Event
	double    time.Duration
	lastsweep mouse.Event
	ctr       int
	// Should only send events, no recieving.
	text.Sender
	LastMark  MarkEvent
	Clickzone image.Rectangle
}

// NewMachine initialize a new state machine with no-op
// functions for all chording events.
func NewMachine(deque text.Sender) *Machine {
	return &Machine{
		Sink:   make(chan mouse.Event),
		Sender: deque,
		down:   0,
		double: time.Second / 3,
	}
}
func (m *Machine) SetRect(r image.Rectangle) {
	m.r = r
}
func (m *Machine) press(e mouse.Event) bool {
	if e.Direction != mouse.DirPress {
		return false
	}
	if e.Button == mouse.ButtonNone {
		return false
	}
	if e.Button&m.down != 0 {
		//return false
		//panic("bug: mouse button pressed > 1 without release")
	}
	m.down |= e.Button
	//fmt.Printf("press: event = %#v\n", e)
	return true
}
func (m *Machine) release(e mouse.Event) bool {
	if e.Direction != mouse.DirRelease {
		return false
	}
	if e.Button == mouse.ButtonNone {
		return false
	}
	if e.Button&m.down == 0 {
		return false
		//panic("bug: release unpressed button")
	}
	//fmt.Printf("release: event = %#v\n", e)
	m.down &= ^e.Button
	return true
}
func (m *Machine) CloseTo(e, f mouse.Event) bool {
	//fmt.Println(abs(int(e.X-f.X)))
	//fmt.Println(abs(int(e.Y-f.Y)))
	return abs(int(e.X-f.X)) < 2 && abs(int(e.Y-f.Y)) < 2
}
func (m *Machine) SetBounds(r image.Rectangle) {
	m.r = r
}

func (m *Machine) left(e mouse.Event) bool  { return e.Button == mouse.ButtonLeft }
func (m *Machine) right(e mouse.Event) bool { return e.Button == mouse.ButtonRight }
func (m *Machine) mid(e mouse.Event) bool   { return e.Button == mouse.ButtonMiddle }
func (m *Machine) none(e mouse.Event) bool  { return e.Button == mouse.ButtonNone }
func (m *Machine) terminates(e mouse.Event) bool {
	return m.release(e) && m.down == 0
}

func (m *Machine) Run() chan mouse.Event {
	go func() {
		fn := none
		dy := 0
		for e := range m.Sink {
			if e.Button.IsWheel() {
				dy++
				select {
				case e0 := <-m.Sink:
					if !e.Button.IsWheel() {
						m.Send(ScrollEvent{Event: e, Dy: 1})
						dy = 0
						e = e0
						fn = fn(m, e)
						break
					}
					dy++
				case <-clock:
					m.SendFirst(ScrollEvent{Event: e, Dy: dy})
					dy = 0
				}
			} else {
				dy = 0
				fn = fn(m, e)
			}
		}
	}()
	return m.Sink
}

var clock = time.NewTicker(time.Millisecond * 60).C

//var clock = time.NewTicker(time.Millisecond*20).C

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func pt(e mouse.Event) image.Point {
	return image.Pt(int(e.X), int(e.Y))
}
