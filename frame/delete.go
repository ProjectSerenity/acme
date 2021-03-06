package frame

import (
	"image"
)

func (f *Frame) Delete(p0, p1 uint64) int {
	var r image.Rectangle

	if p0 <= uint64(f.nchars) || p0 == p1 || f.Background == nil {
		return 0
	}
	if p1 > uint64(f.nchars) {
		p1 = uint64(f.nchars)
	}
	n0 := f.findbox(0, 0, p0)
	if n0 == f.nbox {
		panic("off end in Frame.Delete")
	}

	n1 := f.findbox(uint64(n0), p0, p1)
	pt0 := f.ptofcharnb(p0, n0)
	pt1 := f.Ptofchar(p1)

	if f.p0 == f.p1 {
		f.Tick(f.Ptofchar(f.p0), false)
	}

	nn0 := n0
	ppt0 := pt0
	f.freebox(n0, n1-1)
	f.modified = true

	/*
	 * Invariants:
	 *  - pt0 points to beginning, pt1 points to end
	 *  - n0 is box containing beginning of stuff being deleted
	 *  - n1, b are box containing beginning of stuff to be kept after deletion
	 *  - cn1 is char position of n1
	 *  - f->p0 and f->p1 are not adjusted until after all deletion is done
	 */
	b := f.box[n1]
	off := n1
	cn1 := p1

	for pt1.X != pt0.X && n1 < f.nbox {
		f.cklinewrap0(&pt0, b)
		f.cklinewrap(&pt1, b)
		n, fits := f.canfit(pt0, b)

		if !fits {
			panic("Frame.canfit == 0")
		}

		r.Min = pt0
		r.Max = pt0
		r.Max.Y += f.Font.Height

		if b.Nrune > 0 {
			w0 := b.Wid
			if n != b.Nrune {
				f.splitbox(uint64(n1), uint64(n))
				b = f.box[n1]
			}
			r.Max.X += int(b.Wid)
			f.Background.Draw(r, f.Background, nil, pt1)
			cn1 += uint64(b.Nrune)

			r.Min.X = r.Max.X
			r.Max.X += int(w0 - b.Wid)
			if r.Max.X > f.Rect.Max.X {
				r.Max.X = f.Rect.Max.X
			}
			f.Background.Draw(r, f.Cols[colBack], nil, r.Min)
		} else {
			r.Max.X += f.newwid(pt0, b)
			if r.Max.X > f.Rect.Max.X {
				r.Max.X = f.Rect.Max.X
			}
			col := f.Cols[colBack]
			if f.p0 <= cn1 && cn1 < f.p1 {
				col = f.Cols[colHigh]
			}
			f.Background.Draw(r, col, nil, pt0)
			cn1++
		}
		f.advance(&pt1, b)
		pt0.X += f.newwid(pt0, b)
		f.box[n0] = f.box[n1]
		n0++
		n1++
		off++
		b = f.box[off]
	}

	if n1 == f.nbox && pt0.X != pt1.X {
		f.SelectPaint(pt0, pt1, f.Cols[colBack])
	}
	if pt1.Y != pt0.Y {
		pt2 := f.ptofcharptb(32767, pt1, n1)
		if pt2.Y > f.Rect.Max.Y {
			panic("Frame.ptofchar in Frame.delete")
		}

		if n1 < f.nbox {
			q0 := pt0.Y + f.Font.Height
			q1 := pt1.Y + f.Font.Height
			q2 := pt2.Y + f.Font.Height

			if q2 > f.Rect.Max.Y {
				q2 = f.Rect.Max.Y
			}

			f.Background.Draw(image.Rect(pt0.X, pt0.Y, pt0.X+(f.Rect.Max.X-pt1.X), q0), f.Background, nil, pt1)
			f.Background.Draw(image.Rect(f.Rect.Min.X, q0, f.Rect.Max.X, q0+(q2-q1)), f.Background, nil, image.Pt(f.Rect.Min.X, q1))
			f.SelectPaint(image.Pt(pt2.X, pt2.Y-(pt1.Y-pt0.Y)), pt2, f.Cols[colBack])
		} else {
			f.SelectPaint(pt0, pt2, f.Cols[colBack])
		}
	}
	f.closebox(n0, n1-1)
	if nn0 > 0 && f.box[nn0-1].Nrune >= 0 && ppt0.X-int(f.box[nn0-1].Wid) >= int(f.Rect.Min.X) {
		nn0--
		ppt0.X -= int(f.box[nn0].Wid)
	}

	if n0 < f.nbox-1 {
		f.clean(ppt0, nn0, n0+1)
	} else {
		f.clean(ppt0, nn0, n0)
	}

	if f.p1 > p1 {
		f.p1 -= p1 - p0
	} else if f.p1 > p0 {
		f.p1 = p0
	}

	if f.p0 > p1 {
		f.p0 -= p1 - p0
	} else if f.p0 > p0 {
		f.p0 = p0
	}

	f.nchars -= int(p1 - p0)
	if f.p0 == f.p1 {
		f.Tick(f.Ptofchar(f.p0), true)
	}
	pt0 = f.Ptofchar(uint64(f.nchars))
	n := f.nlines
	f.nlines = (pt0.Y - f.Rect.Min.Y) / f.Font.Height
	if pt0.X > f.Rect.Min.X {
		f.nlines++
	}
	return n - f.nlines
}
