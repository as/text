package main

import (

	//	"github.com/as/clip"
	//
	"fmt"
	"image"
	"image/draw"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/as/frame"
	"github.com/as/text/win"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

var winSize = image.Pt(1800, 1000)
var tagY = 16
var fontdy = 16

func p(e mouse.Event) image.Point {
	return image.Pt(int(e.X), int(e.Y))
}

func swapcolor(f *frame.Frame) {
	f.Color.Pallete, f.Color.Hi = f.Color.Hi, f.Color.Pallete
}
func no(err error){if err != nil{log.Fatalln(err)}}
func main() {
	driver.Main(func(src screen.Screen) {
		wind, _ := src.NewWindow(&screen.NewWindowOptions{winSize.X, winSize.Y, "Win"})
		wind.Send(paint.Event{})
		focused := false
		focused = focused
		pad := image.Pt(15, fontdy*2)
		sp := image.ZP
		ft := frame.NewGoMono(fontdy)
		ft2 := frame.NewGoMono(12)
		w, err := win.New(src, sp, pad, winSize, &ft, nil)
		dbg, _ := win.New(src, image.Pt(pad.X, 0), image.ZP, image.Pt(150,50), &ft2, nil)
		no(err)
		mousein := NewMouse(time.Second/3, wind)
		go func() {
			time.Sleep(2*time.Second)
			file := `\windows\system32\drivers\etc\hosts`
			if len(os.Args) > 1 {
				file = os.Args[1]
			}
			data, err := ioutil.ReadFile(file)
			if err != nil {
				log.Println(err)
				return
			}
			w.Insert(data, 0)
			wind.Send(paint.Event{})
		}()
		var q0, q1 int64
		var but = 0
		var reg = 0
		r := w.Bounds()
		mousein.Machine.SetRect(image.Rect(r.Min.X, r.Min.Y+pad.Y, r.Max.X, r.Max.Y-pad.Y))
		_ = fmt.Println
		s := int64(0)
		for {
			switch e := wind.NextEvent().(type) {
			case Drain:
			DrainLoop:
				for {
					switch wind.NextEvent().(type) {
					case DrainStop:
						break DrainLoop
					}
				}
			case SweepEvent:
				r := w.Frame.Bounds()
				if int(e.Y) < (r.Min.Y) {
					reg = 1
				} else if int(e.Y) > (r.Max.Y) {
					reg = -1
				} else {
					reg = 0
				}
				if reg != 0 {
					//swapcolor(w.Frame)
					if e.Y < float32(pad.Y) {
						w.FrameScroll((int(e.Y)/-(fontdy/3))*1)
					} else {
						w.FrameScroll(1 + ((int(e.Y)-r.Max.Y)/(fontdy/3))*1)
					}
					//swapcolor(w.Frame)
					wind.SendFirst(Drain{})
					wind.Send(DrainStop{})
				}
				q := w.Origin() + w.IndexOf(p(e.Event))
				if s == q0 {
					if q < q0 {
						q1 = q0
						s = q0
						w.Select(q, s)
						q0 = q
					} else {
						w.Select(s, q)
						q1 = q
					}
				} else {
					if q > q1 {
						q0 = q1
						s = q1
						w.Select(s, q)
						q1 = q
					} else {
						w.Select(q, s)
						q0 = q
					}
				}

				if w.Dirty() {
					wind.Send(paint.Event{})
				}
			case mouse.Event:
				e.X -= float32(sp.X)
				e.Y -= float32(sp.Y)
				mousein.Sink <- e
			case key.Event:
				if e.Direction == 2 {
					continue
				}
				if e.Rune == '\r' {
					e.Rune = '\n'
				}
				if e.Code == key.CodeUpArrow {
					w.FrameScroll(-2)
				} else if e.Code == key.CodeDownArrow {
					w.FrameScroll(2)
				} else {
					if e.Rune == -1{
						continue
					}
					q0, _ = w.Dot()
					w.Insert([]byte{byte(e.Rune)}, q0)
					q0++
					w.Select(q0, q0)
				}
				if w.Dirty() {
					wind.Send(paint.Event{})
				}
			case size.Event:
				pt := image.Pt(e.WidthPx, e.HeightPx)
				if pt.X < 10 || pt.Y < 10 {
					println("ignore daft size request:", pt.String())
					continue
				}
				winSize = pt
				w.Resize(winSize)
				wind.Send(paint.Event{})
			case paint.Event:
				if !focused {
					w.Refresh()
				}
				p0,p1 := w.Frame.Dot(); q0,q1 := w.Dot(); org := w.Origin(); fl := w.Frame.Nchars
				dbg.Insert([]byte(fmt.Sprintf("p=%d:%d q=%d:%d org=%d framelen=%d\n\n\n\n",p0,p1,q0,q1,org,fl)), 0)
				w.Upload(wind)
				//dbg.Refresh()
				dbg.Upload(wind)
				wind.Publish()
			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					return
				}
				// NT doesn't repaint the window if another window covers it
				if e.Crosses(lifecycle.StageFocused) == lifecycle.CrossOff {
					focused = false
					wind.Send(paint.Event{})
				} else if e.Crosses(lifecycle.StageFocused) == lifecycle.CrossOn {
					w.Refresh()
					focused = true
					wind.Send(paint.Event{})
				}
			case MarkEvent:
				q0 = w.Origin() + w.IndexOf(p(e.Event))
				q1 = q0
				s = q0
				but = int(e.Button)
			case ClickEvent, SelectEvent:
				if q0 > q1 {
					q0, q1 = q1, q0
				}
				but = but
				//w.Select2(but, q0, q1)
				w.Select(q0, q1)
				//if w.Dirty() {
					wind.Send(paint.Event{})
				//}
			}
		}
	})
}

func region(c, p0, p1 int64) int {
	if c < p0 {
		return -1
	}
	if c >= p1 {
		return 1
	}
	return 0
}

func drawBorder(dst draw.Image, r image.Rectangle, src image.Image, sp image.Point, thick int) {
	draw.Draw(dst, image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+thick), src, sp, draw.Src)
	draw.Draw(dst, image.Rect(r.Min.X, r.Max.Y-thick, r.Max.X, r.Max.Y), src, sp, draw.Src)
	draw.Draw(dst, image.Rect(r.Min.X, r.Min.Y, r.Min.X+thick, r.Max.Y), src, sp, draw.Src)
	draw.Draw(dst, image.Rect(r.Max.X-thick, r.Min.Y, r.Max.X, r.Max.Y), src, sp, draw.Src)
}
