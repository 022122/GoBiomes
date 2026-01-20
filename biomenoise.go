package gobiomes

import (
	"math"
)

// 气候参数索引
const (
	NP_TEMPERATURE     = 0
	NP_HUMIDITY        = 1
	NP_CONTINENTALNESS = 2
	NP_EROSION         = 3
	NP_SHIFT           = 4
	NP_DEPTH           = NP_SHIFT
	NP_WEIRDNESS       = 5
	NP_MAX             = 6
)

// 采样标志
const (
	SAMPLE_NO_SHIFT = 0x1
	SAMPLE_NO_DEPTH = 0x2
	SAMPLE_NO_BIOME = 0x4
)

// Spline 结构体用于 1.18+ 的地形和气候计算
type Spline struct {
	Len int
	Typ int
	Loc [12]float32
	Der [12]float32
	Val []*Spline
	Fix float32
}

type BiomeTree struct {
	Steps []uint32
	Param []int32
	Nodes []uint64
	Order uint32
}

type BiomeNoise struct {
	Climate [NP_MAX]DoublePerlinNoise
	Sp      *Spline
	Mc      int
	NpType  int
}

func (bn *BiomeNoise) Init(mc int) {
	bn.Mc = mc
	bn.NpType = -1
	bn.setupSplines()
}

func (bn *BiomeNoise) setupSplines() {
	// TODO: 实现复杂的样条曲线构建逻辑 (参考 cubiomes/biomenoise.c:1112)
}

func (bn *BiomeNoise) SetSeed(seed uint64, large int) {
	var xr Xoroshiro128
	xr.SetSeed(seed)
	xlo := xr.NextLong()
	xhi := xr.NextLong()

	for i := 0; i < NP_MAX; i++ {
		bn.initClimateSeed(i, uint64(xlo), uint64(xhi), large)
	}
	bn.NpType = -1
}

func (bn *BiomeNoise) initClimateSeed(nptype int, xlo, xhi uint64, large int) {
	var xr Xoroshiro128
	var amp []float64
	var omin, length int

	switch nptype {
	case NP_SHIFT:
		amp = []float64{1, 1, 1, 0}
		xr.lo = xlo ^ 0x080518cf6af25384
		xr.hi = xhi ^ 0x3f3dfb40a54febd5
		omin, length = -3, 4
	case NP_TEMPERATURE:
		amp = []float64{1.5, 0, 1, 0, 0, 0}
		if large != 0 {
			xr.lo = xlo ^ 0x944b0073edf549db
			xr.hi = xhi ^ 0x4ff44347e9d22b96
			omin = -12
		} else {
			xr.lo = xlo ^ 0x5c7e6b29735f0d7f
			xr.hi = xhi ^ 0xf7d86f1bbc734988
			omin = -10
		}
		length = 6
	case NP_HUMIDITY:
		amp = []float64{1, 1, 0, 0, 0, 0}
		if large != 0 {
			xr.lo = xlo ^ 0x71b8ab943dbd5301
			xr.hi = xhi ^ 0xbb63ddcf39ff7a2b
			omin = -10
		} else {
			xr.lo = xlo ^ 0x81bb4d22e8dc168e
			xr.hi = xhi ^ 0xf1c8b4bea16303cd
			omin = -8
		}
		length = 6
	case NP_CONTINENTALNESS:
		amp = []float64{1, 1, 2, 2, 2, 1, 1, 1, 1}
		if large != 0 {
			xr.lo = xlo ^ 0x9a3f51a113fce8dc
			xr.hi = xhi ^ 0xee2dbd157e5dcdad
			omin = -11
		} else {
			xr.lo = xlo ^ 0x83886c9d0ae3a662
			xr.hi = xhi ^ 0xafa638a61b42e8ad
			omin = -9
		}
		length = 9
	case NP_EROSION:
		amp = []float64{1, 1, 0, 1, 1}
		if large != 0 {
			xr.lo = xlo ^ 0x8c984b1f8702a951
			xr.hi = xhi ^ 0xead7b1f92bae535f
			omin = -11
		} else {
			xr.lo = xlo ^ 0xd02491e6058f6fd8
			xr.hi = xhi ^ 0x4792512c94c17a80
			omin = -9
		}
		length = 5
	case NP_WEIRDNESS:
		amp = []float64{1, 2, 1, 0, 0, 0}
		xr.lo = xlo ^ 0xefc8ef4d36102b34
		xr.hi = xhi ^ 0x1beeeb324a0f24ea
		omin, length = -7, 6
	}

	bn.Climate[nptype].InitX(&xr, amp, omin, length)
}

func (bn *BiomeNoise) Sample(x, y, z int, flags uint32) int {
	fx, fz := float64(x), float64(z)

	if flags&SAMPLE_NO_SHIFT == 0 {
		fx += bn.Climate[NP_SHIFT].Sample(float64(x), 0, float64(z)) * 4.0
		fz += bn.Climate[NP_SHIFT].Sample(float64(z), float64(x), 0) * 4.0
	}

	var np [6]uint64
	np[NP_TEMPERATURE] = uint64(int32(bn.Climate[NP_TEMPERATURE].Sample(fx, 0, fz) * 10000.0))
	np[NP_HUMIDITY] = uint64(int32(bn.Climate[NP_HUMIDITY].Sample(fx, 0, fz) * 10000.0))
	np[NP_CONTINENTALNESS] = uint64(int32(bn.Climate[NP_CONTINENTALNESS].Sample(fx, 0, fz) * 10000.0))
	np[NP_EROSION] = uint64(int32(bn.Climate[NP_EROSION].Sample(fx, 0, fz) * 10000.0))
	np[NP_WEIRDNESS] = uint64(int32(bn.Climate[NP_WEIRDNESS].Sample(fx, 0, fz) * 10000.0))

	// Depth calculation
	if flags&SAMPLE_NO_DEPTH == 0 {
		// TODO: Implement spline depth
		np[NP_DEPTH] = uint64(int32((1.0 - float64(y*4)/128.0 - 83.0/160.0) * 10000.0))
	} else {
		np[NP_DEPTH] = 0
	}

	var bt BiomeTree
	if bn.Mc >= MC_1_21 {
		// TODO: Use BTree21
		bt = BiomeTree{Steps: BTree18.Steps, Param: BTree18.Param, Nodes: BTree18.Nodes, Order: BTree18.Order}
	} else {
		bt = BiomeTree{Steps: BTree18.Steps, Param: BTree18.Param, Nodes: BTree18.Nodes, Order: BTree18.Order}
	}

	return ClimateToBiome(&bt, np[:])
}

func GetSplineValue(sp *Spline, vals []float32) float32 {
	if sp == nil || sp.Len == 0 {
		return 0
	}

	if sp.Len == 1 {
		return sp.Fix
	}

	f := vals[sp.Typ]
	i := 0
	for i < sp.Len {
		if sp.Loc[i] >= f {
			break
		}
		i++
	}

	if i == 0 || i == sp.Len {
		if i > 0 {
			i--
		}
		v := GetSplineValue(sp.Val[i], vals)
		return v + sp.Der[i]*(f-sp.Loc[i])
	}

	sp1 := sp.Val[i-1]
	sp2 := sp.Val[i]
	g := sp.Loc[i-1]
	h := sp.Loc[i]
	k := (f - g) / (h - g)
	l := sp.Der[i-1]
	m := sp.Der[i]
	n := GetSplineValue(sp1, vals)
	o := GetSplineValue(sp2, vals)
	p := l*(h-g) - (o - n)
	q := -m*(h-g) + (o - n)

	// 埃尔米特插值 (Hermite interpolation)
	r := lerp32(k, n, o) + k*(1.0-k)*lerp32(k, p, q)
	return r
}

func lerp32(t, a, b float32) float32 {
	return a + t*(b-a)
}

// BiomeTree 遍历逻辑
func get_np_dist(np []uint64, bt *BiomeTree, nodeIdx int) uint64 {
	var ds uint64 = 0
	node := bt.Nodes[nodeIdx]

	for i := 0; i < 6; i++ {
		paramIdx := int((node >> (8 * i)) & 0xFF)
		minVal := uint64(bt.Param[2*paramIdx+0])
		maxVal := uint64(bt.Param[2*paramIdx+1])

		val := np[i]
		var d uint64
		if val > maxVal {
			d = val - maxVal
		} else if val < minVal {
			d = minVal - val
		} else {
			d = 0
		}
		ds += d * d
	}
	return ds
}

func get_resulting_node(np []uint64, bt *BiomeTree, idx int, alt int, ds uint64, depth int) int {
	if bt.Steps[depth] == 0 {
		return idx
	}

	step := bt.Steps[depth]
	for idx+int(step) >= len(bt.Nodes) {
		depth++
		step = bt.Steps[depth]
	}

	node := bt.Nodes[idx]
	inner := int(int16(node >> 48))
	if inner < 0 {
		return idx
	}

	leaf := alt
	n := int(bt.Order)

	for i := 0; i < n; i++ {
		if inner >= len(bt.Nodes) {
			break
		}
		ds_inner := get_np_dist(np, bt, inner)
		if ds_inner < ds {
			leaf2 := get_resulting_node(np, bt, inner, leaf, ds, depth+1)
			var ds_leaf2 uint64
			if inner == leaf2 {
				ds_leaf2 = ds_inner
			} else {
				ds_leaf2 = get_np_dist(np, bt, leaf2)
			}

			if ds_leaf2 < ds {
				ds = ds_leaf2
				leaf = leaf2
			}
		}

		inner += int(step)
		if inner >= len(bt.Nodes) {
			break
		}
	}

	return leaf
}

func ClimateToBiome(bt *BiomeTree, np []uint64) int {
	idx := get_resulting_node(np, bt, 0, 0, math.MaxUint64, 0)
	return int((bt.Nodes[idx] >> 48) & 0xFF)
}
