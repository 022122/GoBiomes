package gobiomes

import (
	"errors"
)

const (
	LARGE_BIOMES         uint32 = 0x0001
	FORCE_OCEAN_VARIANTS uint32 = 0x0002
)

// Generator 是对生物群系生成逻辑的 Go 实现。
type Generator struct {
	Version int
	Seed    uint64
	Dim     Dimension
	Flags   uint32
	SHA     uint64

	// 1.18+
	BN BiomeNoise

	// Pre-1.18
	LS LayerStack
}

// NewGenerator 创建并初始化一个生成器。
func NewGenerator(version int, flags uint32) *Generator {
	gen := &Generator{
		Version: version,
		Flags:   flags,
	}
	if version >= MC_1_18 {
		gen.BN.Init(version)
	} else if version >= MC_B1_8 {
		SetupLayerStack(&gen.LS, version, flags&LARGE_BIOMES != 0)
	}
	return gen
}

// ApplySeed 应用世界种子到指定维度。
func (gen *Generator) ApplySeed(seed uint64, dim Dimension) {
	gen.Seed = seed
	gen.Dim = dim
	if dim == DimOverworld {
		if gen.Version >= MC_1_18 {
			large := 0
			if gen.Flags&LARGE_BIOMES != 0 {
				large = 1
			}
			gen.BN.SetSeed(seed, large)
		} else if gen.Version >= MC_B1_8 {
			SetLayerSeed(gen.LS.Entry1, seed)
		}
	}
	if gen.Version >= MC_1_15 {
		if gen.Version <= MC_1_17 && dim == DimOverworld && gen.LS.Entry1 != nil {
			gen.SHA = gen.LS.Entry1.StartSalt
		} else {
			gen.SHA = GetVoronoiSHA(seed)
		}
	}
}

// GetBiomeAt 获取指定坐标处生物群系 ID。
func (gen *Generator) GetBiomeAt(scale, x, y, z int) Biome {
	if gen.Dim == DimOverworld {
		if gen.Version >= MC_1_18 {
			// 1.18+ 使用多重噪声采样
			// 注意：1.18+ 的 scale 通常是 1:4 (quarter resolution)
			if scale == 1 {
				return Biome(gen.BN.Sample(x, y, z, 0))
			}
			// 简化实现：对于缩放采样，目前直接返回 1:1 结果
			return Biome(gen.BN.Sample(x, y, z, SAMPLE_NO_SHIFT))
		} else if gen.Version >= MC_B1_8 {
			// Pre-1.18 使用 LayerStack
			var entry *Layer
			switch scale {
			case 1:
				entry = gen.LS.Entry1
			case 4:
				entry = gen.LS.Entry4
			case 16:
				entry = gen.LS.Entry16
			case 64:
				entry = gen.LS.Entry64
			case 256:
				entry = gen.LS.Entry256
			default:
				return None
			}
			if entry == nil {
				return None
			}
			out := make([]int, 1)
			entry.GetMap(entry, out, x, z, 1, 1)
			return Biome(out[0])
		}
	}
	return None
}

// GenBiomes 按 Range 批量生成生物群系。
func (gen *Generator) GenBiomes(r Range) ([]int, error) {
	n := r.SX * r.SY * r.SZ
	if n <= 0 {
		return nil, errors.New("invalid range size")
	}
	out := make([]int, n)
	// 简化实现：逐个采样
	// 实际 cubiomes 中会有优化的批量生成逻辑
	for j := 0; j < r.SZ; j++ {
		for i := 0; i < r.SX; i++ {
			for k := 0; k < r.SY; k++ {
				idx := (k*r.SZ+j)*r.SX + i
				out[idx] = int(gen.GetBiomeAt(r.Scale, r.X+i, r.Y+k, r.Z+j))
			}
		}
	}
	return out, nil
}

// IsViableStructurePos 判断结构在指定 block 坐标处是否可能生成。
func (gen *Generator) IsViableStructurePos(stype StructureType, blockX, blockZ int, flags uint32) bool {
	biome := gen.GetBiomeAt(1, blockX, 64, blockZ)

	switch stype {
	case Village:
		if gen.Version >= MC_1_18 {
			return biome == Plains || biome == Desert || biome == Savanna || biome == SnowyPlains || biome == Taiga || biome == Meadow
		}
		return biome == Plains || biome == Desert || biome == Savanna || biome == SnowyTundra || biome == Taiga
	case DesertPyramid:
		return biome == Desert || biome == DesertLakes
	case JungleTemple:
		return biome == Jungle || biome == BambooJungle
	case SwampHut:
		return biome == Swamp
	case Igloo:
		return biome == SnowyTundra || biome == SnowyTaiga
	case OceanRuin:
		return biome.IsOceanic()
	case Shipwreck:
		return biome.IsOceanic() || biome == Beach || biome == SnowyBeach
	case Monument:
		return biome.IsDeepOcean()
	case Mansion:
		return biome == DarkForest || biome == DarkForestHills
	case Outpost:
		return true
	case Fortress:
		return gen.Dim == DimNether
	case Bastion:
		return gen.Dim == DimNether && biome != BasaltDeltas
	case EndCity:
		return gen.Dim == DimEnd && biome == EndHighlands
	case AncientCity:
		return biome == DeepDark
	case TrailRuins:
		return biome == Taiga || biome == SnowyTaiga || biome == OldGrowthBirchForest || biome == OldGrowthPineTaiga || biome == OldGrowthSpruceTaiga || biome == Jungle
	case TrialChambers:
		return true
	}
	return true
}
