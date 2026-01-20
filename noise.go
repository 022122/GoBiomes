package gobiomes

import (
	"math"
)

// PerlinNoise 对应 cubiomes 的 PerlinNoise。
type PerlinNoise struct {
	d          [257]uint8
	h2         uint8
	a, b, c    float64
	amplitude  float64
	lacunarity float64
	d2         float64
	t2         float64
}

// OctaveNoise 对应 cubiomes 的 OctaveNoise。
type OctaveNoise struct {
	Octaves []PerlinNoise
}

// DoublePerlinNoise 对应 cubiomes 的 DoublePerlinNoise。
type DoublePerlinNoise struct {
	Amplitude float64
	OctA      OctaveNoise
	OctB      OctaveNoise
}

func indexedLerp(idx uint8, a, b, c float64) float64 {
	switch idx & 0xf {
	case 0:
		return a + b
	case 1:
		return -a + b
	case 2:
		return a - b
	case 3:
		return -a - b
	case 4:
		return a + c
	case 5:
		return -a + c
	case 6:
		return a - c
	case 7:
		return -a - c
	case 8:
		return b + c
	case 9:
		return -b + c
	case 10:
		return b - c
	case 11:
		return -b - c
	case 12:
		return a + b
	case 13:
		return -b + c
	case 14:
		return -a + b
	case 15:
		return -b - c
	}
	return 0
}

func (p *PerlinNoise) Init(r *Rng) {
	p.a = r.NextDouble() * 256.0
	p.b = r.NextDouble() * 256.0
	p.c = r.NextDouble() * 256.0
	p.amplitude = 1.0
	p.lacunarity = 1.0

	for i := 0; i < 256; i++ {
		p.d[i] = uint8(i)
	}
	for i := 0; i < 256; i++ {
		j := r.NextInt(256-i) + i
		p.d[i], p.d[j] = p.d[j], p.d[i]
	}
	p.d[256] = p.d[0]

	i2 := math.Floor(p.b)
	d2 := p.b - i2
	p.h2 = uint8(int(i2) & 0xff)
	p.d2 = d2
	p.t2 = d2 * d2 * d2 * (d2*(d2*6.0-15.0) + 10.0)
}

func (p *PerlinNoise) InitX(xr *Xoroshiro128) {
	p.a = xr.NextDouble() * 256.0
	p.b = xr.NextDouble() * 256.0
	p.c = xr.NextDouble() * 256.0
	p.amplitude = 1.0
	p.lacunarity = 1.0

	for i := 0; i < 256; i++ {
		p.d[i] = uint8(i)
	}
	for i := 0; i < 256; i++ {
		j := xr.NextInt(uint32(256-i)) + i
		p.d[i], p.d[j] = p.d[j], p.d[i]
	}
	p.d[256] = p.d[0]

	i2 := math.Floor(p.b)
	d2 := p.b - i2
	p.h2 = uint8(int(i2) & 0xff)
	p.d2 = d2
	p.t2 = d2 * d2 * d2 * (d2*(d2*6.0-15.0) + 10.0)
}

func maintainPrecision(d float64) float64 {
	if d > 1e6 || d < -1e6 {
		return d - math.Floor(d/16777216.0)*16777216.0
	}
	return d
}

func (p *PerlinNoise) Sample(d1, d2, d3 float64, yamp, ymin float64) float64 {
	var h1, h2, h3 uint8
	var t1, t2, t3 float64

	if d2 == 0.0 {
		d2 = p.d2
		h2 = p.h2
		t2 = p.t2
	} else {
		d2 += p.b
		i2 := math.Floor(d2)
		d2 -= i2
		h2 = uint8(int(i2) & 0xff)
		t2 = d2 * d2 * d2 * (d2*(d2*6.0-15.0) + 10.0)
	}

	d1 += p.a
	d3 += p.c

	i1 := math.Floor(d1)
	i3 := math.Floor(d3)
	d1 -= i1
	d3 -= i3

	h1 = uint8(int(i1) & 0xff)
	h3 = uint8(int(i3) & 0xff)

	t1 = d1 * d1 * d1 * (d1*(d1*6.0-15.0) + 10.0)
	t3 = d3 * d3 * d3 * (d3*(d3*6.0-15.0) + 10.0)

	if yamp != 0 {
		yclamp := d2
		if ymin < yclamp {
			yclamp = ymin
		}
		d2 -= math.Floor(yclamp/yamp) * yamp
	}

	idx := p.d[:]

	a1 := idx[h1] + h2
	b1 := idx[h1+1] + h2

	a2 := idx[a1] + h3
	b2 := idx[b1] + h3
	a3 := idx[a1+1] + h3
	b3 := idx[b1+1] + h3

	l1 := indexedLerp(idx[a2], d1, d2, d3)
	l2 := indexedLerp(idx[b2], d1-1, d2, d3)
	l3 := indexedLerp(idx[a3], d1, d2-1, d3)
	l4 := indexedLerp(idx[b3], d1-1, d2-1, d3)
	l5 := indexedLerp(idx[a2+1], d1, d2, d3-1)
	l6 := indexedLerp(idx[b2+1], d1-1, d2, d3-1)
	l7 := indexedLerp(idx[a3+1], d1, d2-1, d3-1)
	l8 := indexedLerp(idx[b3+1], d1-1, d2-1, d3-1)

	l1 = lerp64(t1, l1, l2)
	l3 = lerp64(t1, l3, l4)
	l5 = lerp64(t1, l5, l6)
	l7 = lerp64(t1, l7, l8)

	l1 = lerp64(t2, l1, l3)
	l5 = lerp64(t2, l5, l7)

	return lerp64(t3, l1, l5)
}

func (o *OctaveNoise) Init(r *Rng, omin, len int) {
	end := omin + len - 1
	persist := 1.0 / float64((int64(1)<<len)-1)
	lacuna := math.Pow(2.0, float64(end))

	o.Octaves = make([]PerlinNoise, len)
	i := 0
	if end == 0 {
		o.Octaves[0].Init(r)
		o.Octaves[0].amplitude = persist
		o.Octaves[0].lacunarity = lacuna
		persist *= 2.0
		lacuna *= 0.5
		i = 1
	} else {
		r.SkipNextN(uint64(-end * 262))
	}

	for ; i < len; i++ {
		o.Octaves[i].Init(r)
		o.Octaves[i].amplitude = persist
		o.Octaves[i].lacunarity = lacuna
		persist *= 2.0
		lacuna *= 0.5
	}
}

func (o *OctaveNoise) InitX(xr *Xoroshiro128, amplitudes []float64, omin, len int) {
	var md5_octave_n = [13][2]uint64{
		{0xb198de63a8012672, 0x7b84cad43ef7b5a8},
		{0x0fd787bfbc403ec3, 0x74a4a31ca21b48b8},
		{0x36d326eed40efeb2, 0x5be9ce18223c636a},
		{0x082fe255f8be6631, 0x4e96119e22dedc81},
		{0x0ef68ec68504005e, 0x48b6bf93a2789640},
		{0xf11268128982754f, 0x257a1d670430b0aa},
		{0xe51c98ce7d1de664, 0x5f9478a733040c45},
		{0x6d7b49e7e429850a, 0x2e3063c622a24777},
		{0xbd90d5377ba1b762, 0xc07317d419a7548d},
		{0x53d39c6752dac858, 0xbcd1c5a80ab65b3e},
		{0xb4a24d7a84e7677b, 0x023ff9668e89b5c4},
		{0xdffa22b534c5f608, 0xb9b67517d3665ca9},
		{0xd50708086cef4d7c, 0x6e1651ecc7f43309},
	}
	var lacuna_ini = []float64{
		1, .5, .25, 1. / 8, 1. / 16, 1. / 32, 1. / 64, 1. / 128, 1. / 256, 1. / 512, 1. / 1024,
		1. / 2048, 1. / 4096,
	}
	var persist_ini = []float64{
		0, 1, 2. / 3, 4. / 7, 8. / 15, 16. / 31, 32. / 63, 64. / 127, 128. / 255, 256. / 511,
	}

	lacuna := lacuna_ini[-omin]
	persist := persist_ini[len]
	xlo := xr.NextLong()
	xhi := xr.NextLong()

	o.Octaves = nil
	for i := 0; i < len; i++ {
		if amplitudes[i] != 0 {
			var pxr Xoroshiro128
			pxr.lo = xlo ^ md5_octave_n[12+omin+i][0]
			pxr.hi = xhi ^ md5_octave_n[12+omin+i][1]
			var p PerlinNoise
			p.InitX(&pxr)
			p.amplitude = amplitudes[i] * persist
			p.lacunarity = lacuna
			o.Octaves = append(o.Octaves, p)
		}
		lacuna *= 2.0
		persist *= 0.5
	}
}

func (o *OctaveNoise) Sample(x, y, z float64) float64 {
	v := 0.0
	for i := range o.Octaves {
		p := &o.Octaves[i]
		lf := p.lacunarity
		v += p.amplitude * p.Sample(maintainPrecision(x*lf), maintainPrecision(y*lf), maintainPrecision(z*lf), 0, 0)
	}
	return v
}

func (d *DoublePerlinNoise) Init(r *Rng, omin, len int) {
	d.Amplitude = (10.0 / 6.0) * float64(len) / float64(len+1)
	d.OctA.Init(r, omin, len)
	d.OctB.Init(r, omin, len)
}

func (d *DoublePerlinNoise) InitX(xr *Xoroshiro128, amplitudes []float64, omin, len int) {
	d.OctA.InitX(xr, amplitudes, omin, len)
	d.OctB.InitX(xr, amplitudes, omin, len)

	// trim amplitudes
	actualLen := len
	for i := len - 1; i >= 0 && amplitudes[i] == 0.0; i-- {
		actualLen--
	}
	start := 0
	for i := 0; i < actualLen && amplitudes[i] == 0.0; i++ {
		start++
	}
	actualLen -= start

	var amp_ini = []float64{
		0, 5. / 6, 10. / 9, 15. / 12, 20. / 15, 25. / 18, 30. / 21, 35. / 24, 40. / 27, 45. / 30,
	}
	if actualLen < 10 {
		d.Amplitude = amp_ini[actualLen]
	}
}

func (d *DoublePerlinNoise) Sample(x, y, z float64) float64 {
	const f = 337.0 / 331.0
	v := d.OctA.Sample(x, y, z)
	v += d.OctB.Sample(x*f, y*f, z*f)
	return v * d.Amplitude
}
