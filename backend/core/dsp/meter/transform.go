package meter

// Transform is an ideal binary64 real-DFT reference in pocketfft's operation order, not the native binary32 butterflies; one caller at a time.
type Transform struct {
	size     int
	factors  []int
	twiddles [][]float64 // per factor, (ip-1) rows of ido holding pair i at [i-1] and [i], like the data
	work     []float64
	scratch  []float64
}

// NewTransform panics unless size is a power of two in 4..2097152.
func NewTransform(size int) *Transform {
	transformCheckSize(size)
	t := &Transform{size: size, work: make([]float64, size), scratch: make([]float64, size)}
	n := size
	for ; n%4 == 0; n /= 4 {
		t.factors = append(t.factors, 4)
	}
	if n == 2 {
		t.factors = append([]int{2}, t.factors...)
	}
	trig := newTransformTrig(size)
	t.twiddles = make([][]float64, len(t.factors))
	l1 := 1
	for k, ip := range t.factors[:len(t.factors)-1] {
		ido := size / (l1 * ip)
		tw := make([]float64, (ip-1)*ido)
		for j := 1; j < ip; j++ {
			for m := 1; m <= (ido-1)/2; m++ {
				tw[(j-1)*ido+2*m-1], tw[(j-1)*ido+2*m] = trig.at(j * l1 * m)
			}
		}
		t.twiddles[k] = tw
		l1 *= ip
	}
	return t
}

func (t *Transform) Size() int { return t.size }

func transformCheckSize(size int) {
	if size < 4 || size > 1<<21 || size&(size-1) != 0 {
		panic("meter: transform size must be a power of two in 4..2097152")
	}
}

// Packed writes and returns the rfft rounded to binary32 as [DC, Nyquist, re1, im1, ...] with a positive imaginary sign; short src is zero-padded.
func (t *Transform) Packed(dst, src []float32) []float32 {
	n := t.size
	dst = dst[:n]
	for j, v := range src {
		t.work[j] = float64(v)
	}
	clear(t.work[len(src):])
	p1, p2 := t.work, t.scratch
	l1 := n
	for k := len(t.factors) - 1; k >= 0; k-- {
		ip := t.factors[k]
		ido := n / l1
		l1 /= ip
		if ip == 4 {
			transformRadf4(ido, l1, p1, p2, t.twiddles[k])
		} else {
			transformRadf2(ido, l1, p1, p2, t.twiddles[k])
		}
		p1, p2 = p2, p1
	}
	dst[0], dst[1] = float32(p1[0]), float32(p1[n-1])
	for k := 1; k < n/2; k++ {
		dst[2*k], dst[2*k+1] = float32(p1[2*k-1]), float32(-p1[2*k])
	}
	return dst
}

func transformRadf2(ido, l1 int, cc, ch, wa []float64) {
	for k := range l1 {
		a, b := cc[ido*k], cc[ido*(k+l1)]
		ch[2*ido*k], ch[ido-1+ido*(1+2*k)] = a+b, a-b
	}
	if ido&1 == 0 {
		for k := range l1 {
			ch[ido*(1+2*k)] = -cc[ido-1+ido*(k+l1)]
			ch[ido-1+2*ido*k] = cc[ido-1+ido*k]
		}
	}
	w := wa[:ido]
	for k := range l1 {
		c0, c1 := cc[ido*k:][:ido], cc[ido*(k+l1):][:ido]
		h0, h1 := ch[2*ido*k:][:ido], ch[ido*(1+2*k):][:ido]
		for i, ic := 2, ido-2; ic > 0; i, ic = i+2, ic-2 {
			tr2 := float64(w[i-1]*c1[i-1]) + float64(w[i]*c1[i])
			ti2 := float64(w[i-1]*c1[i]) - float64(w[i]*c1[i-1])
			h0[i-1], h1[ic-1] = c0[i-1]+tr2, c0[i-1]-tr2
			h0[i], h1[ic] = ti2+c0[i], ti2-c0[i]
		}
	}
}

func transformRadf4(ido, l1 int, cc, ch, wa []float64) {
	const hsqt2 = 0.707106781186547524400844362104849
	for k := range l1 {
		c0, c1, c2, c3 := cc[ido*k], cc[ido*(k+l1)], cc[ido*(k+2*l1)], cc[ido*(k+3*l1)]
		tr1, tr2 := c3+c1, c0+c2
		ch[ido*(2+4*k)], ch[ido-1+ido*(1+4*k)] = c3-c1, c0-c2
		ch[4*ido*k], ch[ido-1+ido*(3+4*k)] = tr2+tr1, tr2-tr1
	}
	if ido&1 == 0 {
		for k := range l1 {
			e := ido - 1
			c0, c1, c2, c3 := cc[e+ido*k], cc[e+ido*(k+l1)], cc[e+ido*(k+2*l1)], cc[e+ido*(k+3*l1)]
			ti1 := float64(-hsqt2 * (c1 + c3))
			tr1 := float64(hsqt2 * (c1 - c3))
			ch[e+4*ido*k], ch[e+ido*(2+4*k)] = c0+tr1, c0-tr1
			ch[ido*(3+4*k)], ch[ido*(1+4*k)] = ti1+c2, ti1-c2
		}
	}
	if ido <= 2 {
		return
	}
	w1, w2, w3 := wa[:ido], wa[ido:][:ido], wa[2*ido:][:ido]
	for k := range l1 {
		c0, c1 := cc[ido*k:][:ido], cc[ido*(k+l1):][:ido]
		c2, c3 := cc[ido*(k+2*l1):][:ido], cc[ido*(k+3*l1):][:ido]
		h := ch[4*ido*k:][:4*ido]
		h0, h1, h2, h3 := h[:ido], h[ido:][:ido], h[2*ido:][:ido], h[3*ido:][:ido]
		for i, ic := 2, ido-2; ic > 0; i, ic = i+2, ic-2 {
			cr2 := float64(w1[i-1]*c1[i-1]) + float64(w1[i]*c1[i])
			ci2 := float64(w1[i-1]*c1[i]) - float64(w1[i]*c1[i-1])
			cr3 := float64(w2[i-1]*c2[i-1]) + float64(w2[i]*c2[i])
			ci3 := float64(w2[i-1]*c2[i]) - float64(w2[i]*c2[i-1])
			cr4 := float64(w3[i-1]*c3[i-1]) + float64(w3[i]*c3[i])
			ci4 := float64(w3[i-1]*c3[i]) - float64(w3[i]*c3[i-1])
			tr1, tr4 := cr4+cr2, cr4-cr2
			ti1, ti4 := ci2+ci4, ci2-ci4
			tr2, tr3 := c0[i-1]+cr3, c0[i-1]-cr3
			ti2, ti3 := c0[i]+ci3, c0[i]-ci3
			h0[i-1], h3[ic-1] = tr2+tr1, tr2-tr1
			h0[i], h3[ic] = ti1+ti2, ti1-ti2
			h2[i-1], h1[ic-1] = tr3+ti4, tr3-ti4
			h2[i], h1[ic] = tr4+ti3, tr4-ti3
		}
	}
}
