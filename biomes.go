package gobiomes

// Biome 生物群系类型
type Biome int

// 所有 Minecraft 生物群系常量
const (
	// None 对应 cubiomes 的 none = -1
	None                          Biome = -1
	Ocean                         Biome = 0
	Plains                        Biome = 1
	Desert                        Biome = 2
	Mountains                     Biome = 3
	ExtremeHills                  Biome = 3
	Forest                        Biome = 4
	Taiga                         Biome = 5
	Swamp                         Biome = 6
	Swampland                     Biome = 6
	River                         Biome = 7
	NetherWastes                  Biome = 8
	Hell                          Biome = 8
	TheEnd                        Biome = 9
	Sky                           Biome = 9
	FrozenOcean                   Biome = 10
	FrozenRiver                   Biome = 11
	SnowyTundra                   Biome = 12
	IcePlains                     Biome = 12
	SnowyMountains                Biome = 13
	IceMountains                  Biome = 13
	MushroomFields                Biome = 14
	MushroomIsland                Biome = 14
	MushroomFieldShore            Biome = 15
	MushroomIslandShore           Biome = 15
	Beach                         Biome = 16
	DesertHills                   Biome = 17
	WoodedHills                   Biome = 18
	ForestHills                   Biome = 18
	TaigaHills                    Biome = 19
	MountainEdge                  Biome = 20
	ExtremeHillsEdge              Biome = 20
	Jungle                        Biome = 21
	JungleHills                   Biome = 22
	JungleEdge                    Biome = 23
	DeepOcean                     Biome = 24
	StoneShore                    Biome = 25
	StoneBeach                    Biome = 25
	SnowyBeach                    Biome = 26
	ColdBeach                     Biome = 26
	BirchForest                   Biome = 27
	BirchForestHills              Biome = 28
	DarkForest                    Biome = 29
	RoofedForest                  Biome = 29
	SnowyTaiga                    Biome = 30
	ColdTaiga                     Biome = 30
	SnowyTaigaHills               Biome = 31
	ColdTaigaHills                Biome = 31
	GiantTreeTaiga                Biome = 32
	MegaTaiga                     Biome = 32
	GiantTreeTaigaHills           Biome = 33
	MegaTaigaHills                Biome = 33
	WoodedMountains               Biome = 34
	ExtremeHillsPlus              Biome = 34
	Savanna                       Biome = 35
	SavannaPlateau                Biome = 36
	Badlands                      Biome = 37
	Mesa                          Biome = 37
	WoodedBadlandsPlateau         Biome = 38
	MesaPlateauF                  Biome = 38
	BadlandsPlateau               Biome = 39
	MesaPlateau                   Biome = 39
	SmallEndIslands               Biome = 40
	EndMidlands                   Biome = 41
	EndHighlands                  Biome = 42
	EndBarrens                    Biome = 43
	WarmOcean                     Biome = 44
	LukewarmOcean                 Biome = 45
	ColdOcean                     Biome = 46
	DeepWarmOcean                 Biome = 47
	WarmDeepOcean                 Biome = 47
	DeepLukewarmOcean             Biome = 48
	LukewarmDeepOcean             Biome = 48
	DeepColdOcean                 Biome = 49
	ColdDeepOcean                 Biome = 49
	DeepFrozenOcean               Biome = 50
	FrozenDeepOcean               Biome = 50
	SeasonalForest                Biome = 51
	Rainforest                    Biome = 52
	Shrubland                     Biome = 53
	TheVoid                       Biome = 127
	SunflowerPlains               Biome = 129
	DesertLakes                   Biome = 130
	GravellyMountains             Biome = 131
	FlowerForest                  Biome = 132
	TaigaMountains                Biome = 133
	SwampHills                    Biome = 134
	IceSpikes                     Biome = 140
	ModifiedJungle                Biome = 149
	ModifiedJungleEdge            Biome = 151
	TallBirchForest               Biome = 155
	TallBirchHills                Biome = 156
	DarkForestHills               Biome = 157
	SnowyTaigaMountains           Biome = 158
	GiantSpruceTaiga              Biome = 160
	GiantSpruceTaigaHills         Biome = 161
	ModifiedGravellyMountains     Biome = 162
	ShatteredSavanna              Biome = 163
	ShatteredSavannaPlateau       Biome = 164
	ErodedBadlands                Biome = 165
	ModifiedWoodedBadlandsPlateau Biome = 166
	ModifiedBadlandsPlateau       Biome = 167
	BambooJungle                  Biome = 168
	BambooJungleHills             Biome = 169
	SoulSandValley                Biome = 170
	CrimsonForest                 Biome = 171
	WarpedForest                  Biome = 172
	BasaltDeltas                  Biome = 173
	DripstoneCaves                Biome = 174
	LushCaves                     Biome = 175
	Meadow                        Biome = 177
	Grove                         Biome = 178
	SnowySlopes                   Biome = 179
	JaggedPeaks                   Biome = 180
	FrozenPeaks                   Biome = 181
	StonyPeaks                    Biome = 182
	OldGrowthBirchForest          Biome = 183
	OldGrowthPineTaiga            Biome = 184
	OldGrowthSpruceTaiga          Biome = 185
	SnowyPlains                   Biome = 186
	SparseJungle                  Biome = 187
	StonyShore                    Biome = 188
	WindsweptHills                Biome = 189
	WindsweptForest               Biome = 190
	WindsweptGravellyHills        Biome = 191
	WindsweptSavanna              Biome = 192
	WoodedBadlands                Biome = 193
	DeepDark                      Biome = 194
	MangroveSwamp                 Biome = 195
	CherryGrove                   Biome = 196
	PaleGarden                    Biome = 197
)

// Biome Categories
const (
	Oceanic  Biome = 0
	Warm     Biome = 1
	Lush     Biome = 2
	Cold     Biome = 3
	Freezing Biome = 4
)

func (b Biome) IsOceanic() bool {
	switch b {
	case Ocean, FrozenOcean, DeepOcean, WarmOcean, LukewarmOcean, ColdOcean,
		DeepWarmOcean, DeepLukewarmOcean, DeepColdOcean, DeepFrozenOcean:
		return true
	}
	return false
}

func (b Biome) IsShallowOcean() bool {
	switch b {
	case Ocean, FrozenOcean, WarmOcean, LukewarmOcean, ColdOcean:
		return true
	}
	return false
}

func (b Biome) IsDeepOcean() bool {
	switch b {
	case DeepOcean, DeepWarmOcean, DeepLukewarmOcean, DeepColdOcean, DeepFrozenOcean:
		return true
	}
	return false
}

func (b Biome) IsSnowy() bool {
	switch b {
	case FrozenOcean, FrozenRiver, SnowyTundra, SnowyMountains, SnowyBeach,
		SnowyTaiga, SnowyTaigaHills, IceSpikes, SnowyTaigaMountains:
		return true
	}
	return false
}

func (b Biome) IsMesa() bool {
	switch b {
	case Badlands, ErodedBadlands, ModifiedWoodedBadlandsPlateau,
		ModifiedBadlandsPlateau, WoodedBadlandsPlateau, BadlandsPlateau:
		return true
	}
	return false
}

func GetCategory(mc int, id Biome) Biome {
	switch id {
	case Beach, SnowyBeach:
		return Beach
	case Desert, DesertHills, DesertLakes:
		return Desert
	case Mountains, MountainEdge, WoodedMountains, GravellyMountains, ModifiedGravellyMountains:
		return Mountains
	case Forest, WoodedHills, BirchForest, BirchForestHills, DarkForest, FlowerForest,
		TallBirchForest, TallBirchHills, DarkForestHills:
		return Forest
	case SnowyTundra, SnowyMountains, IceSpikes:
		return SnowyTundra
	case Jungle, JungleHills, JungleEdge, ModifiedJungle, ModifiedJungleEdge, BambooJungle, BambooJungleHills:
		return Jungle
	case Badlands, ErodedBadlands, ModifiedWoodedBadlandsPlateau, ModifiedBadlandsPlateau:
		return Mesa
	case WoodedBadlandsPlateau, BadlandsPlateau:
		if mc <= MC_1_15 {
			return Mesa
		}
		return BadlandsPlateau
	case MushroomFields, MushroomFieldShore:
		return MushroomFields
	case StoneShore:
		return StoneShore
	case Ocean, FrozenOcean, DeepOcean, WarmOcean, LukewarmOcean, ColdOcean,
		DeepWarmOcean, DeepLukewarmOcean, DeepColdOcean, DeepFrozenOcean:
		return Ocean
	case Plains, SunflowerPlains:
		return Plains
	case River, FrozenRiver:
		return River
	case Savanna, SavannaPlateau, ShatteredSavanna, ShatteredSavannaPlateau:
		return Savanna
	case Swamp, SwampHills:
		return Swamp
	case Taiga, TaigaHills, SnowyTaiga, SnowyTaigaHills, GiantTreeTaiga, GiantTreeTaigaHills,
		TaigaMountains, SnowyTaigaMountains, GiantSpruceTaiga, GiantSpruceTaigaHills:
		return Taiga
	case NetherWastes, SoulSandValley, CrimsonForest, WarpedForest, BasaltDeltas:
		return NetherWastes
	default:
		return None
	}
}

func AreSimilar(mc int, id1, id2 Biome) bool {
	if id1 == id2 {
		return true
	}
	if mc <= MC_1_15 {
		if (id1 == WoodedBadlandsPlateau || id1 == BadlandsPlateau) &&
			(id2 == WoodedBadlandsPlateau || id2 == BadlandsPlateau) {
			return true
		}
	}
	return GetCategory(mc, id1) == GetCategory(mc, id2)
}

func GetMutated(mc int, id Biome) Biome {
	switch id {
	case Plains:
		return SunflowerPlains
	case Desert:
		return DesertLakes
	case Mountains:
		return GravellyMountains
	case Forest:
		return FlowerForest
	case Taiga:
		return TaigaMountains
	case Swamp:
		return SwampHills
	case SnowyTundra:
		return IceSpikes
	case Jungle:
		return ModifiedJungle
	case JungleEdge:
		return ModifiedJungleEdge
	case BirchForest:
		if mc >= MC_1_9 && mc <= MC_1_10 {
			return TallBirchHills
		}
		return TallBirchForest
	case BirchForestHills:
		if mc >= MC_1_9 && mc <= MC_1_10 {
			return None
		}
		return TallBirchHills
	case DarkForest:
		return DarkForestHills
	case SnowyTaiga:
		return SnowyTaigaMountains
	case GiantTreeTaiga:
		return GiantSpruceTaiga
	case GiantTreeTaigaHills:
		return GiantSpruceTaigaHills
	case WoodedMountains:
		return ModifiedGravellyMountains
	case Savanna:
		return ShatteredSavanna
	case SavannaPlateau:
		return ShatteredSavannaPlateau
	case Badlands:
		return ErodedBadlands
	case WoodedBadlandsPlateau:
		return ModifiedWoodedBadlandsPlateau
	case BadlandsPlateau:
		return ModifiedBadlandsPlateau
	default:
		return None
	}
}
