package gobiomes

import "math"

// Layer 接口定义了生物群系生成层的行为
type Layer struct {
	GetMap    func(l *Layer, out []int, x, z, w, h int) int
	MC        int
	Zoom      int
	Edge      int
	Scale     int
	LayerSalt uint64
	StartSalt uint64
	StartSeed uint64
	Noise     *PerlinNoise
	P         *Layer
	P2        *Layer
}

// LayerStack 存储完整的层级链
type LayerId int

const (
	L_CONTINENT_4096 LayerId = iota
	L_ZOOM_4096
	L_LAND_4096
	L_ZOOM_2048
	L_LAND_2048
	L_ZOOM_1024
	L_LAND_1024_A
	L_LAND_1024_B
	L_LAND_1024_C
	L_ISLAND_1024
	L_SNOW_1024
	L_LAND_1024_D
	L_COOL_1024
	L_HEAT_1024
	L_SPECIAL_1024
	L_ZOOM_512
	L_LAND_512
	L_ZOOM_256
	L_LAND_256
	L_MUSHROOM_256
	L_DEEP_OCEAN_256
	L_BIOME_256
	L_BAMBOO_256
	L_ZOOM_128
	L_ZOOM_64
	L_BIOME_EDGE_64
	L_RIVER_INIT_256
	L_ZOOM_128_HILLS
	L_ZOOM_64_HILLS
	L_HILLS_64
	L_SUNFLOWER_64
	L_ZOOM_32
	L_LAND_32
	L_ZOOM_16
	L_SHORE_16
	L_SWAMP_RIVER_16
	L_ZOOM_8
	L_ZOOM_4
	L_SMOOTH_4
	L_ZOOM_128_RIVER
	L_ZOOM_64_RIVER
	L_ZOOM_32_RIVER
	L_ZOOM_16_RIVER
	L_ZOOM_8_RIVER
	L_ZOOM_4_RIVER
	L_RIVER_4
	L_SMOOTH_4_RIVER
	L_RIVER_MIX_4
	L_OCEAN_TEMP_256
	L_ZOOM_128_OCEAN
	L_ZOOM_64_OCEAN
	L_ZOOM_32_OCEAN
	L_ZOOM_16_OCEAN
	L_ZOOM_8_OCEAN
	L_ZOOM_4_OCEAN
	L_OCEAN_MIX_4
	L_VORONOI_1
	L_ZOOM_LARGE_A
	L_ZOOM_LARGE_B
	L_ZOOM_L_RIVER_A
	L_ZOOM_L_RIVER_B
	L_NUM
)

type LayerStack struct {
	Layers   [L_NUM]Layer
	Entry1   *Layer
	Entry4   *Layer
	Entry16  *Layer
	Entry64  *Layer
	Entry256 *Layer
	OceanRnd PerlinNoise
}

// mcStepSeed 对应 cubiomes 的 mcStepSeed
func mcStepSeed(st, ls uint64) uint64 {
	return st*(st*6364136223846793005+1442695040888963407) + ls
}

// getLayerSalt 对应 cubiomes 的 getLayerSalt
func getLayerSalt(salt uint64) uint64 {
	ls := mcStepSeed(salt, salt)
	ls = mcStepSeed(ls, salt)
	ls = mcStepSeed(ls, salt)
	return ls
}

// getChunkSeed 对应 cubiomes 的 getChunkSeed
func getChunkSeed(ss uint64, x, z int) uint64 {
	cs := ss + uint64(x)
	cs = mcStepSeed(cs, uint64(z))
	cs = mcStepSeed(cs, uint64(x))
	cs = mcStepSeed(cs, uint64(z))
	return cs
}

// mcFirstIsZero 对应 cubiomes 的 mcFirstIsZero
func mcFirstIsZero(cs uint64, mod int) int {
	if mod <= 0 {
		return 0
	}
	res := int((cs >> 24) % uint64(mod))
	if res == 0 {
		return 1
	}
	return 0
}

// mcFirstInt 对应 cubiomes 的 mcFirstInt
func mcFirstInt(cs uint64, mod int) int {
	if mod <= 0 {
		return 0
	}
	return int((cs >> 24) % uint64(mod))
}

// SetLayerSeed 应用世界种子到层级
func SetLayerSeed(l *Layer, worldSeed uint64) {
	if l.P2 != nil {
		SetLayerSeed(l.P2, worldSeed)
	}
	if l.P != nil {
		SetLayerSeed(l.P, worldSeed)
	}

	if l.Noise != nil {
		var s uint64
		// 简单的 LCG 初始化
		s = (worldSeed ^ 0x5deece66d) & mask48
		l.Noise.Init(&Rng{seed: s})
	}

	ls := l.LayerSalt
	if ls == 0 {
		l.StartSalt = 0
		l.StartSeed = 0
	} else {
		st := worldSeed
		st = mcStepSeed(st, ls)
		st = mcStepSeed(st, ls)
		st = mcStepSeed(st, ls)

		l.StartSalt = st
		l.StartSeed = mcStepSeed(st, 0)
	}
}

// MapContinent 对应 mapContinent 层
func MapContinent(l *Layer, out []int, x, z, w, h int) int {
	ss := l.StartSeed
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			cs := getChunkSeed(ss, i+x, j+z)
			out[j*w+i] = mcFirstIsZero(cs, 10)
		}
	}
	if x > -w && x <= 0 && z > -h && z <= 0 {
		out[-z*w-x] = 1
	}
	return 0
}

func select4(cs, st uint32, v00, v01, v10, v11 int) int {
	cv00 := 0
	if v00 == v10 {
		cv00++
	}
	if v00 == v01 {
		cv00++
	}
	if v00 == v11 {
		cv00++
	}

	cv10 := 0
	if v10 == v01 {
		cv10++
	}
	if v10 == v11 {
		cv10++
	}

	cv01 := 0
	if v01 == v11 {
		cv01++
	}

	if cv00 > cv10 && cv00 > cv01 {
		return v00
	} else if cv10 > cv00 {
		return v10
	} else if cv01 > cv00 {
		return v01
	} else {
		cs *= cs*1284865837 + 4150755663
		cs += st
		r := (cs >> 24) & 3
		switch r {
		case 0:
			return v00
		case 1:
			return v10
		case 2:
			return v01
		default:
			return v11
		}
	}
}

// MapZoomFuzzy 对应 mapZoomFuzzy 层
func MapZoomFuzzy(l *Layer, out []int, x, z, w, h int) int {
	pX, pZ := x>>1, z>>1
	pW := ((x + w) >> 1) - pX + 1
	pH := ((z + h) >> 1) - pZ + 1

	parentOut := make([]int, pW*pH)
	err := l.P.GetMap(l.P, parentOut, pX, pZ, pW, pH)
	if err != 0 {
		return err
	}

	newW := pW * 2
	buf := make([]int, (pW*2)*(pH*2))

	st := uint32(l.StartSalt)
	ss := uint32(l.StartSeed)

	for j := 0; j < pH-1; j++ {
		for i := 0; i < pW-1; i++ {
			v00 := parentOut[j*pW+i]
			v10 := parentOut[j*pW+i+1]
			v01 := parentOut[(j+1)*pW+i]
			v11 := parentOut[(j+1)*pW+i+1]

			chunkX, chunkZ := (i+pX)*2, (j+pZ)*2

			cs := ss
			cs += uint32(chunkX)
			cs *= cs*1284865837 + 4150755663
			cs += uint32(chunkZ)
			cs *= cs*1284865837 + 4150755663
			cs += uint32(chunkX)
			cs *= cs*1284865837 + 4150755663
			cs += uint32(chunkZ)

			buf[(j*2)*newW+(i*2)] = v00

			if (cs>>24)&1 != 0 {
				buf[(j*2+1)*newW+(i*2)] = v01
			} else {
				buf[(j*2+1)*newW+(i*2)] = v00
			}

			cs *= cs*1284865837 + 4150755663
			cs += st
			if (cs>>24)&1 != 0 {
				buf[(j*2)*newW+(i*2+1)] = v10
			} else {
				buf[(j*2)*newW+(i*2+1)] = v00
			}

			cs *= cs*1284865837 + 4150755663
			cs += st
			r := (cs >> 24) & 3
			var res int
			if r == 0 {
				res = v00
			} else if r == 1 {
				res = v10
			} else if r == 2 {
				res = v01
			} else {
				res = v11
			}
			buf[(j*2+1)*newW+(i*2+1)] = res
		}
	}

	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			out[j*w+i] = buf[(j+(z&1))*newW+(i+(x&1))]
		}
	}

	return 0
}

// MapZoom 对应 mapZoom 层
func MapZoom(l *Layer, out []int, x, z, w, h int) int {
	pX, pZ := x>>1, z>>1
	pW := ((x + w) >> 1) - pX + 1
	pH := ((z + h) >> 1) - pZ + 1

	parentOut := make([]int, pW*pH)
	err := l.P.GetMap(l.P, parentOut, pX, pZ, pW, pH)
	if err != 0 {
		return err
	}

	newW := pW * 2
	buf := make([]int, (pW*2)*(pH*2))

	st := uint32(l.StartSalt)
	ss := uint32(l.StartSeed)

	for j := 0; j < pH-1; j++ {
		for i := 0; i < pW-1; i++ {
			v00 := parentOut[j*pW+i]
			v10 := parentOut[j*pW+i+1]
			v01 := parentOut[(j+1)*pW+i]
			v11 := parentOut[(j+1)*pW+i+1]

			if v00 == v01 && v00 == v10 && v00 == v11 {
				buf[(j*2)*newW+(i*2)] = v00
				buf[(j*2)*newW+(i*2+1)] = v00
				buf[(j*2+1)*newW+(i*2)] = v00
				buf[(j*2+1)*newW+(i*2+1)] = v00
				continue
			}

			chunkX, chunkZ := (i+pX)*2, (j+pZ)*2

			cs := ss
			cs += uint32(chunkX)
			cs *= cs*1284865837 + 4150755663
			cs += uint32(chunkZ)
			cs *= cs*1284865837 + 4150755663
			cs += uint32(chunkX)
			cs *= cs*1284865837 + 4150755663
			cs += uint32(chunkZ)

			buf[(j*2)*newW+(i*2)] = v00

			if (cs>>24)&1 != 0 {
				buf[(j*2+1)*newW+(i*2)] = v01
			} else {
				buf[(j*2+1)*newW+(i*2)] = v00
			}

			cs *= cs*1284865837 + 4150755663
			cs += st
			if (cs>>24)&1 != 0 {
				buf[(j*2)*newW+(i*2+1)] = v10
			} else {
				buf[(j*2)*newW+(i*2+1)] = v00
			}

			buf[(j*2+1)*newW+(i*2+1)] = select4(cs, st, v00, v01, v10, v11)
		}
	}

	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			out[j*w+i] = buf[(j+(z&1))*newW+(i+(x&1))]
		}
	}

	return 0
}

// MapLand 对应 mapLand 层
func MapLand(l *Layer, out []int, x, z, w, h int) int {
	pX, pZ := x-1, z-1
	pW, pH := w+2, h+2

	parentOut := make([]int, pW*pH)
	err := l.P.GetMap(l.P, parentOut, pX, pZ, pW, pH)
	if err != 0 {
		return err
	}

	st := l.StartSalt
	ss := l.StartSeed

	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			v00 := parentOut[j*pW+i]
			v10 := parentOut[j*pW+i+2]
			v01 := parentOut[(j+2)*pW+i]
			v11 := parentOut[(j+2)*pW+i+2]
			vCenter := parentOut[(j+1)*pW+i+1]

			v := vCenter
			if vCenter == int(Ocean) {
				if v00 != int(Ocean) || v10 != int(Ocean) || v01 != int(Ocean) || v11 != int(Ocean) {
					cs := getChunkSeed(ss, i+x, j+z)
					inc := 0
					v = 1
					if v00 != int(Ocean) {
						inc++
						v = v00
						cs = mcStepSeed(cs, st)
					}
					if v10 != int(Ocean) {
						inc++
						if inc == 1 || mcFirstIsZero(cs, 2) == 1 {
							v = v10
						}
						cs = mcStepSeed(cs, st)
					}
					if v01 != int(Ocean) {
						inc++
						switch inc {
						case 1:
							v = v01
						case 2:
							if mcFirstIsZero(cs, 2) == 1 {
								v = v01
							}
						default:
							if mcFirstIsZero(cs, 3) == 1 {
								v = v01
							}
						}
						cs = mcStepSeed(cs, st)
					}
					if v11 != int(Ocean) {
						inc++
						switch inc {
						case 1:
							v = v11
						case 2:
							if mcFirstIsZero(cs, 2) == 1 {
								v = v11
							}
						case 3:
							if mcFirstIsZero(cs, 3) == 1 {
								v = v11
							}
						default:
							if mcFirstIsZero(cs, 4) == 1 {
								v = v11
							}
						}
						cs = mcStepSeed(cs, st)
					}

					if v != int(Forest) {
						if mcFirstIsZero(cs, 3) == 0 {
							v = int(Ocean)
						}
					}
				}
			} else if vCenter == int(Forest) {
				// keep forest
			} else {
				if v00 == int(Ocean) || v10 == int(Ocean) || v01 == int(Ocean) || v11 == int(Ocean) {
					cs := getChunkSeed(ss, i+x, j+z)
					if mcFirstIsZero(cs, 5) == 1 {
						v = int(Ocean)
					}
				}
			}
			out[j*w+i] = v
		}
	}
	return 0
}

// MapIsland 对应 mapIsland 层
func MapIsland(l *Layer, out []int, x, z, w, h int) int {
	pX, pZ := x-1, z-1
	pW, pH := w+2, h+2

	parentOut := make([]int, pW*pH)
	err := l.P.GetMap(l.P, parentOut, pX, pZ, pW, pH)
	if err != 0 {
		return err
	}

	ss := l.StartSeed
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			v11 := parentOut[(j+1)*pW+i+1]
			out[j*w+i] = v11
			if v11 == int(Oceanic) {
				if parentOut[j*pW+i+1] != int(Oceanic) ||
					parentOut[(j+1)*pW+i+2] != int(Oceanic) ||
					parentOut[(j+1)*pW+i] != int(Oceanic) ||
					parentOut[(j+2)*pW+i+1] != int(Oceanic) {
					cs := getChunkSeed(ss, i+x, j+z)
					if mcFirstIsZero(cs, 2) == 1 {
						out[j*w+i] = 1
					}
				}
			}
		}
	}
	return 0
}

// MapSnow 对应 mapSnow 层
func MapSnow(l *Layer, out []int, x, z, w, h int) int {
	pX, pZ := x-1, z-1
	pW, pH := w+2, h+2

	parentOut := make([]int, pW*pH)
	err := l.P.GetMap(l.P, parentOut, pX, pZ, pW, pH)
	if err != 0 {
		return err
	}

	ss := l.StartSeed
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			v11 := parentOut[(j+1)*pW+i+1]
			if !Biome(v11).IsShallowOcean() {
				cs := getChunkSeed(ss, i+x, j+z)
				r := mcFirstInt(cs, 6)
				if r == 0 {
					v11 = int(Freezing)
				} else if r <= 1 {
					v11 = int(Cold)
				} else {
					v11 = int(Warm)
				}
			}
			out[j*w+i] = v11
		}
	}
	return 0
}

// MapCool 对应 mapCool 层
func MapCool(l *Layer, out []int, x, z, w, h int) int {
	pX, pZ := x-1, z-1
	pW, pH := w+2, h+2

	parentOut := make([]int, pW*pH)
	err := l.P.GetMap(l.P, parentOut, pX, pZ, pW, pH)
	if err != 0 {
		return err
	}

	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			v11 := parentOut[(j+1)*pW+i+1]
			if v11 == int(Warm) {
				v10 := parentOut[j*pW+i+1]
				v21 := parentOut[(j+1)*pW+i+2]
				v01 := parentOut[(j+1)*pW+i]
				v12 := parentOut[(j+2)*pW+i+1]

				if v10 == int(Cold) || v21 == int(Cold) || v01 == int(Cold) || v12 == int(Cold) ||
					v10 == int(Freezing) || v21 == int(Freezing) || v01 == int(Freezing) || v12 == int(Freezing) {
					v11 = int(Lush)
				}
			}
			out[j*w+i] = v11
		}
	}
	return 0
}

// MapHeat 对应 mapHeat 层
func MapHeat(l *Layer, out []int, x, z, w, h int) int {
	pX, pZ := x-1, z-1
	pW, pH := w+2, h+2

	parentOut := make([]int, pW*pH)
	err := l.P.GetMap(l.P, parentOut, pX, pZ, pW, pH)
	if err != 0 {
		return err
	}

	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			v11 := parentOut[(j+1)*pW+i+1]
			if v11 == int(Freezing) {
				v10 := parentOut[j*pW+i+1]
				v21 := parentOut[(j+1)*pW+i+2]
				v01 := parentOut[(j+1)*pW+i]
				v12 := parentOut[(j+2)*pW+i+1]

				if v10 == int(Warm) || v21 == int(Warm) || v01 == int(Warm) || v12 == int(Warm) ||
					v10 == int(Lush) || v21 == int(Lush) || v01 == int(Lush) || v12 == int(Lush) {
					v11 = int(Cold)
				}
			}
			out[j*w+i] = v11
		}
	}
	return 0
}

// MapSpecial 对应 mapSpecial 层
func MapSpecial(l *Layer, out []int, x, z, w, h int) int {
	err := l.P.GetMap(l.P, out, x, z, w, h)
	if err != 0 {
		return err
	}

	st := l.StartSalt
	ss := l.StartSeed
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			v := out[j*w+i]
			if v == int(Oceanic) {
				continue
			}
			cs := getChunkSeed(ss, i+x, j+z)
			if mcFirstIsZero(cs, 13) == 1 {
				cs = mcStepSeed(cs, st)
				v |= (int(1+mcFirstInt(cs, 15)) << 8) & 0xf00
				out[j*w+i] = v
			}
		}
	}
	return 0
}

// MapMushroom 对应 mapMushroom 层
func MapMushroom(l *Layer, out []int, x, z, w, h int) int {
	pX, pZ := x-1, z-1
	pW, pH := w+2, h+2

	parentOut := make([]int, pW*pH)
	err := l.P.GetMap(l.P, parentOut, pX, pZ, pW, pH)
	if err != 0 {
		return err
	}

	ss := l.StartSeed
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			v11 := parentOut[(j+1)*pW+i+1]
			if v11 == 0 &&
				parentOut[j*pW+i] == 0 && parentOut[j*pW+i+2] == 0 &&
				parentOut[(j+2)*pW+i] == 0 && parentOut[(j+2)*pW+i+2] == 0 {
				cs := getChunkSeed(ss, i+x, j+z)
				if mcFirstIsZero(cs, 100) == 1 {
					v11 = int(MushroomFields)
				}
			}
			out[j*w+i] = v11
		}
	}
	return 0
}

// MapDeepOcean 对应 mapDeepOcean 层
func MapDeepOcean(l *Layer, out []int, x, z, w, h int) int {
	pX, pZ := x-1, z-1
	pW, pH := w+2, h+2

	parentOut := make([]int, pW*pH)
	err := l.P.GetMap(l.P, parentOut, pX, pZ, pW, pH)
	if err != 0 {
		return err
	}

	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			v11 := parentOut[(j+1)*pW+i+1]
			if Biome(v11).IsShallowOcean() {
				oceans := 0
				if Biome(parentOut[j*pW+i+1]).IsShallowOcean() {
					oceans++
				}
				if Biome(parentOut[(j+1)*pW+i+2]).IsShallowOcean() {
					oceans++
				}
				if Biome(parentOut[(j+1)*pW+i]).IsShallowOcean() {
					oceans++
				}
				if Biome(parentOut[(j+2)*pW+i+1]).IsShallowOcean() {
					oceans++
				}

				if oceans >= 4 {
					switch Biome(v11) {
					case WarmOcean:
						v11 = int(DeepWarmOcean)
					case LukewarmOcean:
						v11 = int(DeepLukewarmOcean)
					case Ocean:
						v11 = int(DeepOcean)
					case ColdOcean:
						v11 = int(DeepColdOcean)
					case FrozenOcean:
						v11 = int(DeepFrozenOcean)
					default:
						v11 = int(DeepOcean)
					}
				}
			}
			out[j*w+i] = v11
		}
	}
	return 0
}

var warmBiomes = []Biome{Desert, Desert, Desert, Savanna, Savanna, Plains}
var lushBiomes = []Biome{Forest, DarkForest, Mountains, Plains, BirchForest, Swamp}
var coldBiomes = []Biome{Forest, Mountains, Taiga, Plains}
var snowBiomes = []Biome{SnowyTundra, SnowyTundra, SnowyTundra, SnowyTaiga}

// MapBiome 对应 mapBiome 层
func MapBiome(l *Layer, out []int, x, z, w, h int) int {
	err := l.P.GetMap(l.P, out, x, z, w, h)
	if err != 0 {
		return err
	}

	mc := l.MC
	ss := l.StartSeed
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			id := out[j*w+i]
			hasHighBit := id & 0xf00
			id &= ^0xf00

			var v Biome
			if mc <= MC_1_6 {
				if Biome(id) == Ocean || Biome(id) == MushroomFields {
					continue
				}
				// Legacy biome selection not fully implemented here yet
				// but we can add it if needed.
				v = Plains // fallback
			} else {
				if Biome(id).IsOceanic() || Biome(id) == MushroomFields {
					continue
				}
				cs := getChunkSeed(ss, i+x, j+z)
				switch Biome(id) {
				case Warm:
					if hasHighBit != 0 {
						if mcFirstIsZero(cs, 3) == 1 {
							v = BadlandsPlateau
						} else {
							v = WoodedBadlandsPlateau
						}
					} else {
						v = warmBiomes[mcFirstInt(cs, 6)]
					}
				case Lush:
					if hasHighBit != 0 {
						v = Jungle
					} else {
						v = lushBiomes[mcFirstInt(cs, 6)]
					}
				case Cold:
					if hasHighBit != 0 {
						v = GiantTreeTaiga
					} else {
						v = coldBiomes[mcFirstInt(cs, 4)]
					}
				case Freezing:
					v = snowBiomes[mcFirstInt(cs, 4)]
				default:
					v = MushroomFields
				}
			}
			out[j*w+i] = int(v)
		}
	}
	return 0
}

// MapNoise 对应 mapNoise 层
func MapNoise(l *Layer, out []int, x, z, w, h int) int {
	err := l.P.GetMap(l.P, out, x, z, w, h)
	if err != 0 {
		return err
	}

	ss := l.StartSeed
	mod := 299999
	if l.MC <= MC_1_6 {
		mod = 2
	}

	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			if out[j*w+i] > 0 {
				cs := getChunkSeed(ss, i+x, j+z)
				out[j*w+i] = mcFirstInt(cs, mod) + 2
			} else {
				out[j*w+i] = 0
			}
		}
	}
	return 0
}

// MapBamboo 对应 mapBamboo 层
func MapBamboo(l *Layer, out []int, x, z, w, h int) int {
	err := l.P.GetMap(l.P, out, x, z, w, h)
	if err != 0 {
		return err
	}

	ss := l.StartSeed
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			if Biome(out[j*w+i]) != Jungle {
				continue
			}
			cs := getChunkSeed(ss, i+x, j+z)
			if mcFirstIsZero(cs, 10) == 1 {
				out[j*w+i] = int(BambooJungle)
			}
		}
	}
	return 0
}

func replaceEdge(out []int, idx, mc int, v10, v21, v01, v12 int, id, baseID, edgeID Biome) bool {
	if Biome(id) != baseID {
		return false
	}
	if AreSimilar(mc, Biome(v10), baseID) && AreSimilar(mc, Biome(v21), baseID) &&
		AreSimilar(mc, Biome(v01), baseID) && AreSimilar(mc, Biome(v12), baseID) {
		out[idx] = int(id)
	} else {
		out[idx] = int(edgeID)
	}
	return true
}

// MapBiomeEdge 对应 mapBiomeEdge 层
func MapBiomeEdge(l *Layer, out []int, x, z, w, h int) int {
	pX, pZ := x-1, z-1
	pW, pH := w+2, h+2

	parentOut := make([]int, pW*pH)
	err := l.P.GetMap(l.P, parentOut, pX, pZ, pW, pH)
	if err != 0 {
		return err
	}

	mc := l.MC
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			v11 := parentOut[(j+1)*pW+i+1]
			v10 := parentOut[j*pW+i+1]
			v21 := parentOut[(j+1)*pW+i+2]
			v01 := parentOut[(j+1)*pW+i]
			v12 := parentOut[(j+2)*pW+i+1]

			if replaceEdge(out, j*w+i, mc, v10, v21, v01, v12, Biome(v11), WoodedBadlandsPlateau, Badlands) ||
				replaceEdge(out, j*w+i, mc, v10, v21, v01, v12, Biome(v11), BadlandsPlateau, Badlands) ||
				replaceEdge(out, j*w+i, mc, v10, v21, v01, v12, Biome(v11), GiantTreeTaiga, Taiga) {
				continue
			}

			if Biome(v11) == Desert {
				if Biome(v10) == SnowyTundra || Biome(v21) == SnowyTundra || Biome(v01) == SnowyTundra || Biome(v12) == SnowyTundra {
					out[j*w+i] = int(WoodedMountains)
				} else {
					out[j*w+i] = v11
				}
			} else if Biome(v11) == Swamp {
				if Biome(v10) == Desert || Biome(v21) == Desert || Biome(v01) == Desert || Biome(v12) == Desert ||
					Biome(v10) == SnowyTaiga || Biome(v21) == SnowyTaiga || Biome(v01) == SnowyTaiga || Biome(v12) == SnowyTaiga ||
					Biome(v10) == SnowyTundra || Biome(v21) == SnowyTundra || Biome(v01) == SnowyTundra || Biome(v12) == SnowyTundra {
					out[j*w+i] = int(Plains)
				} else if Biome(v10) == Jungle || Biome(v21) == Jungle || Biome(v01) == Jungle || Biome(v12) == Jungle ||
					Biome(v10) == BambooJungle || Biome(v21) == BambooJungle || Biome(v01) == BambooJungle || Biome(v12) == BambooJungle {
					out[j*w+i] = int(JungleEdge)
				} else {
					out[j*w+i] = v11
				}
			} else {
				out[j*w+i] = v11
			}
		}
	}
	return 0
}

// MapHills 对应 mapHills 层
func MapHills(l *Layer, out []int, x, z, w, h int) int {
	pX, pZ := x-1, z-1
	pW, pH := w+2, h+2

	parentOut := make([]int, pW*pH)
	err := l.P.GetMap(l.P, parentOut, pX, pZ, pW, pH)
	if err != 0 {
		return err
	}

	riverOut := make([]int, pW*pH)
	err = l.P2.GetMap(l.P2, riverOut, pX, pZ, pW, pH)
	if err != 0 {
		return err
	}

	mc := l.MC
	st := l.StartSalt
	ss := l.StartSeed

	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			a11 := parentOut[(j+1)*pW+i+1]
			b11 := riverOut[(j+1)*pW+i+1]
			bn := -1
			if mc >= MC_1_7 {
				bn = (b11 - 2) % 29
			}

			if bn == 1 && b11 >= 2 && !Biome(a11).IsShallowOcean() {
				m := GetMutated(mc, Biome(a11))
				if m != None {
					out[j*w+i] = int(m)
				} else {
					out[j*w+i] = a11
				}
			} else {
				cs := getChunkSeed(ss, i+x, j+z)
				if bn == 0 || mcFirstIsZero(cs, 3) == 1 {
					hillID := Biome(a11)
					switch Biome(a11) {
					case Desert:
						hillID = DesertHills
					case Forest:
						hillID = WoodedHills
					case BirchForest:
						hillID = BirchForestHills
					case DarkForest:
						hillID = Plains
					case Taiga:
						hillID = TaigaHills
					case GiantTreeTaiga:
						hillID = GiantTreeTaigaHills
					case SnowyTaiga:
						hillID = SnowyTaigaHills
					case Plains:
						if mc <= MC_1_6 {
							hillID = Forest
						} else {
							cs = mcStepSeed(cs, st)
							if mcFirstIsZero(cs, 3) == 1 {
								hillID = WoodedHills
							} else {
								hillID = Forest
							}
						}
					case SnowyTundra:
						hillID = SnowyMountains
					case Jungle:
						hillID = JungleHills
					case BambooJungle:
						hillID = BambooJungleHills
					case Ocean:
						if mc >= MC_1_7 {
							hillID = DeepOcean
						}
					case Mountains:
						if mc >= MC_1_7 {
							hillID = WoodedMountains
						}
					case Savanna:
						hillID = SavannaPlateau
					default:
						if AreSimilar(mc, Biome(a11), WoodedBadlandsPlateau) {
							hillID = Badlands
						} else if Biome(a11).IsDeepOcean() {
							cs = mcStepSeed(cs, st)
							if mcFirstIsZero(cs, 3) == 1 {
								cs = mcStepSeed(cs, st)
								if mcFirstIsZero(cs, 2) == 1 {
									hillID = Plains
								} else {
									hillID = Forest
								}
							}
						}
					}

					if bn == 0 && hillID != Biome(a11) {
						m := GetMutated(mc, hillID)
						if m != None {
							hillID = m
						} else {
							hillID = Biome(a11)
						}
					}

					if hillID != Biome(a11) {
						a10 := parentOut[j*pW+i+1]
						a21 := parentOut[(j+1)*pW+i+2]
						a01 := parentOut[(j+1)*pW+i]
						a12 := parentOut[(j+2)*pW+i+1]
						equals := 0
						if AreSimilar(mc, Biome(a10), Biome(a11)) {
							equals++
						}
						if AreSimilar(mc, Biome(a21), Biome(a11)) {
							equals++
						}
						if AreSimilar(mc, Biome(a01), Biome(a11)) {
							equals++
						}
						if AreSimilar(mc, Biome(a12), Biome(a11)) {
							equals++
						}

						limit := 3
						if mc <= MC_1_6 {
							limit = 4
						}
						if equals >= limit {
							out[j*w+i] = int(hillID)
						} else {
							out[j*w+i] = a11
						}
					} else {
						out[j*w+i] = a11
					}
				} else {
					out[j*w+i] = a11
				}
			}
		}
	}
	return 0
}

func reduceID(id int) int {
	if id >= 2 {
		return 2 + (id & 1)
	}
	return id
}

// MapRiver 对应 mapRiver 层
func MapRiver(l *Layer, out []int, x, z, w, h int) int {
	pX, pZ := x-1, z-1
	pW, pH := w+2, h+2

	parentOut := make([]int, pW*pH)
	err := l.P.GetMap(l.P, parentOut, pX, pZ, pW, pH)
	if err != 0 {
		return err
	}

	mc := l.MC
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			v01 := parentOut[(j+1)*pW+i]
			v11 := parentOut[(j+1)*pW+i+1]
			v21 := parentOut[(j+1)*pW+i+2]
			v10 := parentOut[j*pW+i+1]
			v12 := parentOut[(j+2)*pW+i+1]

			if mc >= MC_1_7 {
				v01 = reduceID(v01)
				v11 = reduceID(v11)
				v21 = reduceID(v21)
				v10 = reduceID(v10)
				v12 = reduceID(v12)
			} else if v11 == 0 {
				out[j*w+i] = int(River)
				continue
			}

			if v11 == v01 && v11 == v10 && v11 == v12 && v11 == v21 {
				out[j*w+i] = -1
			} else {
				out[j*w+i] = int(River)
			}
		}
	}
	return 0
}

// MapSmooth 对应 mapSmooth 层
func MapSmooth(l *Layer, out []int, x, z, w, h int) int {
	pX, pZ := x-1, z-1
	pW, pH := w+2, h+2

	parentOut := make([]int, pW*pH)
	err := l.P.GetMap(l.P, parentOut, pX, pZ, pW, pH)
	if err != 0 {
		return err
	}

	ss := l.StartSeed
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			v11 := parentOut[(j+1)*pW+i+1]
			v01 := parentOut[(j+1)*pW+i]
			v10 := parentOut[j*pW+i+1]
			v21 := parentOut[(j+1)*pW+i+2]
			v12 := parentOut[(j+2)*pW+i+1]

			if v11 != v01 || v11 != v10 {
				if v01 == v21 && v10 == v12 {
					cs := getChunkSeed(ss, i+x, j+z)
					if cs&(1<<24) != 0 {
						v11 = v10
					} else {
						v11 = v01
					}
				} else {
					if v01 == v21 {
						v11 = v01
					}
					if v10 == v12 {
						v11 = v10
					}
				}
			}
			out[j*w+i] = v11
		}
	}
	return 0
}

// MapSunflower 对应 mapSunflower 层
func MapSunflower(l *Layer, out []int, x, z, w, h int) int {
	err := l.P.GetMap(l.P, out, x, z, w, h)
	if err != 0 {
		return err
	}

	ss := l.StartSeed
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			if Biome(out[j*w+i]) == Plains {
				cs := getChunkSeed(ss, i+x, j+z)
				if mcFirstIsZero(cs, 57) == 1 {
					out[j*w+i] = int(SunflowerPlains)
				}
			}
		}
	}
	return 0
}

func isAll4JFTO(mc int, a, b, c, d Biome) bool {
	return (GetCategory(mc, a) == Jungle || a == Forest || a == Taiga || a.IsOceanic()) &&
		(GetCategory(mc, b) == Jungle || b == Forest || b == Taiga || b.IsOceanic()) &&
		(GetCategory(mc, c) == Jungle || c == Forest || c == Taiga || c.IsOceanic()) &&
		(GetCategory(mc, d) == Jungle || d == Forest || d == Taiga || d.IsOceanic())
}

// MapShore 对应 mapShore 层
func MapShore(l *Layer, out []int, x, z, w, h int) int {
	pX, pZ := x-1, z-1
	pW, pH := w+2, h+2

	parentOut := make([]int, pW*pH)
	err := l.P.GetMap(l.P, parentOut, pX, pZ, pW, pH)
	if err != 0 {
		return err
	}

	mc := l.MC
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			v11 := parentOut[(j+1)*pW+i+1]
			v10 := parentOut[j*pW+i+1]
			v21 := parentOut[(j+1)*pW+i+2]
			v01 := parentOut[(j+1)*pW+i]
			v12 := parentOut[(j+2)*pW+i+1]

			if Biome(v11) == MushroomFields {
				if Biome(v10) == Ocean || Biome(v21) == Ocean || Biome(v01) == Ocean || Biome(v12) == Ocean {
					out[j*w+i] = int(MushroomFieldShore)
				} else {
					out[j*w+i] = v11
				}
				continue
			}

			if mc <= MC_1_0 {
				out[j*w+i] = v11
				continue
			}

			if mc <= MC_1_6 {
				if Biome(v11) == Mountains {
					if Biome(v10) != Mountains || Biome(v21) != Mountains || Biome(v01) != Mountains || Biome(v12) != Mountains {
						v11 = int(MountainEdge)
					}
				} else if Biome(v11) != Ocean && Biome(v11) != River && Biome(v11) != Swamp {
					if Biome(v10) == Ocean || Biome(v21) == Ocean || Biome(v01) == Ocean || Biome(v12) == Ocean {
						v11 = int(Beach)
					}
				}
				out[j*w+i] = v11
			} else if GetCategory(mc, Biome(v11)) == Jungle {
				if isAll4JFTO(mc, Biome(v10), Biome(v21), Biome(v01), Biome(v12)) {
					if Biome(v10).IsOceanic() || Biome(v21).IsOceanic() || Biome(v01).IsOceanic() || Biome(v12).IsOceanic() {
						out[j*w+i] = int(Beach)
					} else {
						out[j*w+i] = v11
					}
				} else {
					out[j*w+i] = int(JungleEdge)
				}
			} else if Biome(v11) == Mountains || Biome(v11) == WoodedMountains {
				if Biome(v10).IsOceanic() || Biome(v21).IsOceanic() || Biome(v01).IsOceanic() || Biome(v12).IsOceanic() {
					out[j*w+i] = int(StoneShore)
				} else {
					out[j*w+i] = v11
				}
			} else if Biome(v11).IsSnowy() {
				if Biome(v10).IsOceanic() || Biome(v21).IsOceanic() || Biome(v01).IsOceanic() || Biome(v12).IsOceanic() {
					out[j*w+i] = int(SnowyBeach)
				} else {
					out[j*w+i] = v11
				}
			} else if Biome(v11) == Badlands || Biome(v11) == WoodedBadlandsPlateau {
				if !Biome(v10).IsOceanic() && !Biome(v21).IsOceanic() && !Biome(v01).IsOceanic() && !Biome(v12).IsOceanic() {
					if Biome(v10).IsMesa() && Biome(v21).IsMesa() && Biome(v01).IsMesa() && Biome(v12).IsMesa() {
						out[j*w+i] = v11
					} else {
						out[j*w+i] = int(Desert)
					}
				} else {
					out[j*w+i] = v11
				}
			} else {
				if Biome(v11) != Ocean && Biome(v11) != DeepOcean && Biome(v11) != River && Biome(v11) != Swamp {
					if Biome(v10).IsOceanic() || Biome(v21).IsOceanic() || Biome(v01).IsOceanic() || Biome(v12).IsOceanic() {
						out[j*w+i] = int(Beach)
					} else {
						out[j*w+i] = v11
					}
				} else {
					out[j*w+i] = v11
				}
			}
		}
	}
	return 0
}

// MapSwampRiver 对应 mapSwampRiver 层
func MapSwampRiver(l *Layer, out []int, x, z, w, h int) int {
	err := l.P.GetMap(l.P, out, x, z, w, h)
	if err != 0 {
		return err
	}

	ss := l.StartSeed
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			v := Biome(out[j*w+i])
			if v != Swamp && v != Jungle && v != JungleHills {
				continue
			}
			cs := getChunkSeed(ss, i+x, j+z)
			mod := 8
			if v == Swamp {
				mod = 6
			}
			if mcFirstIsZero(cs, mod) == 1 {
				out[j*w+i] = int(River)
			}
		}
	}
	return 0
}

// MapRiverMix 对应 mapRiverMix 层
func MapRiverMix(l *Layer, out []int, x, z, w, h int) int {
	err := l.P.GetMap(l.P, out, x, z, w, h)
	if err != 0 {
		return err
	}

	riverOut := make([]int, w*h)
	err = l.P2.GetMap(l.P2, riverOut, x, z, w, h)
	if err != 0 {
		return err
	}

	mc := l.MC
	for i := 0; i < w*h; i++ {
		v := Biome(out[i])
		if Biome(riverOut[i]) == River && v != Ocean && (mc <= MC_1_6 || !v.IsOceanic()) {
			if v == SnowyTundra {
				out[i] = int(FrozenRiver)
			} else if v == MushroomFields || v == MushroomFieldShore {
				out[i] = int(MushroomFieldShore)
			} else {
				out[i] = int(River)
			}
		}
	}
	return 0
}

// MapOceanTemp 对应 mapOceanTemp 层
func MapOceanTemp(l *Layer, out []int, x, z, w, h int) int {
	rnd := l.Noise
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			tmp := rnd.Sample(float64(i+x)/8.0, float64(j+z)/8.0, 0, 0, 0)
			if tmp > 0.4 {
				out[j*w+i] = int(WarmOcean)
			} else if tmp > 0.2 {
				out[j*w+i] = int(LukewarmOcean)
			} else if tmp < -0.4 {
				out[j*w+i] = int(FrozenOcean)
			} else if tmp < -0.2 {
				out[j*w+i] = int(ColdOcean)
			} else {
				out[j*w+i] = int(Ocean)
			}
		}
	}
	return 0
}

// MapOceanMix 对应 mapOceanMix 层
func MapOceanMix(l *Layer, out []int, x, z, w, h int) int {
	err := l.P2.GetMap(l.P2, out, x, z, w, h)
	if err != 0 {
		return err
	}

	lx0, lx1 := 0, w
	lz0, lz1 := 0, h

	for j := 0; j < h; j++ {
		jcentre := (j-8 > 0 && j+9 < h)
		for i := 0; i < w; i++ {
			if jcentre && i-8 > 0 && i+9 < w {
				continue
			}
			oceanID := Biome(out[j*w+i])
			if oceanID == WarmOcean || oceanID == FrozenOcean {
				if i-8 < lx0 {
					lx0 = i - 8
				}
				if i+9 > lx1 {
					lx1 = i + 9
				}
				if j-8 < lz0 {
					lz0 = j - 8
				}
				if j+9 > lz1 {
					lz1 = j + 9
				}
			}
		}
	}

	lw := lx1 - lx0
	lh := lz1 - lz0
	land := make([]int, lw*lh)
	err = l.P.GetMap(l.P, land, x+lx0, z+lz0, lw, lh)
	if err != 0 {
		return err
	}

	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			landID := Biome(land[(i-lx0)+(j-lz0)*lw])
			oceanID := Biome(out[j*w+i])
			replaceID := None

			if !landID.IsOceanic() {
				out[j*w+i] = int(landID)
				continue
			}

			if oceanID == WarmOcean {
				replaceID = LukewarmOcean
			}
			if oceanID == FrozenOcean {
				replaceID = ColdOcean
			}

			if replaceID != None {
				foundLand := false
				for ii := -8; ii <= 8; ii += 4 {
					for jj := -8; jj <= 8; jj += 4 {
						id := Biome(land[(i+ii-lx0)+(j+jj-lz0)*lw])
						if !id.IsOceanic() {
							out[j*w+i] = int(replaceID)
							foundLand = true
							break
						}
					}
					if foundLand {
						break
					}
				}
				if foundLand {
					continue
				}
			}

			if landID == DeepOcean {
				switch oceanID {
				case LukewarmOcean:
					oceanID = DeepLukewarmOcean
				case Ocean:
					oceanID = DeepOcean
				case ColdOcean:
					oceanID = DeepColdOcean
				case FrozenOcean:
					oceanID = DeepFrozenOcean
				}
			}
			out[j*w+i] = int(oceanID)
		}
	}
	return 0
}

// MapVoronoi114 对应 mapVoronoi114 层
func MapVoronoi114(l *Layer, out []int, x, z, w, h int) int {
	x -= 2
	z -= 2
	pX, pZ := x>>2, z>>2
	pW := ((x + w) >> 2) - pX + 2
	pH := ((z + h) >> 2) - pZ + 2

	parentOut := make([]int, pW*pH)
	if l.P != nil {
		err := l.P.GetMap(l.P, parentOut, pX, pZ, pW, pH)
		if err != 0 {
			return err
		}
	}

	st := l.StartSalt
	ss := l.StartSeed

	for pj := 0; pj < pH-1; pj++ {
		v00 := parentOut[pj*pW]
		v01 := parentOut[(pj+1)*pW]
		j4 := (pZ+pj)*4 - z

		for pi := 0; pi < pW-1; pi++ {
			v10 := parentOut[pj*pW+pi+1]
			v11 := parentOut[(pj+1)*pW+pi+1]
			i4 := (pX+pi)*4 - x

			if v00 == v01 && v00 == v10 && v00 == v11 {
				for jj := 0; jj < 4; jj++ {
					j := j4 + jj
					if j < 0 || j >= h {
						continue
					}
					for ii := 0; ii < 4; ii++ {
						i := i4 + ii
						if i < 0 || i >= w {
							continue
						}
						out[j*w+i] = v00
					}
				}
			} else {
				cs00 := getChunkSeed(ss, (pi+pX)*4, (pj+pZ)*4)
				da1 := (int64(mcFirstInt(cs00, 1024)) - 512) * 36
				cs00 = mcStepSeed(cs00, st)
				da2 := (int64(mcFirstInt(cs00, 1024)) - 512) * 36

				cs10 := getChunkSeed(ss, (pi+pX+1)*4, (pj+pZ)*4)
				db1 := (int64(mcFirstInt(cs10, 1024))-512)*36 + 40*1024
				cs10 = mcStepSeed(cs10, st)
				db2 := (int64(mcFirstInt(cs10, 1024)) - 512) * 36

				cs01 := getChunkSeed(ss, (pi+pX)*4, (pj+pZ+1)*4)
				dc1 := (int64(mcFirstInt(cs01, 1024)) - 512) * 36
				cs01 = mcStepSeed(cs01, st)
				dc2 := (int64(mcFirstInt(cs01, 1024))-512)*36 + 40*1024

				cs11 := getChunkSeed(ss, (pi+pX+1)*4, (pj+pZ+1)*4)
				dd1 := (int64(mcFirstInt(cs11, 1024))-512)*36 + 40*1024
				cs11 = mcStepSeed(cs11, st)
				dd2 := (int64(mcFirstInt(cs11, 1024))-512)*36 + 40*1024

				for jj := 0; jj < 4; jj++ {
					j := j4 + jj
					if j < 0 || j >= h {
						continue
					}
					mj := int64(10240 * jj)
					sja := (mj - da2) * (mj - da2)
					sjb := (mj - db2) * (mj - db2)
					sjc := (mj - dc2) * (mj - dc2)
					sjd := (mj - dd2) * (mj - dd2)

					for ii := 0; ii < 4; ii++ {
						i := i4 + ii
						if i < 0 || i >= w {
							continue
						}
						mi := int64(10240 * ii)
						da := (mi-da1)*(mi-da1) + sja
						db := (mi-db1)*(mi-db1) + sjb
						dc := (mi-dc1)*(mi-dc1) + sjc
						dd := (mi-dd1)*(mi-dd1) + sjd

						var v int
						if da < db && da < dc && da < dd {
							v = v00
						} else if db < da && db < dc && db < dd {
							v = v10
						} else if dc < da && dc < db && dc < dd {
							v = v01
						} else {
							v = v11
						}
						out[j*w+i] = v
					}
				}
			}
			v00 = v10
			v01 = v11
		}
	}
	return 0
}

func bswap32(x uint32) uint32 {
	return ((x & 0x000000ff) << 24) | ((x & 0x0000ff00) << 8) |
		((x & 0x00ff0000) >> 8) | ((x & 0xff000000) >> 24)
}

func rotr32(a uint32, b uint8) uint32 {
	return (a >> b) | (a << (32 - b))
}

func GetVoronoiSHA(seed uint64) uint64 {
	var K = [64]uint32{
		0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5,
		0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
		0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3,
		0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
		0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc,
		0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
		0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7,
		0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
		0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13,
		0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
		0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3,
		0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
		0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5,
		0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
		0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208,
		0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
	}
	var B = [8]uint32{
		0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a,
		0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19,
	}

	var m [64]uint32
	m[0] = bswap32(uint32(seed))
	m[1] = bswap32(uint32(seed >> 32))
	m[2] = 0x80000000
	m[15] = 0x00000040

	for i := 16; i < 64; i++ {
		m[i] = m[i-7] + m[i-16]
		x := m[i-15]
		m[i] += rotr32(x, 7) ^ rotr32(x, 18) ^ (x >> 3)
		x = m[i-2]
		m[i] += rotr32(x, 17) ^ rotr32(x, 19) ^ (x >> 10)
	}

	a0, a1, a2, a3, a4, a5, a6, a7 := B[0], B[1], B[2], B[3], B[4], B[5], B[6], B[7]

	for i := 0; i < 64; i++ {
		x := a7 + K[i] + m[i]
		x += rotr32(a4, 6) ^ rotr32(a4, 11) ^ rotr32(a4, 25)
		x += (a4 & a5) ^ (^a4 & a6)

		y := rotr32(a0, 2) ^ rotr32(a0, 13) ^ rotr32(a0, 22)
		y += (a0 & a1) ^ (a0 & a2) ^ (a1 & a2)

		a7 = a6
		a6 = a5
		a5 = a4
		a4 = a3 + x
		a3 = a2
		a2 = a1
		a1 = a0
		a0 = x + y
	}

	a0 += B[0]
	a1 += B[1]

	return uint64(bswap32(a0)) | (uint64(bswap32(a1)) << 32)
}

func getVoronoiCell(sha uint64, a, b, c int, x, y, z *int) {
	s := sha
	s = mcStepSeed(s, uint64(a))
	s = mcStepSeed(s, uint64(b))
	s = mcStepSeed(s, uint64(c))
	s = mcStepSeed(s, uint64(a))
	s = mcStepSeed(s, uint64(b))
	s = mcStepSeed(s, uint64(c))

	*x = (int((s>>24)&1023) - 512) * 36
	s = mcStepSeed(s, sha)
	*y = (int((s>>24)&1023) - 512) * 36
	s = mcStepSeed(s, sha)
	*z = (int((s>>24)&1023) - 512) * 36
}

func MapVoronoiPlane(sha uint64, out, src []int, x, z, w, h, y, px, pz, pw, ph int) {
	x -= 2
	y -= 2
	z -= 2

	for pj := 0; pj < ph-1; pj++ {
		v00 := src[pj*pw]
		v10 := src[(pj+1)*pw]
		pjz := pz + pj
		j4 := pjz*4 - z
		prev_skip := true

		var x000, x001, x100, x101 int
		var y000, y001, y100, y101 int
		var z000, z001, z100, z101 int

		for pi := 0; pi < pw-1; pi++ {
			v01 := src[pj*pw+pi+1]
			v11 := src[(pj+1)*pw+pi+1]
			pix := px + pi
			i4 := pix*4 - x

			if v00 == v01 && v00 == v10 && v00 == v11 {
				for jj := 0; jj < 4; jj++ {
					j := j4 + jj
					if j < 0 || j >= h {
						continue
					}
					for ii := 0; ii < 4; ii++ {
						i := i4 + ii
						if i < 0 || i >= w {
							continue
						}
						out[j*w+i] = v00
					}
				}
				prev_skip = true
				v00 = v01
				v10 = v11
				continue
			}

			if prev_skip {
				getVoronoiCell(sha, pix, y-1, pjz, &x000, &y000, &z000)
				getVoronoiCell(sha, pix, y, pjz, &x001, &y001, &z001)
				getVoronoiCell(sha, pix, y-1, pjz+1, &x100, &y100, &z100)
				getVoronoiCell(sha, pix, y, pjz+1, &x101, &y101, &z101)
				prev_skip = false
			}

			var x010, x011, x110, x111 int
			var y010, y011, y110, y111 int
			var z010, z011, z110, z111 int
			getVoronoiCell(sha, pix+1, y-1, pjz, &x010, &y010, &z010)
			getVoronoiCell(sha, pix+1, y, pjz, &x011, &y011, &z011)
			getVoronoiCell(sha, pix+1, y-1, pjz+1, &x110, &y110, &z110)
			getVoronoiCell(sha, pix+1, y, pjz+1, &x111, &y111, &z111)

			for jj := 0; jj < 4; jj++ {
				j := j4 + jj
				if j < 0 || j >= h {
					continue
				}
				for ii := 0; ii < 4; ii++ {
					i := i4 + ii
					if i < 0 || i >= w {
						continue
					}

					const A = 40 * 1024
					const B = 20 * 1024
					dx := int64(ii * 10 * 1024)
					dz := int64(jj * 10 * 1024)
					dmin := uint64(math.MaxUint64)

					v := v00
					r := int64(x000) + dx
					d := uint64(r * r)
					r = int64(y000) + B
					d += uint64(r * r)
					r = int64(z000) + dz
					d += uint64(r * r)
					if d < dmin {
						dmin = d
					}

					r = int64(x001) + dx
					d = uint64(r * r)
					r = int64(y001) - B
					d += uint64(r * r)
					r = int64(z001) + dz
					d += uint64(r * r)
					if d < dmin {
						dmin = d
					}

					r = int64(x010) - A + dx
					d = uint64(r * r)
					r = int64(y010) + B
					d += uint64(r * r)
					r = int64(z010) + dz
					d += uint64(r * r)
					if d < dmin {
						dmin = d
						v = v01
					}

					r = int64(x011) - A + dx
					d = uint64(r * r)
					r = int64(y011) - B
					d += uint64(r * r)
					r = int64(z011) + dz
					d += uint64(r * r)
					if d < dmin {
						dmin = d
						v = v01
					}

					r = int64(x100) + dx
					d = uint64(r * r)
					r = int64(y100) + B
					d += uint64(r * r)
					r = int64(z100) - A + dz
					d += uint64(r * r)
					if d < dmin {
						dmin = d
						v = v10
					}

					r = int64(x101) + dx
					d = uint64(r * r)
					r = int64(y101) - B
					d += uint64(r * r)
					r = int64(z101) - A + dz
					d += uint64(r * r)
					if d < dmin {
						dmin = d
						v = v10
					}

					r = int64(x110) - A + dx
					d = uint64(r * r)
					r = int64(y110) + B
					d += uint64(r * r)
					r = int64(z110) - A + dz
					d += uint64(r * r)
					if d < dmin {
						dmin = d
						v = v11
					}

					r = int64(x111) - A + dx
					d = uint64(r * r)
					r = int64(y111) - B
					d += uint64(r * r)
					r = int64(z111) - A + dz
					d += uint64(r * r)
					if d < dmin {
						dmin = d
						v = v11
					}

					out[j*w+i] = v
				}
			}
			x000, y000, z000 = x010, y010, z010
			x100, y100, z100 = x110, y110, z110
			x001, y001, z001 = x011, y011, z011
			x101, y101, z101 = x111, y111, z111
			v00 = v01
			v10 = v11
		}
	}
}

// MapVoronoi 对应 mapVoronoi 层
func MapVoronoi(l *Layer, out []int, x, z, w, h int) int {
	x -= 2
	z -= 2
	px := x >> 2
	pz := z >> 2
	pw := ((x + w) >> 2) - px + 2
	ph := ((z + h) >> 2) - pz + 2

	src := make([]int, pw*ph)
	if l.P != nil {
		err := l.P.GetMap(l.P, src, px, pz, pw, ph)
		if err != 0 {
			return err
		}
	}

	MapVoronoiPlane(l.StartSalt, out, src, x, z, w, h, 0, px, pz, pw, ph)
	return 0
}

func SetupLayer(l *Layer, mapFunc func(*Layer, []int, int, int, int, int) int, mc int, zoom, edge int, saltbase uint64, p, p2 *Layer) *Layer {
	l.GetMap = mapFunc
	l.MC = mc
	l.Zoom = zoom
	l.Edge = edge
	l.Scale = 0
	if saltbase == 0 || saltbase == math.MaxUint64 {
		l.LayerSalt = saltbase
	} else {
		l.LayerSalt = getLayerSalt(saltbase)
	}
	l.StartSalt = 0
	l.StartSeed = 0
	l.Noise = nil
	l.P = p
	l.P2 = p2
	return l
}

func setupScale(l *Layer, scale int) {
	l.Scale = scale
	if l.P != nil {
		setupScale(l.P, scale*l.Zoom)
	}
	if l.P2 != nil {
		setupScale(l.P2, scale*l.Zoom)
	}
}

func SetupLayerStack(g *LayerStack, mc int, largeBiomes bool) {
	l := g.Layers[:]
	var map_land = MapLand

	if mc == MC_B1_8 {
		// map_land = MapLandB18 // Not implemented yet
		p := SetupLayer(&l[L_CONTINENT_4096], MapContinent, mc, 1, 0, 1, nil, nil)
		p = SetupLayer(&l[L_ZOOM_4096], MapZoomFuzzy, mc, 2, 3, 2000, p, nil)
		p = SetupLayer(&l[L_LAND_4096], map_land, mc, 1, 2, 1, p, nil)
		p = SetupLayer(&l[L_ZOOM_2048], MapZoom, mc, 2, 3, 2001, p, nil)
		p = SetupLayer(&l[L_LAND_2048], map_land, mc, 1, 2, 2, p, nil)
		p = SetupLayer(&l[L_ZOOM_1024], MapZoom, mc, 2, 3, 2002, p, nil)
		p = SetupLayer(&l[L_LAND_1024_A], map_land, mc, 1, 2, 3, p, nil)
		p = SetupLayer(&l[L_ZOOM_512], MapZoom, mc, 2, 3, 2003, p, nil)
		p = SetupLayer(&l[L_LAND_512], map_land, mc, 1, 2, 3, p, nil)
		p = SetupLayer(&l[L_ZOOM_256], MapZoom, mc, 2, 3, 2004, p, nil)
		p = SetupLayer(&l[L_LAND_256], map_land, mc, 1, 2, 3, p, nil)
		p = SetupLayer(&l[L_BIOME_256], MapBiome, mc, 1, 0, 200, p, nil)
		p = SetupLayer(&l[L_ZOOM_128], MapZoom, mc, 2, 3, 1000, p, nil)
		p = SetupLayer(&l[L_ZOOM_64], MapZoom, mc, 2, 3, 1001, p, nil)
		SetupLayer(&l[L_RIVER_INIT_256], MapNoise, mc, 1, 0, 100, &l[L_LAND_256], nil)
	} else if mc <= MC_1_6 {
		// map_land = MapLand16 // Not implemented yet
		p := SetupLayer(&l[L_CONTINENT_4096], MapContinent, mc, 1, 0, 1, nil, nil)
		p = SetupLayer(&l[L_ZOOM_2048], MapZoomFuzzy, mc, 2, 3, 2000, p, nil)
		p = SetupLayer(&l[L_LAND_2048], map_land, mc, 1, 2, 1, p, nil)
		p = SetupLayer(&l[L_ZOOM_1024], MapZoom, mc, 2, 3, 2001, p, nil)
		p = SetupLayer(&l[L_LAND_1024_A], map_land, mc, 1, 2, 2, p, nil)
		// p = SetupLayer(&l[L_SNOW_1024], MapSnow16, mc, 1, 2, 2, p, nil) // Not implemented yet
		p = SetupLayer(&l[L_ZOOM_512], MapZoom, mc, 2, 3, 2002, p, nil)
		p = SetupLayer(&l[L_LAND_512], map_land, mc, 1, 2, 3, p, nil)
		p = SetupLayer(&l[L_ZOOM_256], MapZoom, mc, 2, 3, 2003, p, nil)
		p = SetupLayer(&l[L_LAND_256], map_land, mc, 1, 2, 4, p, nil)
		p = SetupLayer(&l[L_MUSHROOM_256], MapMushroom, mc, 1, 2, 5, p, nil)
		p = SetupLayer(&l[L_BIOME_256], MapBiome, mc, 1, 0, 200, p, nil)
		p = SetupLayer(&l[L_ZOOM_128], MapZoom, mc, 2, 3, 1000, p, nil)
		p = SetupLayer(&l[L_ZOOM_64], MapZoom, mc, 2, 3, 1001, p, nil)
		SetupLayer(&l[L_RIVER_INIT_256], MapNoise, mc, 1, 0, 100, &l[L_MUSHROOM_256], nil)
	} else {
		p := SetupLayer(&l[L_CONTINENT_4096], MapContinent, mc, 1, 0, 1, nil, nil)
		p = SetupLayer(&l[L_ZOOM_2048], MapZoomFuzzy, mc, 2, 3, 2000, p, nil)
		p = SetupLayer(&l[L_LAND_2048], map_land, mc, 1, 2, 1, p, nil)
		p = SetupLayer(&l[L_ZOOM_1024], MapZoom, mc, 2, 3, 2001, p, nil)
		p = SetupLayer(&l[L_LAND_1024_A], map_land, mc, 1, 2, 2, p, nil)
		p = SetupLayer(&l[L_LAND_1024_B], map_land, mc, 1, 2, 50, p, nil)
		p = SetupLayer(&l[L_LAND_1024_C], map_land, mc, 1, 2, 70, p, nil)
		p = SetupLayer(&l[L_ISLAND_1024], MapIsland, mc, 1, 2, 2, p, nil)
		p = SetupLayer(&l[L_SNOW_1024], MapSnow, mc, 1, 2, 2, p, nil)
		p = SetupLayer(&l[L_LAND_1024_D], map_land, mc, 1, 2, 3, p, nil)
		p = SetupLayer(&l[L_COOL_1024], MapCool, mc, 1, 2, 2, p, nil)
		p = SetupLayer(&l[L_HEAT_1024], MapHeat, mc, 1, 2, 2, p, nil)
		p = SetupLayer(&l[L_SPECIAL_1024], MapSpecial, mc, 1, 2, 3, p, nil)
		p = SetupLayer(&l[L_ZOOM_512], MapZoom, mc, 2, 3, 2002, p, nil)
		p = SetupLayer(&l[L_ZOOM_256], MapZoom, mc, 2, 3, 2003, p, nil)
		p = SetupLayer(&l[L_LAND_256], map_land, mc, 1, 2, 4, p, nil)
		p = SetupLayer(&l[L_MUSHROOM_256], MapMushroom, mc, 1, 2, 5, p, nil)
		p = SetupLayer(&l[L_DEEP_OCEAN_256], MapDeepOcean, mc, 1, 2, 4, p, nil)
		p = SetupLayer(&l[L_BIOME_256], MapBiome, mc, 1, 0, 200, p, nil)
		if mc >= MC_1_14 {
			p = SetupLayer(&l[L_BAMBOO_256], MapBamboo, mc, 1, 0, 1001, p, nil)
		}
		p = SetupLayer(&l[L_ZOOM_128], MapZoom, mc, 2, 3, 1000, p, nil)
		p = SetupLayer(&l[L_ZOOM_64], MapZoom, mc, 2, 3, 1001, p, nil)
		p = SetupLayer(&l[L_BIOME_EDGE_64], MapBiomeEdge, mc, 1, 2, 1000, p, nil)
		SetupLayer(&l[L_RIVER_INIT_256], MapNoise, mc, 1, 0, 100, &l[L_DEEP_OCEAN_256], nil)
	}

	var p_hills *Layer
	if mc <= MC_1_0 {
	} else if mc <= MC_1_12 {
		p_hills = SetupLayer(&l[L_ZOOM_128_HILLS], MapZoom, mc, 2, 3, 0, &l[L_ZOOM_128], nil)
		p_hills = SetupLayer(&l[L_ZOOM_64_HILLS], MapZoom, mc, 2, 3, 0, p_hills, nil)
	} else {
		p_hills = SetupLayer(&l[L_ZOOM_128_HILLS], MapZoom, mc, 2, 3, 1000, &l[L_ZOOM_128], nil)
		p_hills = SetupLayer(&l[L_ZOOM_64_HILLS], MapZoom, mc, 2, 3, 1001, p_hills, nil)
	}

	var p_final *Layer
	if mc <= MC_1_0 {
		p_final = SetupLayer(&l[L_ZOOM_32], MapZoom, mc, 2, 3, 1000, &l[L_ZOOM_64], nil)
		p_final = SetupLayer(&l[L_LAND_32], map_land, mc, 1, 2, 3, p_final, nil)
		p_final = SetupLayer(&l[L_SHORE_16], MapShore, mc, 1, 2, 1000, p_final, nil)
		p_final = SetupLayer(&l[L_ZOOM_16], MapZoom, mc, 2, 3, 1001, p_final, nil)
		p_final = SetupLayer(&l[L_ZOOM_8], MapZoom, mc, 2, 3, 1002, p_final, nil)
		p_final = SetupLayer(&l[L_ZOOM_4], MapZoom, mc, 2, 3, 1003, p_final, nil)
		p_final = SetupLayer(&l[L_SMOOTH_4], MapSmooth, mc, 1, 2, 1000, p_final, nil)

		p_river := SetupLayer(&l[L_ZOOM_128_RIVER], MapZoom, mc, 2, 3, 1000, &l[L_RIVER_INIT_256], nil)
		p_river = SetupLayer(&l[L_ZOOM_64_RIVER], MapZoom, mc, 2, 3, 1001, p_river, nil)
		p_river = SetupLayer(&l[L_ZOOM_32_RIVER], MapZoom, mc, 2, 3, 1002, p_river, nil)
		p_river = SetupLayer(&l[L_ZOOM_16_RIVER], MapZoom, mc, 2, 3, 1003, p_river, nil)
		p_river = SetupLayer(&l[L_ZOOM_8_RIVER], MapZoom, mc, 2, 3, 1004, p_river, nil)
		p_river = SetupLayer(&l[L_ZOOM_4_RIVER], MapZoom, mc, 2, 3, 1005, p_river, nil)
		p_river = SetupLayer(&l[L_RIVER_4], MapRiver, mc, 1, 2, 1, p_river, nil)
		p_river = SetupLayer(&l[L_SMOOTH_4_RIVER], MapSmooth, mc, 1, 2, 1000, p_river, nil)
	} else if mc <= MC_1_6 {
		p_final = SetupLayer(&l[L_HILLS_64], MapHills, mc, 1, 2, 1000, &l[L_ZOOM_64], p_hills)
		p_final = SetupLayer(&l[L_ZOOM_32], MapZoom, mc, 2, 3, 1000, p_final, nil)
		p_final = SetupLayer(&l[L_LAND_32], map_land, mc, 1, 2, 3, p_final, nil)
		p_final = SetupLayer(&l[L_ZOOM_16], MapZoom, mc, 2, 3, 1001, p_final, nil)
		p_final = SetupLayer(&l[L_SHORE_16], MapShore, mc, 1, 2, 1000, p_final, nil)
		p_final = SetupLayer(&l[L_SWAMP_RIVER_16], MapSwampRiver, mc, 1, 0, 1000, p_final, nil)
		p_final = SetupLayer(&l[L_ZOOM_8], MapZoom, mc, 2, 3, 1002, p_final, nil)
		p_final = SetupLayer(&l[L_ZOOM_4], MapZoom, mc, 2, 3, 1003, p_final, nil)
		if largeBiomes {
			p_final = SetupLayer(&l[L_ZOOM_LARGE_A], MapZoom, mc, 2, 3, 1004, p_final, nil)
			p_final = SetupLayer(&l[L_ZOOM_LARGE_B], MapZoom, mc, 2, 3, 1005, p_final, nil)
		}
		p_final = SetupLayer(&l[L_SMOOTH_4], MapSmooth, mc, 1, 2, 1000, p_final, nil)

		p_river := SetupLayer(&l[L_ZOOM_128_RIVER], MapZoom, mc, 2, 3, 1000, &l[L_RIVER_INIT_256], nil)
		p_river = SetupLayer(&l[L_ZOOM_64_RIVER], MapZoom, mc, 2, 3, 1001, p_river, nil)
		p_river = SetupLayer(&l[L_ZOOM_32_RIVER], MapZoom, mc, 2, 3, 1002, p_river, nil)
		p_river = SetupLayer(&l[L_ZOOM_16_RIVER], MapZoom, mc, 2, 3, 1003, p_river, nil)
		p_river = SetupLayer(&l[L_ZOOM_8_RIVER], MapZoom, mc, 2, 3, 1004, p_river, nil)
		p_river = SetupLayer(&l[L_ZOOM_4_RIVER], MapZoom, mc, 2, 3, 1005, p_river, nil)
		if largeBiomes {
			p_river = SetupLayer(&l[L_ZOOM_L_RIVER_A], MapZoom, mc, 2, 3, 1006, p_river, nil)
			p_river = SetupLayer(&l[L_ZOOM_L_RIVER_B], MapZoom, mc, 2, 3, 1007, p_river, nil)
		}
		p_river = SetupLayer(&l[L_RIVER_4], MapRiver, mc, 1, 2, 1, p_river, nil)
		p_river = SetupLayer(&l[L_SMOOTH_4_RIVER], MapSmooth, mc, 1, 2, 1000, p_river, nil)
	} else {
		p_final = SetupLayer(&l[L_HILLS_64], MapHills, mc, 1, 2, 1000, &l[L_BIOME_EDGE_64], p_hills)
		p_final = SetupLayer(&l[L_SUNFLOWER_64], MapSunflower, mc, 1, 0, 1001, p_final, nil)
		p_final = SetupLayer(&l[L_ZOOM_32], MapZoom, mc, 2, 3, 1000, p_final, nil)
		p_final = SetupLayer(&l[L_LAND_32], map_land, mc, 1, 2, 3, p_final, nil)
		p_final = SetupLayer(&l[L_ZOOM_16], MapZoom, mc, 2, 3, 1001, p_final, nil)
		p_final = SetupLayer(&l[L_SHORE_16], MapShore, mc, 1, 2, 1000, p_final, nil)
		p_final = SetupLayer(&l[L_ZOOM_8], MapZoom, mc, 2, 3, 1002, p_final, nil)
		p_final = SetupLayer(&l[L_ZOOM_4], MapZoom, mc, 2, 3, 1003, p_final, nil)
		if largeBiomes {
			p_final = SetupLayer(&l[L_ZOOM_LARGE_A], MapZoom, mc, 2, 3, 1004, p_final, nil)
			p_final = SetupLayer(&l[L_ZOOM_LARGE_B], MapZoom, mc, 2, 3, 1005, p_final, nil)
		}
		p_final = SetupLayer(&l[L_SMOOTH_4], MapSmooth, mc, 1, 2, 1000, p_final, nil)

		p_river := SetupLayer(&l[L_ZOOM_128_RIVER], MapZoom, mc, 2, 3, 1000, &l[L_RIVER_INIT_256], nil)
		p_river = SetupLayer(&l[L_ZOOM_64_RIVER], MapZoom, mc, 2, 3, 1001, p_river, nil)
		p_river = SetupLayer(&l[L_ZOOM_32_RIVER], MapZoom, mc, 2, 3, 1000, p_river, nil)
		p_river = SetupLayer(&l[L_ZOOM_16_RIVER], MapZoom, mc, 2, 3, 1001, p_river, nil)
		p_river = SetupLayer(&l[L_ZOOM_8_RIVER], MapZoom, mc, 2, 3, 1002, p_river, nil)
		p_river = SetupLayer(&l[L_ZOOM_4_RIVER], MapZoom, mc, 2, 3, 1003, p_river, nil)
		if largeBiomes && mc == MC_1_7 {
			p_river = SetupLayer(&l[L_ZOOM_L_RIVER_A], MapZoom, mc, 2, 3, 1004, p_river, nil)
			p_river = SetupLayer(&l[L_ZOOM_L_RIVER_B], MapZoom, mc, 2, 3, 1005, p_river, nil)
		}
		p_river = SetupLayer(&l[L_RIVER_4], MapRiver, mc, 1, 2, 1, p_river, nil)
		p_river = SetupLayer(&l[L_SMOOTH_4_RIVER], MapSmooth, mc, 1, 2, 1000, p_river, nil)
	}

	p_mix := SetupLayer(&l[L_RIVER_MIX_4], MapRiverMix, mc, 1, 0, 100, &l[L_SMOOTH_4], &l[L_SMOOTH_4_RIVER])

	if mc <= MC_1_12 {
		p_final = SetupLayer(&l[L_VORONOI_1], MapVoronoi114, mc, 4, 3, 10, p_mix, nil)
	} else {
		p_ocean := SetupLayer(&l[L_OCEAN_TEMP_256], MapOceanTemp, mc, 1, 0, 2, nil, nil)
		l[L_OCEAN_TEMP_256].Noise = &g.OceanRnd
		p_ocean = SetupLayer(&l[L_ZOOM_128_OCEAN], MapZoom, mc, 2, 3, 2001, p_ocean, nil)
		p_ocean = SetupLayer(&l[L_ZOOM_64_OCEAN], MapZoom, mc, 2, 3, 2002, p_ocean, nil)
		p_ocean = SetupLayer(&l[L_ZOOM_32_OCEAN], MapZoom, mc, 2, 3, 2003, p_ocean, nil)
		p_ocean = SetupLayer(&l[L_ZOOM_16_OCEAN], MapZoom, mc, 2, 3, 2004, p_ocean, nil)
		p_ocean = SetupLayer(&l[L_ZOOM_8_OCEAN], MapZoom, mc, 2, 3, 2005, p_ocean, nil)
		p_ocean = SetupLayer(&l[L_ZOOM_4_OCEAN], MapZoom, mc, 2, 3, 2006, p_ocean, nil)
		p_mix = SetupLayer(&l[L_OCEAN_MIX_4], MapOceanMix, mc, 1, 17, 100, p_mix, p_ocean)

		if mc <= MC_1_14 {
			p_final = SetupLayer(&l[L_VORONOI_1], MapVoronoi114, mc, 4, 3, 10, p_mix, nil)
		} else {
			p_final = SetupLayer(&l[L_VORONOI_1], MapVoronoi, mc, 4, 3, math.MaxUint64, p_mix, nil)
		}
	}

	g.Entry1 = p_final
	if mc <= MC_1_12 {
		g.Entry4 = &l[L_RIVER_MIX_4]
	} else {
		g.Entry4 = &l[L_OCEAN_MIX_4]
	}

	if largeBiomes {
		g.Entry16 = &l[L_ZOOM_4]
		if mc <= MC_1_6 {
			g.Entry64 = &l[L_SWAMP_RIVER_16]
			g.Entry256 = &l[L_HILLS_64]
		} else {
			g.Entry64 = &l[L_SHORE_16]
			g.Entry256 = &l[L_SUNFLOWER_64]
		}
	} else if mc >= MC_1_1 {
		if mc <= MC_1_6 {
			g.Entry16 = &l[L_SWAMP_RIVER_16]
			g.Entry64 = &l[L_HILLS_64]
		} else {
			g.Entry16 = &l[L_SHORE_16]
			g.Entry64 = &l[L_HILLS_64]
		}
		if mc <= MC_1_14 {
			g.Entry256 = &l[L_BIOME_256]
		} else {
			g.Entry256 = &l[L_BAMBOO_256]
		}
	} else {
		g.Entry16 = &l[L_ZOOM_16]
		g.Entry64 = &l[L_ZOOM_64]
		g.Entry256 = &l[L_BIOME_256]
	}
	setupScale(g.Entry1, 1)
}
