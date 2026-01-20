package gobiomes

import (
	"fmt"
)

// StructureConfig 对应 cubiomes 的 StructureConfig。
type StructureConfig struct {
	Salt       int32
	RegionSize int
	ChunkRange int
	StructType StructureType
	Dim        Dimension
	Rarity     float32
}

// Finder 是对结构定位相关逻辑的封装。
type Finder struct {
	Version int
}

// NewFinder 创建 Finder。
func NewFinder(version int) *Finder {
	return &Finder{Version: version}
}

// GetStructureConfig 返回指定结构在该版本下的配置。
func (f *Finder) GetStructureConfig(st StructureType) (StructureConfig, error) {
	mc := f.Version
	var sconf StructureConfig
	found := false

	// 对应 finders.c 中的静态配置
	s_feature := StructureConfig{14357617, 32, 24, Feature, 0, 0}
	s_igloo_112 := StructureConfig{14357617, 32, 24, Igloo, 0, 0}
	s_swamp_hut_112 := StructureConfig{14357617, 32, 24, SwampHut, 0, 0}
	s_desert_pyramid_112 := StructureConfig{14357617, 32, 24, DesertPyramid, 0, 0}
	s_jungle_temple_112 := StructureConfig{14357617, 32, 24, JunglePyramid, 0, 0}
	s_ocean_ruin_115 := StructureConfig{14357621, 16, 8, OceanRuin, 0, 0}
	s_shipwreck_115 := StructureConfig{165745295, 16, 8, Shipwreck, 0, 0}
	s_desert_pyramid := StructureConfig{14357617, 32, 24, DesertPyramid, 0, 0}
	s_igloo := StructureConfig{14357618, 32, 24, Igloo, 0, 0}
	s_jungle_temple := StructureConfig{14357619, 32, 24, JunglePyramid, 0, 0}
	s_swamp_hut := StructureConfig{14357620, 32, 24, SwampHut, 0, 0}
	s_outpost := StructureConfig{165745296, 32, 24, Outpost, 0, 0}
	s_village_117 := StructureConfig{10387312, 32, 24, Village, 0, 0}
	s_village := StructureConfig{10387312, 34, 26, Village, 0, 0}
	s_ocean_ruin := StructureConfig{14357621, 20, 12, OceanRuin, 0, 0}
	s_shipwreck := StructureConfig{165745295, 24, 20, Shipwreck, 0, 0}
	s_monument_117 := StructureConfig{10387313, 32, 27, Monument, 0, 0}
	s_monument := StructureConfig{10387313, 32, 24, Monument, 0, 0}
	s_mansion := StructureConfig{10387319, 80, 60, Mansion, 0, 0}
	s_ruined_portal := StructureConfig{34222645, 40, 25, RuinedPortal, 0, 0}
	s_ruined_portal_n := StructureConfig{34222645, 40, 25, RuinedPortal, DimNether, 0}
	s_ruined_portal_n_117 := StructureConfig{34222645, 25, 15, RuinedPortalN, DimNether, 0}
	s_ancient_city := StructureConfig{20083232, 24, 16, AncientCity, 0, 0}
	s_trail_ruins := StructureConfig{83469867, 34, 26, TrailRuins, 0, 0}
	s_trial_chambers := StructureConfig{94251327, 34, 22, TrialChambers, 0, 0}
	s_treasure := StructureConfig{10387320, 1, 1, Treasure, 0, 0}
	s_mineshaft := StructureConfig{0, 1, 1, Mineshaft, 0, 0}
	s_desert_well_115 := StructureConfig{30010, 1, 1, DesertWell, 0, 1.0 / 1000.0}
	s_desert_well_117 := StructureConfig{40013, 1, 1, DesertWell, 0, 1.0 / 1000.0}
	s_desert_well := StructureConfig{40002, 1, 1, DesertWell, 0, 1.0 / 1000.0}
	s_geode_117 := StructureConfig{20000, 1, 1, Geode, 0, 1.0 / 24.0}
	s_geode := StructureConfig{20002, 1, 1, Geode, 0, 1.0 / 24.0}
	s_fortress_115 := StructureConfig{0, 16, 8, Fortress, DimNether, 0}
	s_fortress := StructureConfig{30084232, 27, 23, Fortress, DimNether, 0}
	s_bastion := StructureConfig{30084232, 27, 23, Bastion, DimNether, 0}
	s_end_city := StructureConfig{10387313, 20, 9, EndCity, DimEnd, 0}

	switch st {
	case Feature:
		sconf = s_feature
		found = mc <= MC_1_12
	case DesertPyramid:
		if mc <= MC_1_12 {
			sconf = s_desert_pyramid_112
		} else {
			sconf = s_desert_pyramid
		}
		found = mc >= MC_1_3
	case JunglePyramid:
		if mc <= MC_1_12 {
			sconf = s_jungle_temple_112
		} else {
			sconf = s_jungle_temple
		}
		found = mc >= MC_1_3
	case SwampHut:
		if mc <= MC_1_12 {
			sconf = s_swamp_hut_112
		} else {
			sconf = s_swamp_hut
		}
		found = mc >= MC_1_4
	case Igloo:
		if mc <= MC_1_12 {
			sconf = s_igloo_112
		} else {
			sconf = s_igloo
		}
		found = mc >= MC_1_9
	case Village:
		if mc <= MC_1_17 {
			sconf = s_village_117
		} else {
			sconf = s_village
		}
		found = mc >= MC_B1_8
	case OceanRuin:
		if mc <= MC_1_15 {
			sconf = s_ocean_ruin_115
		} else {
			sconf = s_ocean_ruin
		}
		found = mc >= MC_1_13
	case Shipwreck:
		if mc <= MC_1_15 {
			sconf = s_shipwreck_115
		} else {
			sconf = s_shipwreck
		}
		found = mc >= MC_1_13
	case RuinedPortal:
		sconf = s_ruined_portal
		found = mc >= MC_1_16_1
	case RuinedPortalN:
		if mc <= MC_1_17 {
			sconf = s_ruined_portal_n_117
		} else {
			sconf = s_ruined_portal_n
		}
		found = mc >= MC_1_16_1
	case Monument:
		if mc <= MC_1_17 {
			sconf = s_monument_117
		} else {
			sconf = s_monument
		}
		found = mc >= MC_1_8
	case EndCity:
		sconf = s_end_city
		found = mc >= MC_1_9
	case Mansion:
		sconf = s_mansion
		found = mc >= MC_1_11
	case Outpost:
		sconf = s_outpost
		found = mc >= MC_1_14
	case AncientCity:
		sconf = s_ancient_city
		found = mc >= MC_1_19_2
	case Treasure:
		sconf = s_treasure
		found = mc >= MC_1_13
	case Mineshaft:
		sconf = s_mineshaft
		found = mc >= MC_B1_8
	case Fortress:
		if mc <= MC_1_15 {
			sconf = s_fortress_115
		} else {
			sconf = s_fortress
		}
		found = mc >= MC_1_0
	case Bastion:
		sconf = s_bastion
		found = mc >= MC_1_16_1
	case DesertWell:
		if mc <= MC_1_15 {
			sconf = s_desert_well_115
		} else if mc <= MC_1_17 {
			sconf = s_desert_well_117
		} else {
			sconf = s_desert_well
		}
		found = mc >= MC_1_13
	case Geode:
		if mc <= MC_1_17 {
			sconf = s_geode_117
		} else {
			sconf = s_geode
		}
		found = mc >= MC_1_17
	case TrailRuins:
		sconf = s_trail_ruins
		found = mc >= MC_1_20
	case TrialChambers:
		sconf = s_trial_chambers
		found = mc >= MC_1_21_1
	}

	if !found {
		return StructureConfig{}, fmt.Errorf("structure type %v not supported in version %v", st, mc)
	}
	return sconf, nil
}

// getFeatureChunkInRegion 计算结构在 region 内的 chunk 偏移。
func getFeatureChunkInRegion(config StructureConfig, seed uint64, regX, regZ int) (int, int) {
	const (
		K = 0x5deece66d
		M = (1 << 48) - 1
		b = 0xb
	)

	// set seed
	s := seed + uint64(regX)*341873128712 + uint64(regZ)*132897987541 + uint64(config.Salt)
	s = (s ^ K)
	s = (s*K + b) & M

	r := uint64(config.ChunkRange)
	var px, pz int
	if r&(r-1) != 0 {
		px = int((s >> 17) % r)
		s = (s*K + b) & M
		pz = int((s >> 17) % r)
	} else {
		px = int((r * (s >> 17)) >> 31)
		s = (s*K + b) & M
		pz = int((r * (s >> 17)) >> 31)
	}
	return px, pz
}

// getFeaturePos 计算结构的 block 坐标。
func getFeaturePos(config StructureConfig, seed uint64, regX, regZ int) Pos {
	px, pz := getFeatureChunkInRegion(config, seed, regX, regZ)
	return Pos{
		X: (regX*config.RegionSize + px) << 4,
		Z: (regZ*config.RegionSize + pz) << 4,
	}
}

// getLargeStructureChunkInRegion 计算大型结构（神庙、府邸）在 region 内的 chunk 偏移。
func getLargeStructureChunkInRegion(config StructureConfig, seed uint64, regX, regZ int) (int, int) {
	const (
		K = 0x5deece66d
		M = (1 << 48) - 1
		b = 0xb
	)

	s := seed + uint64(regX)*341873128712 + uint64(regZ)*132897987541 + uint64(config.Salt)
	s = (s ^ K)

	s = (s*K + b) & M
	px := int((s >> 17) % uint64(config.ChunkRange))
	s = (s*K + b) & M
	px += int((s >> 17) % uint64(config.ChunkRange))

	s = (s*K + b) & M
	pz := int((s >> 17) % uint64(config.ChunkRange))
	s = (s*K + b) & M
	pz += int((s >> 17) % uint64(config.ChunkRange))

	return px >> 1, pz >> 1
}

// getLargeStructurePos 计算大型结构的 block 坐标。
func getLargeStructurePos(config StructureConfig, seed uint64, regX, regZ int) Pos {
	px, pz := getLargeStructureChunkInRegion(config, seed, regX, regZ)
	return Pos{
		X: (regX*config.RegionSize + px) << 4,
		Z: (regZ*config.RegionSize + pz) << 4,
	}
}

// setAttemptSeed 对应 finders.c 中的 setAttemptSeed。
func setAttemptSeed(s *uint64, cx, cz int) {
	*s ^= uint64(cx>>4) ^ (uint64(cz>>4) << 4)
	*s = (*s ^ 0x5deece66d) & mask48
	// next(s, 31)
	*s = (*s*0x5deece66d + 0xb) & mask48
}

// ChunkGenerateRnd 返回 chunk 生成用的 48-bit RNG seed。
func (f *Finder) ChunkGenerateRnd(worldSeed uint64, chunkX, chunkZ int) uint64 {
	r := NewRng(worldSeed)
	a := uint64(r.NextLong())
	b := uint64(r.NextLong())
	rnd := (a*uint64(chunkX) ^ b*uint64(chunkZ) ^ worldSeed) & mask48
	r.SetSeed(rnd)
	return r.seed
}

// GetStructurePos 返回结构在指定 region 内的生成尝试位置。
func (f *Finder) GetStructurePos(st StructureType, seed uint64, regX, regZ int) (*Pos, error) {
	config, err := f.GetStructureConfig(st)
	if err != nil {
		return nil, err
	}

	seed &= mask48
	var pos Pos

	switch st {
	case Feature, DesertPyramid, JunglePyramid, SwampHut,
		Igloo, Village, OceanRuin, Shipwreck,
		RuinedPortal, RuinedPortalN, AncientCity,
		TrailRuins, TrialChambers:
		pos = getFeaturePos(config, seed, regX, regZ)
		return &pos, nil

	case Monument:
		if f.Version >= MC_1_18 {
			pos = getFeaturePos(config, seed, regX, regZ)
		} else {
			pos = getLargeStructurePos(config, seed, regX, regZ)
		}
		return &pos, nil

	case Mansion:
		pos = getLargeStructurePos(config, seed, regX, regZ)
		return &pos, nil

	case EndCity:
		pos = getLargeStructurePos(config, seed, regX, regZ)
		if int64(pos.X)*int64(pos.X)+int64(pos.Z)*int64(pos.Z) < 1008*1008 {
			return nil, nil
		}
		return &pos, nil

	case Outpost:
		pos = getFeaturePos(config, seed, regX, regZ)
		s := seed
		setAttemptSeed(&s, pos.X>>4, pos.Z>>4)
		// nextInt(&s, 5) == 0
		r := &Rng{seed: s}
		if r.NextInt(5) == 0 {
			return &pos, nil
		}
		return nil, nil

	case Treasure:
		pos.X = regX*16 + 9
		pos.Z = regZ*16 + 9
		s := uint64(regX)*341873128712 + uint64(regZ)*132897987541 + seed + uint64(config.Salt)
		r := NewRng(s)
		if r.NextFloat() < 0.01 {
			return &pos, nil
		}
		return nil, nil

	case Mineshaft:
		res := f.GetMineshafts(seed, regX, regZ, 1, 1, 1)
		if len(res) > 0 {
			return &res[0], nil
		}
		return nil, nil

	case Fortress:
		if f.Version >= MC_1_18 {
			pos = getFeaturePos(config, seed, regX, regZ)
			return &pos, nil
		} else if f.Version >= MC_1_16_1 {
			// getRegPos 逻辑
			s := seed + uint64(regX)*341873128712 + uint64(regZ)*132897987541 + uint64(config.Salt)
			r := NewRng(s)
			pos.X = (regX*config.RegionSize + r.NextInt(config.ChunkRange)) << 4
			pos.Z = (regZ*config.RegionSize + r.NextInt(config.ChunkRange)) << 4
			if r.NextInt(5) < 2 {
				return &pos, nil
			}
			return nil, nil
		} else {
			s := seed
			setAttemptSeed(&s, regX*16, regZ*16)
			r := &Rng{seed: s}
			valid := r.NextInt(3) == 0
			pos.X = (regX*16 + r.NextInt(8) + 4) * 16
			pos.Z = (regZ*16 + r.NextInt(8) + 4) * 16
			if valid {
				return &pos, nil
			}
			return nil, nil
		}

	case Bastion:
		if f.Version >= MC_1_18 {
			pos = getFeaturePos(config, seed, regX, regZ)
			s := f.ChunkGenerateRnd(seed, pos.X>>4, pos.Z>>4)
			r := &Rng{seed: s}
			if r.NextInt(5) >= 2 {
				return &pos, nil
			}
			return nil, nil
		} else {
			s := seed + uint64(regX)*341873128712 + uint64(regZ)*132897987541 + uint64(config.Salt)
			r := NewRng(s)
			pos.X = (regX*config.RegionSize + r.NextInt(config.ChunkRange)) << 4
			pos.Z = (regZ*config.RegionSize + r.NextInt(config.ChunkRange)) << 4
			if r.NextInt(5) >= 2 {
				return &pos, nil
			}
			return nil, nil
		}

	case Stronghold:
		return nil, fmt.Errorf("Stronghold search requires specialized logic (not region-based)")

	default:
		return nil, fmt.Errorf("GetStructurePos not implemented for %v in pure Go", st)
	}
}

// GetMineshafts 在指定 chunk 范围内查找废弃矿井。
func (f *Finder) GetMineshafts(seed uint64, chunkX, chunkZ, chunkW, chunkH, maxCount int) []Pos {
	return getMineshaftsGo(f.Version, seed, chunkX, chunkZ, chunkX+chunkW-1, chunkZ+chunkH-1, maxCount)
}

// getMineshaftsGo 内部实现。
func getMineshaftsGo(mc int, seed uint64, cx0, cz0, cx1, cz1, nout int) []Pos {
	r := NewRng(seed)
	a := uint64(r.NextLong())
	b := uint64(r.NextLong())
	var out []Pos

	for i := cx0; i <= cx1; i++ {
		aix := uint64(i)*a ^ seed
		for j := cz0; j <= cz1; j++ {
			r.SetSeed(aix ^ uint64(j)*b)
			if mc >= MC_1_13 {
				if r.NextDouble() < 0.004 {
					out = append(out, Pos{X: i * 16, Z: j * 16})
					if len(out) >= nout {
						return out
					}
				}
			} else {
				r.SkipNextN(1)
				if r.NextDouble() < 0.004 {
					d := i
					if -i > d {
						d = -i
					}
					if j > d {
						d = j
					}
					if -j > d {
						d = -j
					}
					if d >= 80 || r.NextInt(80) < d {
						out = append(out, Pos{X: i * 16, Z: j * 16})
						if len(out) >= nout {
							return out
						}
					}
				}
			}
		}
	}
	return out
}
