//go:build cgo

package gobiomes

/*
#include <stdint.h>

#include "helpers.h"
#include "cubiomes/rng.h"
*/
import "C"

// Rng 是对 cubiomes Java Random 实现的 Go 封装。
//
// 对应原项目 Python 用法：[`src/objects/rng.c`](src/objects/rng.c:1)
//   - set_seed
//   - next(bits)
//   - next_long
//   - next_int(n)
//   - next_float
//   - next_double
//
// 注意：该 RNG 与 Minecraft/Java Random 一致（48-bit LCG）。
type Rng struct {
	seed C.ulonglong
}

// NewRng 创建 RNG，并将内部状态初始化为 Java Random 的 seed 格式。
func NewRng(seed uint64) *Rng {
	r := &Rng{}
	r.SetSeed(seed)
	return r
}

// SetSeed 对应 setSeed(&seed, value)。
func (r *Rng) SetSeed(value uint64) {
	C.go_setSeed((*C.uint64_t)(&r.seed), C.ulonglong(value))
}

// Next 对应 next(&seed, bits)。
func (r *Rng) Next(bits int) int {
	return int(C.go_next((*C.uint64_t)(&r.seed), C.int(bits)))
}

// NextLong 对应 nextLong(&seed)。
func (r *Rng) NextLong() uint64 {
	return uint64(C.go_nextLong((*C.uint64_t)(&r.seed)))
}

// NextInt 对应 nextInt(&seed, n)。
func (r *Rng) NextInt(n int) int {
	return int(C.go_nextInt((*C.uint64_t)(&r.seed), C.int(n)))
}

// NextFloat 对应 nextFloat(&seed)。
func (r *Rng) NextFloat() float32 {
	return float32(C.go_nextFloat((*C.uint64_t)(&r.seed)))
}

// NextDouble 对应 nextDouble(&seed)。
func (r *Rng) NextDouble() float64 {
	return float64(C.go_nextDouble((*C.uint64_t)(&r.seed)))
}
