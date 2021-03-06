package frame

import (
	"image"
	"unicode/utf8"
)

func (f *Frame) canfit(pt image.Point, b *frbox) (int, bool) {
	left := f.Rect.Max.X - pt.X
	if b.Nrune < 0 {
		if b.Minwid <= byte(left) {
			return 1, true
		} else {
			return 0, false
		}
	}

	if left > b.Wid {
		return b.Nrune, (b.Nrune != 0)
	}

	w := 0
	for nr := 0; nr < len(b.Ptr); nr += w {
		_, w = utf8.DecodeRune(b.Ptr[nr:])
		left -= f.Font.StringWidth(string(b.Ptr[nr : nr+1]))
		if left < 0 {
			return nr, nr != 0
		}
	}
	return 0, false
}

func (f *Frame) cklinewrap(p *image.Point, b *frbox) {
	if b.Nrune < 0 {
		if b.Minwid > byte(f.Rect.Max.X-p.X) {
			p.X = f.Rect.Min.X
			p.Y += f.Font.Height
		}
	} else {
		if b.Wid > f.Rect.Max.X-p.X {
			p.X = f.Rect.Min.X
			p.Y += f.Font.Height
		}
	}
}

func (f *Frame) cklinewrap0(p *image.Point, b *frbox) {
	if _, ok := f.canfit(*p, b); !ok {
		p.X = f.Rect.Min.X
		p.Y += f.Font.Height
	}
}

func (f *Frame) advance(p *image.Point, b *frbox) {
	if b.Nrune < 0 && b.Bc == '\n' {
		p.X = f.Rect.Min.X
		p.Y += f.Font.Height
	} else {
		p.X += b.Wid
	}
}

func (f *Frame) newwid(pt image.Point, b *frbox) int {
	b.Wid = f.newwid0(pt, b)
	return b.Wid
}

func (f *Frame) newwid0(pt image.Point, b *frbox) int {
	c := f.Rect.Max.X
	x := pt.X
	if b.Nrune >= 0 || b.Bc != '\t' {
		return b.Wid
	}
	if x+int(b.Minwid) > c {
		pt.X = f.Rect.Min.X
		x = pt.X
	}
	x += f.maxtab
	x -= (x - f.Rect.Min.X) % f.maxtab
	if x-pt.X < int(b.Minwid) || x > c {
		x = pt.X + int(b.Minwid)
	}
	return x - pt.X
}

func (f *Frame) clean(pt image.Point, n0, n1 int) {
	c := f.Rect.Max.X
	nb := 0
	for nb = n0; nb < n1-1; nb++ {
		b := f.box[nb]
		f.cklinewrap(&pt, b)
		for f.box[nb].Nrune >= 0 &&
			nb < n1-1 &&
			f.box[nb+1].Nrune >= 0 &&
			pt.X+f.box[nb].Wid+f.box[nb+1].Wid < c {
			f.mergebox(nb)
			n1--
			b = f.box[nb]
		}
		f.advance(&pt, f.box[nb])
	}

	for ; nb < f.nbox; nb++ {
		b := f.box[nb]
		f.cklinewrap(&pt, b)
		f.advance(&pt, f.box[nb])
	}
	f.lastlinefull = 0
	if pt.Y >= f.Rect.Max.Y {
		f.lastlinefull = 1
	}
}

func nbyte(f *frbox) uint {
	if f.Nrune < 0 {
		return 1 // treat as single break character
	} else {
		return uint(f.Nrune)
	}
}

func nrune(f *frbox) int {
	return len(f.Ptr)
}

func Rpt(min, max image.Point) image.Rectangle {
	return image.Rectangle{Min: min, Max: max}
}
