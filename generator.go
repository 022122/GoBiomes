//go:build cgo

package gobiomes

/*
#include <stdlib.h>

#include "cubiomes/finders.h"
#include "cubiomes/generator.h"
*/
import "C"

import (
	"errors"
	"unsafe"
)

// Generator 是对 cubiomes [`Generator`](GO/cubiomes/generator.h:15) 的 Go 封装。
//
// 对应原项目 Python 用法：
//   - apply_seed
//   - get_biome_at
//   - gen_biomes
//   - is_viable_structure_pos
type Generator struct {
	g C.Generator
}

// NewGenerator 创建并初始化一个生成器。
//
// version 对应 [`constants.MC_*`](GO/constants/versions.go:6)
// flags 对应 cubiomes generator flags（见 [`GO/cubiomes/generator.h:8`](GO/cubiomes/generator.h:8)）。
func NewGenerator(version int, flags uint32) *Generator {
	var gg Generator
	C.setupGenerator(&gg.g, C.int(version), C.uint(flags))
	return &gg
}

// ApplySeed 应用世界种子到指定维度。
//
// dim:
//   - 0: overworld
//   - -1: nether
//   - +1: end
func (gen *Generator) ApplySeed(seed uint64, dim int) {
	C.applySeed(&gen.g, C.int(dim), C.ulonglong(seed))
}

// GetBiomeAt 获取指定坐标处生物群系 ID。
//
// scale 通常使用 1(方块) 或 4(生物群系坐标)。
func (gen *Generator) GetBiomeAt(scale, x, y, z int) int {
	id := C.getBiomeAt(&gen.g, C.int(scale), C.int(x), C.int(y), C.int(z))
	return int(id)
}

// GenBiomes 按 Range 批量生成生物群系。
// 返回的切片长度为 getMinCacheSize()。
func (gen *Generator) GenBiomes(r Range) ([]int, error) {
	cr := r.toC()
	cache := C.allocCache(&gen.g, cr)
	if cache == nil {
		return nil, errors.New("allocCache failed")
	}
	defer C.free(unsafe.Pointer(cache))

	// genBiomes: return 0 on success
	if rc := C.genBiomes(&gen.g, cache, cr); rc != 0 {
		return nil, errors.New("genBiomes failed")
	}

	n := int(C.getMinCacheSize(&gen.g, C.int(r.Scale), C.int(r.SX), C.int(r.SY), C.int(r.SZ)))
	out := make([]int, n)
	copy(out, intsFromC(cache, n))
	return out, nil
}

// IsViableStructurePos 判断结构在指定 block 坐标处是否可能生成。
//
// flags 结构特定参数（例如村庄变体），一般传 0。
func (gen *Generator) IsViableStructurePos(structType int, blockX, blockZ int, flags uint32) bool {
	ret := C.isViableStructurePos(C.int(structType), &gen.g, C.int(blockX), C.int(blockZ), C.uint(flags))
	return ret != 0
}
