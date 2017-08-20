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
	"golang.org/x/image/font/gofont/gomedium"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"github.com/as/text"
	"github.com/as/worm"

	kbd "github.com/as/text/kbd"
	mous "github.com/as/text/mouse"
)

var winSize = image.Pt(2560,1350)
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

		sp  := image.Pt(0,0)
		sp1 := image.Pt(winSize.X-winSize.X/3,0 )

		r := image.Rect(0,0,winSize.X,winSize.Y)
		r.Max.X -= r.Max.X/3
		b, _ := src.NewBuffer(r.Max)



		lg := worm.NewCoalescer(worm.NewLogger(), time.Second*1)



		ed, _ := text.Open(text.NewBuffer())
		ed = text.NewHistory( text.Trace(ed),lg )
		w := win.New(r.Min, pad, b.RGBA(), ed, font.NewTTF(gomedium.TTF, fontdy))

		r.Min.X = r.Max.X
		r.Max.X = winSize.X
		b1, _ := src.NewBuffer(image.Pt(winSize.X/3,winSize.Y))
		st0 := win.New(sp1, pad, b1.RGBA(), nil, font.NewTTF(gomedium.TTF, 11))

		wind.Upload(image.Pt(0,0), b, b.Bounds())
		wind.Upload(r.Min, b1, b.Bounds())
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
		r = w.Bounds()
		mousein.Machine.SetRect(image.Rect(r.Min.X, r.Min.Y+pad.Y, r.Max.X, r.Max.Y-pad.Y))
		_ = fmt.Println
		last := int64(0)
		ckdirt := func(){
				dirty := false
				for ;last != lg.Len(); last++{
					p, _ := lg.ReadAt(last)
					st0.Insert([]byte(fmt.Sprintf("%d\t%s", last, p)), 0)
					dirty=true
				}
				if dirty ||st0.Dirty() || w.Dirty(){
					wind.Send(paint.Event{})
				}
		}
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
				ckdirt()
			case mous.MarkEvent:
				q0 = w.Origin() + w.IndexOf(p(e.Event))
				q1 = q0
				s = q0
				but = int(e.Button)
				w.Select(q0, q1)
				ckdirt()
			case mous.ClickEvent, mous.SelectEvent:
				but = but
				//w.Select2(but, q0, q1)
				w.Select(q0, q1)
				ckdirt()
			case key.Event:
				if e.Direction == 2 {
					continue
				}
				if byte(e.Rune) == '\x1a'{
					
				}
				kbd.SendClient(w, e)
				ckdirt()
			case size.Event:
				pt := image.Pt(e.WidthPx, e.HeightPx)
				if pt.X < 10 || pt.Y < 10 {
					println("ignore daft size request:", pt.String())
					continue
				}
				winSize = pt
				//w.Resize(winSize.Sub(image.Pt(0, tagY*2)))
				wind.Send(paint.Event{})
			case paint.Event:
				var wg sync.WaitGroup
				wg.Add(len(w.Cache()))
				for _, r := range w.Cache() {
					go func(r image.Rectangle) { wind.Upload(sp.Add(r.Min), b, r); wg.Done() }(r)
				}
				wg.Add(len(st0.Cache()))
				for _, r := range st0.Cache() {
					go func(r image.Rectangle) { wind.Upload(sp1.Add(r.Min), b1, r); wg.Done() }(r)
				}
				//wind.Upload(image.Pt(winSize.X-winSize.X/3,0), b1, b1.Bounds())
				wind.Publish()
				w.Flush()
				st0.Flush()
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
