package gobiomes

const (
	mask48 = (1 << 48) - 1
)

// Rng 实现了 Java Random (48-bit LCG) 算法。
type Rng struct {
	seed uint64
}

// NewRng 创建一个新的 Rng 实例。
func NewRng(seed uint64) *Rng {
	r := &Rng{}
	r.SetSeed(seed)
	return r
}

// SetSeed 设置种子。
func (r *Rng) SetSeed(seed uint64) {
	r.seed = (seed ^ 0x5deece66d) & mask48
}

// Next 返回指定位数的随机位。
func (r *Rng) Next(bits int) int32 {
	r.seed = (r.seed*0x5deece66d + 0xb) & mask48
	return int32(int64(r.seed) >> (48 - bits))
}

// NextInt 返回 [0, n) 之间的随机整数。
func (r *Rng) NextInt(n int) int {
	if n&(n-1) == 0 { // n 是 2 的幂
		return int((int64(n) * int64(r.Next(31))) >> 31)
	}

	var bits, val int32
	for {
		bits = r.Next(31)
		val = bits % int32(n)
		if bits-val+(int32(n)-1) >= 0 {
			break
		}
	}
	return int(val)
}

// NextLong 返回随机 int64。
func (r *Rng) NextLong() int64 {
	return (int64(r.Next(32)) << 32) + int64(r.Next(32))
}

// NextFloat 返回 [0.0, 1.0) 之间的随机 float32。
func (r *Rng) NextFloat() float32 {
	return float32(r.Next(24)) / float32(1<<24)
}

// NextDouble 返回 [0.0, 1.0) 之间的随机 float64。
func (r *Rng) NextDouble() float64 {
	return float64((int64(r.Next(26))<<27)+int64(r.Next(27))) / float64(int64(1)<<53)
}

// SkipNextN 模拟调用 n 次 next()。
func (r *Rng) SkipNextN(n uint64) {
	var m uint64 = 1
	var a uint64 = 0
	var im uint64 = 0x5deece66d
	var ia uint64 = 0xb

	for k := n; k > 0; k >>= 1 {
		if k&1 != 0 {
			m *= im
			a = im*a + ia
		}
		ia = (im + 1) * ia
		im *= im
	}

	r.seed = (r.seed*m + a) & mask48
}

// Xoroshiro128 实现了 Minecraft 1.18+ 使用的 Xoroshiro128++ 算法。
type Xoroshiro128 struct {
	lo, hi uint64
}

func rotl(x uint64, k uint) uint64 {
	return (x << k) | (x >> (64 - k))
}

// SetSeed 初始化 Xoroshiro 状态。
func (xr *Xoroshiro128) SetSeed(seed uint64) {
	const (
		xl = 0x9e3779b97f4a7c15
		xh = 0x6a09e667f3bcc909
		a  = 0xbf58476d1ce4e5b9
		b  = 0x94d049bb133111eb
	)
	l := seed ^ xh
	h := l + xl
	l = (l ^ (l >> 30)) * a
	h = (h ^ (h >> 30)) * a
	l = (l ^ (l >> 27)) * b
	h = (h ^ (h >> 27)) * b
	l = l ^ (l >> 31)
	h = h ^ (h >> 31)
	xr.lo = l
	xr.hi = h
}

// NextLong 返回下一个随机 uint64。
func (xr *Xoroshiro128) NextLong() uint64 {
	l := xr.lo
	h := xr.hi
	n := rotl(l+h, 17) + l
	h ^= l
	xr.lo = rotl(l, 49) ^ h ^ (h << 21)
	xr.hi = rotl(h, 28)
	return n
}

// NextInt 返回 [0, n) 之间的随机整数。
func (xr *Xoroshiro128) NextInt(n uint32) int {
	r := (xr.NextLong() & 0xFFFFFFFF) * uint64(n)
	if uint32(r) < n {
		threshold := uint32(-int32(n)) % n
		for uint32(r) < threshold {
			r = (xr.NextLong() & 0xFFFFFFFF) * uint64(n)
		}
	}
	return int(r >> 32)
}

// NextDouble 返回 [0.0, 1.0) 之间的随机 float64。
func (xr *Xoroshiro128) NextDouble() float64 {
	return float64(xr.NextLong()>>(64-53)) * 1.1102230246251565e-16
}

// NextFloat 返回 [0.0, 1.0) 之间的随机 float32。
func (xr *Xoroshiro128) NextFloat() float32 {
	return float32(xr.NextLong()>>(64-24)) * 5.9604645e-8
}

// NextLongJ 模拟 Java 的 nextLong() (两次 32 位采样)。
func (xr *Xoroshiro128) NextLongJ() int64 {
	a := int32(xr.NextLong() >> 32)
	b := int32(xr.NextLong() >> 32)
	return (int64(a) << 32) + int64(b)
}

// NextIntJ 模拟 Java 的 nextInt(n) (基于 31 位采样)。
func (xr *Xoroshiro128) NextIntJ(n uint32) int {
	if n&(n-1) == 0 {
		return int((uint64(n) * (xr.NextLong() >> 33)) >> 31)
	}
	var bits, val uint32
	for {
		bits = uint32(xr.NextLong() >> 33)
		val = bits % n
		if int32(bits-val+(n-1)) >= 0 {
			break
		}
	}
	return int(val)
}

// floorDiv 整数向下取整除法。
func floorDiv(a, b int) int {
	q := a / b
	r := a % b
	if (a^b) < 0 && r != 0 {
		q--
	}
	return q
}

// floorDiv64 int64 向下取整除法。
func floorDiv64(a, b int64) int64 {
	q := a / b
	r := a % b
	if (a^b) < 0 && r != 0 {
		q--
	}
	return q
}

// absInt 整数绝对值。
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// lerp64 线性插值。
func lerp64(part, from, to float64) float64 {
	return from + part*(to-from)
}

// clampedLerp 限制范围的线性插值。
func clampedLerp(part, from, to float64) float64 {
	if part <= 0 {
		return from
	}
	if part >= 1 {
		return to
	}
	return lerp64(part, from, to)
}
