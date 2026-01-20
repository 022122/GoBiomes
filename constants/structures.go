package constants

// StructureType 对应 cubiomes 的 enum StructureType。
//
// 注意：数值必须与 [`GO/cubiomes/finders.h:15`](GO/cubiomes/finders.h:15) 一致。
// Feature(0) 是 1.13 前用于神庙生成尝试的位置类型。
//
// 为了与原 Python 用法一致，这里导出与 C 枚举同名（驼峰）的常量。
// 例如：Outpost、Village、TrialChambers 等。
//
// 参考：[`src/modules/structures.c:1`](src/modules/structures.c:1)
//
//go:generate echo "no codegen"

type StructureType int

const (
	Feature       StructureType = 0
	DesertPyramid StructureType = 1
	JungleTemple  StructureType = 2
	JunglePyramid StructureType = 2
	SwampHut      StructureType = 3
	Igloo         StructureType = 4
	Village       StructureType = 5
	OceanRuin     StructureType = 6
	Shipwreck     StructureType = 7
	Monument      StructureType = 8
	Mansion       StructureType = 9
	Outpost       StructureType = 10
	RuinedPortal  StructureType = 11
	RuinedPortalN StructureType = 12
	AncientCity   StructureType = 13
	Treasure      StructureType = 14
	Mineshaft     StructureType = 15
	DesertWell    StructureType = 16
	Geode         StructureType = 17
	Fortress      StructureType = 18
	Bastion       StructureType = 19
	EndCity       StructureType = 20
	EndGateway    StructureType = 21
	EndIsland     StructureType = 22
	TrailRuins    StructureType = 23
	TrialChambers StructureType = 24
)
