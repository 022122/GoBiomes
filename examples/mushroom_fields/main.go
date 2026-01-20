//go:build cgo

package main

import (
	"fmt"

	"github.com/scriptlinestudios/gobiomes"
	"github.com/scriptlinestudios/gobiomes/constants"
)

// 对应原项目 README 的“Searching for mushroom islands”示例。
// 参考：[`README.md:15`](../../README.md:15)
func main() {
	gen := gobiomes.NewGenerator(constants.MC_1_21_1, 0)

	for seed := uint64(0); seed < 100000; seed++ {
		gen.ApplySeed(seed, int(constants.DimOverworld))
		biomeID := gen.GetBiomeAt(1, 0, 60, 0)
		if biomeID == int(constants.MushroomFields) {
			fmt.Println(seed)
			return
		}
	}
}
