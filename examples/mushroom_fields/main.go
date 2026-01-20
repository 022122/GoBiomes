package main

import (
	"fmt"

	"gobiomes"
)

// 对应原项目 README 的“Searching for mushroom islands”示例。
func main() {
	gen := gobiomes.NewGenerator(gobiomes.MC_1_21_1, 0)

	for seed := uint64(0); seed < 100000; seed++ {
		gen.ApplySeed(seed, gobiomes.DimOverworld)
		biomeID := gen.GetBiomeAt(1, 0, 60, 0)
		if biomeID == gobiomes.MushroomFields {
			fmt.Println(seed)
			return
		}
	}
}
