package main

import (
	"fmt"
	"gobiomes"
)

func main() {
	mc := gobiomes.MC_1_21_1
	seed := uint64(12345)

	finder := gobiomes.NewFinder(mc)
	gen := gobiomes.NewGenerator(mc, 0)
	gen.ApplySeed(seed, gobiomes.DimOverworld)

	// 搜索村庄
	fmt.Printf("正在搜索种子 %d (版本 %d) 中的村庄...\n", seed, mc)

	for rz := -10; rz <= 10; rz++ {
		for rx := -10; rx <= 10; rx++ {
			pos, err := finder.GetStructurePos(gobiomes.Village, seed, rx, rz)
			if err != nil {
				continue
			}
			if pos != nil {
				// 验证生物群系
				if gen.IsViableStructurePos(gobiomes.Village, pos.X, pos.Z, 0) {
					biome := gen.GetBiomeAt(1, pos.X, 64, pos.Z)
					fmt.Printf("找到村庄: x=%d, z=%d, 生物群系=%d\n", pos.X, pos.Z, biome)
				}
			}
		}
	}

	// 1.17 版本的生物群系采样示例
	mc117 := gobiomes.MC_1_17_1
	gen117 := gobiomes.NewGenerator(mc117, 0)
	gen117.ApplySeed(seed, gobiomes.DimOverworld)
	biome117 := gen117.GetBiomeAt(1, 100, 64, 100)
	fmt.Printf("\n1.17.1 种子 %d 在 (100, 100) 的生物群系: %d\n", seed, biome117)

	// 1.18 版本的生物群系采样示例
	mc118 := gobiomes.MC_1_18
	gen118 := gobiomes.NewGenerator(mc118, 0)
	gen118.ApplySeed(seed, gobiomes.DimOverworld)
	biome118 := gen118.GetBiomeAt(1, 100, 64, 100)
	fmt.Printf("1.18 种子 %d 在 (100, 100) 的生物群系: %d\n", seed, biome118)
}
