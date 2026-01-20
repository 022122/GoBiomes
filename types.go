package gobiomes

// Pos 对应 cubiomes 的 Pos（block 坐标，x/z）。
type Pos struct {
	X int
	Z int
}

// Pos3 对应 cubiomes 的 Pos3（3D block 坐标）。
type Pos3 struct {
	X, Y, Z int
}

// Range 对应 cubiomes 的 Range。
type Range struct {
	Scale int
	X     int
	Z     int
	SX    int
	SZ    int
	Y     int
	SY    int
}

// NewRange2D 创建一个 2D Range。
func NewRange2D(scale, x, z, sx, sz int) Range {
	return Range{Scale: scale, X: x, Z: z, SX: sx, SZ: sz, SY: 1}
}

// NewRange3D 创建一个 3D Range。
func NewRange3D(scale, x, z, sx, sz, y, sy int) Range {
	return Range{Scale: scale, X: x, Z: z, SX: sx, SZ: sz, Y: y, SY: sy}
}
