//go:build cgo

package gobiomes

/*
#include "cubiomes/biomenoise.h"
#include "cubiomes/finders.h"
*/
import "C"

import "unsafe"

// Pos 对应 cubiomes 的 Pos（block 坐标，x/z）。
// 参考：[`GO/cubiomes/finders.h:58`](GO/cubiomes/finders.h:58)
type Pos struct {
	X int
	Z int
}

func posFromC(p C.Pos) Pos {
	return Pos{X: int(p.x), Z: int(p.z)}
}

// Range 对应 cubiomes 的 Range。
// 注意：字段顺序/含义需要与 C 结构体一致。
// 参考：[`GO/cubiomes/biomenoise.h:8`](GO/cubiomes/biomenoise.h:8)
type Range struct {
	Scale int
	X     int
	Z     int
	SX    int
	SZ    int
	Y     int
	SY    int
}

// NewRange2D 创建一个 2D Range（sy 会按 cubiomes 规则当作 1）。
func NewRange2D(scale, x, z, sx, sz int) Range {
	return Range{Scale: scale, X: x, Z: z, SX: sx, SZ: sz}
}

// NewRange3D 创建一个 3D Range。
func NewRange3D(scale, x, z, sx, sz, y, sy int) Range {
	return Range{Scale: scale, X: x, Z: z, SX: sx, SZ: sz, Y: y, SY: sy}
}

func (r Range) toC() C.Range {
	return C.Range{
		scale: C.int(r.Scale),
		x:     C.int(r.X),
		z:     C.int(r.Z),
		sx:    C.int(r.SX),
		sz:    C.int(r.SZ),
		y:     C.int(r.Y),
		sy:    C.int(r.SY),
	}
}

// intsFromC 从 C int* + 长度构造 Go []int（不拷贝）。
// 调用方必须保证 ptr 在 slice 生命周期内有效。
func intsFromC(ptr *C.int, n int) []int {
	if ptr == nil || n <= 0 {
		return nil
	}
	return unsafe.Slice((*int)(unsafe.Pointer(ptr)), n)
}
