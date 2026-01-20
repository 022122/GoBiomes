//go:build cgo

package gobiomes

/*
#include <stdint.h>

#include "cubiomes/finders.h"

// chunkGenerateRnd 在 finders.h 中是 static inline。
// cgo 无法直接引用内联符号，因此这里提供一个可链接的包装。
static inline uint64_t go_chunkGenerateRnd(uint64_t worldSeed, int chunkX, int chunkZ) {
    return chunkGenerateRnd(worldSeed, chunkX, chunkZ);
}
*/
import "C"

import "errors"

// StructureConfig 对应 cubiomes 的 [`StructureConfig`](GO/cubiomes/finders.h:47)。
//
// 注意：字段类型保持与 C 侧一致（int32/int8/uint8/float）。
type StructureConfig struct {
	Salt       int32
	RegionSize int8
	ChunkRange int8
	StructType uint8
	Dim        int8
	Rarity     float32
}

func structureConfigFromC(sc C.StructureConfig) StructureConfig {
	return StructureConfig{
		Salt:       int32(sc.salt),
		RegionSize: int8(sc.regionSize),
		ChunkRange: int8(sc.chunkRange),
		StructType: uint8(sc.structType),
		Dim:        int8(sc.dim),
		Rarity:     float32(sc.rarity),
	}
}

// Finder 是对 cubiomes 结构定位相关函数的轻量封装。
//
// 对应原项目 Python 用法：
//   - get_structure_config
//   - chunk_generate_rnd
//   - get_structure_pos
type Finder struct {
	Version int
}

// NewFinder 创建 Finder。
func NewFinder(version int) *Finder {
	return &Finder{Version: version}
}

// GetStructureConfig 返回指定结构在该版本下的配置。
// 如果结构在该版本不存在，则返回 error。
func (f *Finder) GetStructureConfig(structureType int) (StructureConfig, error) {
	var sc C.StructureConfig
	ok := C.getStructureConfig(C.int(structureType), C.int(f.Version), &sc)
	if ok == 0 {
		return StructureConfig{}, errors.New("getStructureConfig failed (unsupported version or structure)")
	}
	return structureConfigFromC(sc), nil
}

// ChunkGenerateRnd 对应 cubiomes 的 [`chunkGenerateRnd()`](GO/cubiomes/finders.h:389)。
// 返回 chunk 生成用的 48-bit RNG seed（存放在 uint64）。
func (f *Finder) ChunkGenerateRnd(worldSeed uint64, chunkX, chunkZ int) uint64 {
	// 注意：chunkGenerateRnd 在 finders.h 里是 inline static，cgo 不能直接调。
	return uint64(C.go_chunkGenerateRnd(C.ulonglong(worldSeed), C.int(chunkX), C.int(chunkZ)))
}

// GetStructurePos 返回结构在指定 region 内的生成尝试位置（block 坐标）。
// 若该 region 无有效位置，返回 (nil, nil)。
//
// 参数 seed：仅低 48 位影响结果，但这里接受完整 64 位。
func (f *Finder) GetStructurePos(structureType int, seed uint64, regX, regZ int) (*Pos, error) {
	var p C.Pos
	ok := C.getStructurePos(C.int(structureType), C.int(f.Version), C.ulonglong(seed), C.int(regX), C.int(regZ), &p)
	if ok == 0 {
		return nil, nil
	}
	pp := posFromC(p)
	return &pp, nil
}
