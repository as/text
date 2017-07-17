package main

import (

	//	"github.com/as/clip"
	//
	"image"
	"image/draw"
	"io/ioutil"
	"log"
	"os"
	"time"
	"sync"
	"fmt"
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

var winSize = image.Pt(1920,1080)
var tagY = 11

func p(e mouse.Event) image.Point {
	return image.Pt(int(e.X), int(e.Y))
}

func swapcolor(f *frame.Frame){
	f.Color.Pallete, f.Color.Hi = f.Color.Hi,	f.Color.Pallete
}

func main() {
	driver.Main(func(src screen.Screen) {
		wind, _ := src.NewWindow(&screen.NewWindowOptions{winSize.X, winSize.Y, "Win"})
		wind.Send(paint.Event{})
		focused := false
		focused = focused
		pad := image.Pt(15,15)
		b, err := src.NewBuffer(winSize)
		if err != nil {
			log.Fatalln(err)
		}
		draw.Draw(b.RGBA(), b.Bounds(), frame.Acme.Back, image.ZP, draw.Src)
		wind.Upload(image.ZP, b, b.Bounds())
		w := win.New(image.ZP, pad, b.RGBA())
		mousein := NewMouse(time.Second/3, wind)

		go func() {
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
		}()
		var q0, q1 int64
		var but = 0
		var reg = 0
		r := w.Bounds()
		mousein.Machine.SetRect(image.Rect(r.Min.X, r.Min.Y+pad.Y, r.Max.X, r.Max.Y-pad.Y))
_=fmt.Println
		s := int64(0)
		for {
			switch e := wind.NextEvent().(type) {
			case Drain:
				DrainLoop:
				for{
				switch wind.NextEvent().(type){
				case DrainStop:
					break DrainLoop
				}}
			case SweepEvent:
				r := w.Bounds()
				if int(e.Y) < (r.Min.Y+pad.Y){
					reg = 1
				} else if int(e.Y) > (r.Max.Y+13){
					reg = -1
				} else {
					reg = 0
				}
				if reg != 0 {
						//swapcolor(w.Frame)
						if e.Y < float32(pad.Y) {
							w.FrameScroll(-1+((int(e.Y)-pad.Y)/13))
						} else {
							w.FrameScroll(1+((int(e.Y)-r.Max.Y)/13))
						}
						//swapcolor(w.Frame)
						wind.SendFirst(Drain{})
						wind.Send(DrainStop{})
				} 
				q := w.Origin()+w.IndexOf(p(e.Event))
				if s == q0{
					if q < q0{
						q1 = q0
						s = q0
						w.Select(q, s); q0=q
					} else {
						w.Select(s, q); q1=q
					}
				} else {
					if q > q1{
						q0 = q1
						s = q1
						w.Select(s, q); q1=q
					} else {
						w.Select(q, s); q0=q
					}
				}
			
				if w.Dirty() {
					wind.Send(paint.Event{})
				}
			case mouse.Event:
				e.X -= float32(pad.X)
				e.Y -= float32(pad.Y)
				mousein.Sink <- e
			case key.Event:
				if e.Direction == 2 {
					continue
				}
				if e.Rune == '\r' {
					e.Rune = '\n'
				}
				if e.Code == key.CodeUpArrow {
					w.FrameScroll(-3)
				} else if e.Code == key.CodeDownArrow{
					w.FrameScroll(3)
				} else {
				q0, _ := w.Dot()
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
				w.Resize(winSize.Sub(image.Pt(0, tagY*2)))
				wind.Send(paint.Event{})
			case paint.Event:
				var wg sync.WaitGroup
				for _, r := range w.Cache() {
					wg.Add(1)
					go func(r image.Rectangle){wind.Upload(r.Min.Add(pad), b, r); wg.Done()}(r)
				}
				wg.Wait()
				wind.Publish()
				w.Flush()
			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					return
				}
				// NT doesn't repaint the window if another window covers it
				if e.Crosses(lifecycle.StageFocused) == lifecycle.CrossOff {
					focused = false
				} else if e.Crosses(lifecycle.StageFocused) == lifecycle.CrossOn {
					focused = true
				}
			case MarkEvent:
				q0 = w.Origin()+w.IndexOf(p(e.Event))
				q1 = q0
				s = q0
				but = int(e.Button)
			case ClickEvent, SelectEvent:
				if q0 > q1 {
					q0, q1 = q1, q0
				}
but=but
				//w.Select2(but, q0, q1)
				w.Select(q0, q1)
				if w.Dirty() {
					wind.Send(paint.Event{})
				}
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
