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
	"sync"
	"time"

	"github.com/as/frame/font"
	"github.com/as/text/win"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"

	kbd "github.com/as/text/kbd"
	mous "github.com/as/text/mouse"
)

var winSize = image.Pt(1800, 1000)
var tagY = 16
var fontdy = 16

func p(e mouse.Event) image.Point {
	return image.Pt(int(e.X), int(e.Y))
}

func main() {
	driver.Main(func(src screen.Screen) {
		wind, _ := src.NewWindow(&screen.NewWindowOptions{winSize.X, winSize.Y, "Win"})
		wind.Send(paint.Event{})
		focused := false
		focused = focused
		pad := image.Pt(fontdy, fontdy)
		b, err := src.NewBuffer(winSize)
		if err != nil {
			log.Fatalln(err)
		}
		wind.Upload(image.ZP, b, b.Bounds())
		sp := image.ZP
		w := win.New(sp, pad, b.RGBA(), font.NewGoMono(fontdy))
		wind.Upload(sp, b, b.Bounds())
		wind.Send(paint.Event{})
		mousein := mous.NewMouse(time.Second/3, wind)

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
			wind.Send(paint.Event{})
		}()
		var q0, q1, s int64
		var but = 0
		r := w.Bounds()
		mousein.Machine.SetRect(image.Rect(r.Min.X, r.Min.Y+pad.Y, r.Max.X, r.Max.Y-pad.Y))
		_ = fmt.Println
		for {
			switch e := wind.NextEvent().(type) {
			case mouse.Event:
				e.X -= float32(sp.X)
				e.Y -= float32(sp.Y)
				mousein.Sink <- e
			case mous.Drain:
			DrainLoop:
				for {
					switch wind.NextEvent().(type) {
					case mous.DrainStop:
						break DrainLoop
					}
				}
			case mous.SweepEvent:
				s, q0, q1 = mous.Sweep(w, e, pad.Y, s, q0, q1, wind)
				if w.Dirty() {
					wind.Send(paint.Event{})
				}
			case mous.MarkEvent:
				q0 = w.Origin() + w.IndexOf(p(e.Event))
				q1 = q0
				s = q0
				but = int(e.Button)
				w.Select(q0, q1)
				if w.Dirty() {
					wind.Send(paint.Event{})
				}
			case mous.ClickEvent, mous.SelectEvent:
				but = but
				//w.Select2(but, q0, q1)
				w.Select(q0, q1)
				if w.Dirty() {
					wind.Send(paint.Event{})
				}
			case key.Event:
				if e.Direction == 2 {
					continue
				}
				kbd.Send(w, e)
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
				wg.Add(len(w.Cache()))
				for _, r := range w.Cache() {
					go func(r image.Rectangle) { wind.Upload(sp.Add(r.Min), b, r); wg.Done() }(r)
				}
				wind.Publish()
				w.Flush()
				wg.Wait()
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
