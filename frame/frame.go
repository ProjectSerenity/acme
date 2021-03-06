package frame

import (
	"9fans.net/go/draw"
	"image"
)

const (
	colBack = iota
	colHigh
	colBord
	colText
	colHText
	NumColours

	frtickw = 3
)

type frbox struct {
	Wid    int // in pixels
	Nrune  int
	Ptr    []byte
	Bc     rune
	Minwid byte
}

type Frame struct {
	Font         *draw.Font
	Display      *draw.Display           // on which the frame is displayed
	Background   *draw.Image             // on which the frame appears
	Cols         [NumColours]*draw.Image // background and text colours
	Rect         image.Rectangle         // in which the text appears
	Entire       image.Rectangle         // size of full frame
	Scroll       func(*Frame, int)       // function provided by application
	box          []*frbox
	p0, p1       uint64 // bounds of a selection
	nbox, nalloc int
	maxtab       int // max size of a tab (in pixels)
	nchars       int // number of runes in frame
	nlines       int // number of lines with text
	maxlines     int // total number of lines in frame
	lastlinefull int
	modified     bool
	tick         *draw.Image // typing tick
	tickback     *draw.Image // image under tick
	ticked       bool
	noredraw     bool
	tickscale    int // tick scaling factor
}

// NewFrame creates a new Frame with Font ft, background image b, colours cols, and
// of the size r
func NewFrame(r image.Rectangle, ft *draw.Font, b *draw.Image, cols [NumColours]*draw.Image) *Frame {
	f := new(Frame)
	f.Font = ft
	f.Display = b.Display
	f.maxtab = 8 * ft.StringWidth("0")
	f.nbox = 0
	f.nalloc = 0
	f.nchars = 0
	f.nlines = 0
	f.p0 = 0
	f.p1 = 0
	f.box = nil
	f.lastlinefull = 0
	f.Cols = cols
	f.SetRects(r, b)

	if f.tick == nil && f.Cols[colBack] != nil {
		f.InitTick()
	}
	return f
}

// InitTick
func (f *Frame) InitTick() {
	var err error
	if f.Cols[colBack] == nil || f.Display == nil {
		return
	}

	f.tickscale = f.Display.ScaleSize(1)
	b := f.Display.ScreenImage
	ft := f.Font

	if f.tick != nil {
		f.tick.Free()
	}

	f.tick, err = f.Display.AllocImage(image.Rect(0, 0, f.tickscale*frtickw, ft.Height), b.Pix, false, draw.White)
	if err != nil {
		return
	}

	f.tickback, err = f.Display.AllocImage(f.tick.R, b.Pix, false, draw.White)
	if err != nil {
		f.tick.Free()
		f.tick = nil
		return
	}

	// background colour
	f.tick.Draw(f.tick.R, f.Cols[colBack], nil, image.Pt(0, 0))
	// vertical line
	f.tick.Draw(image.Rect(f.tickscale*(frtickw/2), 0, f.tickscale*(frtickw/2+1), ft.Height), f.Display.Black, nil, image.Pt(0, 0))
	// box on each end
	f.tick.Draw(image.Rect(0, 0, f.tickscale*frtickw, f.tickscale*frtickw), f.Cols[colText], nil, image.Pt(0, 0))
	f.tick.Draw(image.Rect(0, ft.Height-f.tickscale*frtickw, f.tickscale*frtickw, ft.Height), f.Cols[colText], nil, image.Pt(0, 0))
}

// SetRects
func (f *Frame) SetRects(r image.Rectangle, b *draw.Image) {
	f.Background = b
	f.Entire = r
	f.Rect = r
	f.Rect.Max.Y -= (r.Max.Y - r.Min.Y) % f.Font.Height
	f.maxlines = (r.Max.Y - r.Min.Y) / f.Font.Height
}

// Clear
func (f *Frame) Clear(freeall bool) {
	if f.nbox != 0 {
		f.delbox(0, f.nbox-1)
	}
	if f.box != nil {
		f.box = nil
	}
	if freeall {
		f.tick.Free()
		f.tickback.Free()
		f.tick = nil
		f.tickback = nil
	}
	f.box = nil
	f.ticked = false
}
