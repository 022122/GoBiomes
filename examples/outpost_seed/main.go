//go:build cgo

package main

import (
	"fmt"

	"github.com/scriptlinestudios/gobiomes"
	"github.com/scriptlinestudios/gobiomes/constants"
)

// 复刻 [`README.md`](README.md:31) 的 Outpost 搜索逻辑（示例）。
//
// 说明：该示例用于演示 API 调用方式，真实大范围搜索会非常耗时。
func main() {
	finder := gobiomes.NewFinder(constants.MC_1_21_1)
	gen := gobiomes.NewGenerator(constants.MC_1_21_1, 0)

	for lower48 := uint64(0); lower48 < 1000000; lower48++ {
		pos, err := finder.GetStructurePos(int(constants.Outpost), lower48, 0, 0)
		if err != nil {
			panic(err)
		}
		if pos == nil {
			continue
		}

		if pos.X >= 16 || pos.Z >= 16 {
			continue
		}

		for upper16 := uint64(0); upper16 < 0x10000; upper16++ {
			seed := lower48 | (upper16 << 48)
			gen.ApplySeed(seed, int(constants.DimOverworld))

			if gen.IsViableStructurePos(int(constants.Outpost), pos.X, pos.Z, 0) {
				fmt.Println(seed)
				return
			}
		}
	}
}
