package main

import (
	"container/heap"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"gobiomes"
)

// monument_pairs
// - 搜索范围：默认 [-200000,200000] x [-200000,200000]（40w*40w）
//   或使用 -radius=1000000 覆盖为 [-1000000,+1000000]（100w 半径）
// - 结构：海底神殿（Monument / “海货塔”）
// - 找“二联”：两处神殿中心点 XZ 距离 <= maxd
// - 几何条件（3D）：存在点 P(x,y,z)，满足：
//   1) 以 P 为中心、半径 outerR=128 的球体可以包含两个神殿的全部“刷怪体积”
//   2) 以 P 为中心、半径 innerR=24 的球体内不包含两个神殿的任何刷怪体积
//
// 这里对刷怪体积做了简化建模：
// - X/Z：58×58（以结构生成尝试点为中心，半径 29）
// - Y：39..61（绝对高度，默认守卫者刷怪高度区间）
//
// 注：如果你需要与游戏内“精确边界/偏移”100%一致，需要进一步引入结构的 bounding box/variant。

type pos2 struct {
	X int `json:"x"`
	Z int `json:"z"`
}

type pos3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type pair struct {
	A      pos2    `json:"a"`
	B      pos2    `json:"b"`
	Dist   float64 `json:"dist"`
	Weight float64 `json:"weight"`
	Center *pos3   `json:"center,omitempty"`
}

type outJSON struct {
	Seed        uint64  `json:"seed"`
	MC          int     `json:"mc"`
	Area        [4]int  `json:"area"` // minx,maxx,minz,maxz
	MaxDist     float64 `json:"maxDist"`
	OuterR      float64 `json:"outerR"`
	InnerR      float64 `json:"innerR"`
	BoxHalfXZ   float64 `json:"boxHalfXZ"`
	YMin        float64 `json:"yMin"`
	YMax        float64 `json:"yMax"`
	Monuments   int     `json:"monuments"`
	PairsTotal  int64   `json:"pairsTotal"`
	TopK        int     `json:"topK"`
	TopPairs    []pair  `json:"topPairs"`
	GeneratedAt string  `json:"generatedAt"`
}

func main() {
	var (
		seed    = flag.Uint64("seed", 20251223, "世界种子(64-bit)")
		mc      = flag.Int("mc", gobiomes.MC_1_21_1, "版本常量")
		radius  = flag.Int("radius", 0, "若>0：覆盖搜索范围为 [-radius,+radius]（例如 1000000 表示 100w 半径）")
		minX    = flag.Int("minx", -200000, "最小 X(block)")
		maxX    = flag.Int("maxx", 200000, "最大 X(block)")
		minZ    = flag.Int("minz", -200000, "最小 Z(block)")
		maxZ    = flag.Int("maxz", 200000, "最大 Z(block)")
		maxD    = flag.Float64("maxd", 220, "二联最大 XZ 距离(方块)，默认 220")
		step    = flag.Float64("step", 4, "搜索中心点 P 时的网格步长（XYZ 同步），越小越严格但越慢")
		format  = flag.String("format", "json", "输出格式: json 或 md")
		outPath = flag.String("out", "monument_pairs.json", "输出文件路径")
		topK    = flag.Int("top", 2000, "只输出 TopK 对（按权重排序）")
		workers = flag.Int("workers", runtime.NumCPU(), "并发 worker 数（扫描+配对）")
		quiet   = flag.Bool("quiet", false, "安静模式（不打印进度条/提示）")
	)
	flag.Parse()

	if *workers <= 0 {
		*workers = runtime.NumCPU()
	}
	if *workers < 1 {
		*workers = 1
	}

	if *radius > 0 {
		*minX, *maxX = -*radius, *radius
		*minZ, *maxZ = -*radius, *radius
	}
	if *minX > *maxX {
		*minX, *maxX = *maxX, *minX
	}
	if *minZ > *maxZ {
		*minZ, *maxZ = *maxZ, *minZ
	}
	if *maxD <= 0 {
		panic("maxd must be > 0")
	}
	if *step <= 0 {
		*step = 4
	}
	if *topK < 1 {
		*topK = 1
	}

	finder := gobiomes.NewFinder(*mc)
	gen := gobiomes.NewGenerator(*mc, 0)
	gen.ApplySeed(*seed, gobiomes.DimOverworld)

	sc, err := finder.GetStructureConfig(gobiomes.Monument)
	if err != nil {
		panic(err)
	}
	regionBlocks := int(sc.RegionSize) * 16
	if regionBlocks <= 0 {
		panic("invalid region size")
	}

	rx0 := floorDiv(*minX, regionBlocks)
	rx1 := floorDiv(*maxX, regionBlocks)
	rz0 := floorDiv(*minZ, regionBlocks)
	rz1 := floorDiv(*maxZ, regionBlocks)

	// 1) 扫描区域内所有可行的海底神殿位置（并行）
	type regionTask struct{ rx, rz int }
	tasks := make(chan regionTask, 4096)
	found := make(chan pos2, 4096)

	positions := make([]pos2, 0, 1024)
	collectDone := make(chan struct{})
	go func() {
		for p := range found {
			positions = append(positions, p)
		}
		close(collectDone)
	}()

	regionsTotal := int64(rx1-rx0+1) * int64(rz1-rz0+1)
	var regionsDone int64

	// 进度条
	var progStop chan struct{}
	if !*quiet {
		progStop = make(chan struct{})
		go func() {
			t := time.NewTicker(200 * time.Millisecond)
			defer t.Stop()
			for {
				select {
				case <-t.C:
					printProgress("scan regions", atomic.LoadInt64(&regionsDone), regionsTotal)
				case <-progStop:
					printProgress("scan regions", regionsTotal, regionsTotal)
					fmt.Println()
					return
				}
			}
		}()
	}

	var wg sync.WaitGroup
	for w := 0; w < *workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wFinder := gobiomes.NewFinder(*mc)
			wGen := gobiomes.NewGenerator(*mc, 0)
			wGen.ApplySeed(*seed, gobiomes.DimOverworld)

			for tk := range tasks {
				p, err := wFinder.GetStructurePos(gobiomes.Monument, *seed, tk.rx, tk.rz)
				if err != nil {
					panic(err)
				}
				if p != nil {
					if p.X >= *minX && p.X <= *maxX && p.Z >= *minZ && p.Z <= *maxZ {
						if wGen.IsViableStructurePos(gobiomes.Monument, p.X, p.Z, 0) {
							found <- pos2{X: p.X, Z: p.Z}
						}
					}
				}
				atomic.AddInt64(&regionsDone, 1)
			}
		}()
	}

	for rz := rz0; rz <= rz1; rz++ {
		for rx := rx0; rx <= rx1; rx++ {
			tasks <- regionTask{rx: rx, rz: rz}
		}
	}
	close(tasks)
	wg.Wait()
	close(found)
	<-collectDone
	if !*quiet {
		close(progStop)
	}

	if len(positions) == 0 {
		fmt.Println("no Monuments found in area")
		return
	}

	// 2) 二联查找：空间哈希 + 并行 + TopK
	cellSize := int(math.Ceil(*maxD))
	if cellSize < 1 {
		cellSize = 1
	}
	cells := make(map[[2]int][]int, len(positions))
	for i, p := range positions {
		cx := floorDiv(p.X, cellSize)
		cz := floorDiv(p.Z, cellSize)
		cells[[2]int{cx, cz}] = append(cells[[2]int{cx, cz}], i)
	}

	// 配对任务：按 i 分块
	type pairTask struct{ i0, i1 int }
	pairTasks := make(chan pairTask, 1024)
	pairRes := make(chan []pair, *workers)

	var checked int64
	var pairsTotal int64

	// calc 进度
	if !*quiet {
		stop := make(chan struct{})
		go func() {
			t := time.NewTicker(300 * time.Millisecond)
			defer t.Stop()
			for {
				select {
				case <-t.C:
					fmt.Printf("\r[calc pairs] checked=%d pairsTotal=%d", atomic.LoadInt64(&checked), atomic.LoadInt64(&pairsTotal))
				case <-stop:
					fmt.Printf("\r[calc pairs] checked=%d pairsTotal=%d\n", atomic.LoadInt64(&checked), atomic.LoadInt64(&pairsTotal))
					return
				}
			}
		}()
		defer close(stop)
	}

	outerR := 128.0
	innerR := 24.0
	boxHalfXZ := 29.0
	yMin := 39.0
	yMax := 61.0

	var wgPairs sync.WaitGroup
	for w := 0; w < *workers; w++ {
		wgPairs.Add(1)
		go func() {
			defer wgPairs.Done()
			local := make(pairHeap, 0, *topK)

			for tk := range pairTasks {
				for i := tk.i0; i < tk.i1; i++ {
					a := positions[i]
					cx := floorDiv(a.X, cellSize)
					cz := floorDiv(a.Z, cellSize)
					for dz := -1; dz <= 1; dz++ {
						for dx := -1; dx <= 1; dx++ {
							lst := cells[[2]int{cx + dx, cz + dz}]
							for _, j := range lst {
								if j <= i {
									continue
								}
								b := positions[j]
								d := dist2D(a, b)
								if d <= *maxD {
									// 3D 可行性判定（找点 P）
									ok, c := existsCenter3D(a, b, outerR, innerR, boxHalfXZ, yMin, yMax, *step)
									if ok {
										atomic.AddInt64(&pairsTotal, 1)
										wgt := (*maxD - d)
										pushTopK(&local, pair{A: a, B: b, Dist: d, Weight: wgt, Center: &c}, *topK)
									}
								}
								atomic.AddInt64(&checked, 1)
							}
						}
					}
				}
			}

			out := make([]pair, 0, len(local))
			for local.Len() > 0 {
				out = append(out, heap.Pop(&local).(pair))
			}
			pairRes <- out
		}()
	}

	chunk := 4096
	for i := 0; i < len(positions); i += chunk {
		j := i + chunk
		if j > len(positions) {
			j = len(positions)
		}
		pairTasks <- pairTask{i0: i, i1: j}
	}
	close(pairTasks)
	wgPairs.Wait()
	close(pairRes)

	pq := make(pairHeap, 0, *topK)
	for lst := range pairRes {
		for _, p := range lst {
			pushTopK(&pq, p, *topK)
		}
	}

	topPairs := make([]pair, 0, len(pq))
	for pq.Len() > 0 {
		topPairs = append(topPairs, heap.Pop(&pq).(pair))
	}
	reversePairs(topPairs)
	// 稳定排序：权重降序，距离升序
	sort.SliceStable(topPairs, func(i, j int) bool {
		if topPairs[i].Weight != topPairs[j].Weight {
			return topPairs[i].Weight > topPairs[j].Weight
		}
		return topPairs[i].Dist < topPairs[j].Dist
	})

	o := outJSON{
		Seed:        *seed,
		MC:          *mc,
		Area:        [4]int{*minX, *maxX, *minZ, *maxZ},
		MaxDist:     *maxD,
		OuterR:      outerR,
		InnerR:      innerR,
		BoxHalfXZ:   boxHalfXZ,
		YMin:        yMin,
		YMax:        yMax,
		Monuments:   len(positions),
		PairsTotal:  pairsTotal,
		TopK:        *topK,
		TopPairs:    topPairs,
		GeneratedAt: time.Now().Format(time.RFC3339),
	}

	if err := writeOut(*format, *outPath, o); err != nil {
		panic(err)
	}

	fmt.Printf("seed=%d mc=%d area=[%d,%d]x[%d,%d] monuments=%d pairs(found)=%d top=%d\n",
		o.Seed, o.MC, o.Area[0], o.Area[1], o.Area[2], o.Area[3], o.Monuments, o.PairsTotal, len(o.TopPairs))
	fmt.Printf("已输出: %s\n", *outPath)
}

// --------------------
// Geometry
// --------------------

type aabb3 struct {
	minX, maxX float64
	minY, maxY float64
	minZ, maxZ float64
}

type p3 struct{ x, y, z float64 }

func monumentBox(center pos2, halfXZ, yMin, yMax float64) aabb3 {
	cx := float64(center.X)
	cz := float64(center.Z)
	return aabb3{
		minX: cx - halfXZ,
		maxX: cx + halfXZ,
		minZ: cz - halfXZ,
		maxZ: cz + halfXZ,
		minY: yMin,
		maxY: yMax,
	}
}

func cornersOfBox(b aabb3) [8]p3 {
	x0, x1 := b.minX, b.maxX
	y0, y1 := b.minY, b.maxY
	z0, z1 := b.minZ, b.maxZ
	return [8]p3{
		{x0, y0, z0}, {x0, y0, z1}, {x0, y1, z0}, {x0, y1, z1},
		{x1, y0, z0}, {x1, y0, z1}, {x1, y1, z0}, {x1, y1, z1},
	}
}

func existsCenter3D(a, b pos2, outerR, innerR, halfXZ, yMin, yMax, step float64) (bool, pos3) {
	boxA := monumentBox(a, halfXZ, yMin, yMax)
	boxB := monumentBox(b, halfXZ, yMin, yMax)
	ca := cornersOfBox(boxA)
	cb := cornersOfBox(boxB)

	// 对每个 corner：P 必须落在以 corner 为球心、半径 outerR 的球内。
	// 用球的 AABB 包围盒做快速裁剪：
	//   x in [cx-R, cx+R]，y in [cy-R,cy+R]，z in [cz-R,cz+R]
	xmin, xmax := math.Inf(-1), math.Inf(+1)
	ymin, ymax := math.Inf(-1), math.Inf(+1)
	zmin, zmax := math.Inf(-1), math.Inf(+1)

	for _, c := range ca {
		xmin = math.Max(xmin, c.x-outerR)
		xmax = math.Min(xmax, c.x+outerR)
		ymin = math.Max(ymin, c.y-outerR)
		ymax = math.Min(ymax, c.y+outerR)
		zmin = math.Max(zmin, c.z-outerR)
		zmax = math.Min(zmax, c.z+outerR)
	}
	for _, c := range cb {
		xmin = math.Max(xmin, c.x-outerR)
		xmax = math.Min(xmax, c.x+outerR)
		ymin = math.Max(ymin, c.y-outerR)
		ymax = math.Min(ymax, c.y+outerR)
		zmin = math.Max(zmin, c.z-outerR)
		zmax = math.Min(zmax, c.z+outerR)
	}
	if xmin > xmax || ymin > ymax || zmin > zmax {
		return false, pos3{}
	}
	if step <= 0 {
		step = 4
	}

	outerR2 := outerR * outerR
	innerR2 := innerR * innerR

	withinAllCorners := func(p p3) bool {
		for _, c := range ca {
			dx := p.x - c.x
			dy := p.y - c.y
			dz := p.z - c.z
			if dx*dx+dy*dy+dz*dz > outerR2 {
				return false
			}
		}
		for _, c := range cb {
			dx := p.x - c.x
			dy := p.y - c.y
			dz := p.z - c.z
			if dx*dx+dy*dy+dz*dz > outerR2 {
				return false
			}
		}
		return true
	}

	okInner := func(p p3) bool {
		// inner sphere must NOT intersect volumes: dist(point, AABB) >= innerR
		if pointAABBDist2(p, boxA) < innerR2 {
			return false
		}
		if pointAABBDist2(p, boxB) < innerR2 {
			return false
		}
		return true
	}

	// 网格扫描
	for x := xmin; x <= xmax; x += step {
		for z := zmin; z <= zmax; z += step {
			for y := ymin; y <= ymax; y += step {
				p := p3{x: x, y: y, z: z}
				if !withinAllCorners(p) {
					continue
				}
				if !okInner(p) {
					continue
				}
				return true, pos3{X: x, Y: y, Z: z}
			}
		}
	}

	// 退化尝试：用盒中心
	p := p3{x: (xmin + xmax) / 2, y: (ymin + ymax) / 2, z: (zmin + zmax) / 2}
	if withinAllCorners(p) && okInner(p) {
		return true, pos3{X: p.x, Y: p.y, Z: p.z}
	}

	return false, pos3{}
}

func pointAABBDist2(p p3, b aabb3) float64 {
	dx := dist1D(p.x, b.minX, b.maxX)
	dy := dist1D(p.y, b.minY, b.maxY)
	dz := dist1D(p.z, b.minZ, b.maxZ)
	return dx*dx + dy*dy + dz*dz
}

func dist1D(v, lo, hi float64) float64 {
	if v < lo {
		return lo - v
	}
	if v > hi {
		return v - hi
	}
	return 0
}

// --------------------
// Output
// --------------------

func writeOut(format, path string, o outJSON) error {
	switch format {
	case "json":
		b, err := json.MarshalIndent(o, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(path, b, 0o644)
	case "md", "markdown":
		return os.WriteFile(path, []byte(renderMD(o)), 0o644)
	default:
		return fmt.Errorf("unknown format: %s (use json|md)", format)
	}
}

func renderMD(o outJSON) string {
	s := "# 海底神殿二联（Monument Pairs）\n\n"
	s += fmt.Sprintf("- seed: `%d`\n", o.Seed)
	s += fmt.Sprintf("- mc: `%d`\n", o.MC)
	s += fmt.Sprintf("- area: `[%d,%d]×[%d,%d]`\n", o.Area[0], o.Area[1], o.Area[2], o.Area[3])
	s += fmt.Sprintf("- maxDist(xz): `%.0f`\n", o.MaxDist)
	s += fmt.Sprintf("- outerR: `%.0f`\n", o.OuterR)
	s += fmt.Sprintf("- innerR: `%.0f`\n", o.InnerR)
	s += fmt.Sprintf("- boxHalfXZ: `%.0f` (58×58)\n", o.BoxHalfXZ)
	s += fmt.Sprintf("- yRange: `[%.0f,%.0f]`\n", o.YMin, o.YMax)
	s += fmt.Sprintf("- monuments: `%d`\n", o.Monuments)
	s += fmt.Sprintf("- pairs(found): `%d`\n", o.PairsTotal)
	s += fmt.Sprintf("- topK: `%d`\n", o.TopK)
	s += fmt.Sprintf("- generatedAt: `%s`\n\n", o.GeneratedAt)

	s += "## Top Pairs\n\n"
	s += "| # | A(x,z) | B(x,z) | dist | weight | center(x,y,z) |\n"
	s += "|---:|---|---|---:|---:|---|\n"
	for i, p := range o.TopPairs {
		c := ""
		if p.Center != nil {
			c = fmt.Sprintf("%.1f,%.1f,%.1f", p.Center.X, p.Center.Y, p.Center.Z)
		}
		s += fmt.Sprintf("| %d | %d,%d | %d,%d | %.2f | %.2f | %s |\n",
			i+1, p.A.X, p.A.Z, p.B.X, p.B.Z, p.Dist, p.Weight, c)
	}
	return s
}

// --------------------
// Utils
// --------------------

func dist2D(a, b pos2) float64 {
	dx := float64(a.X - b.X)
	dz := float64(a.Z - b.Z)
	return math.Hypot(dx, dz)
}

func floorDiv(a, b int) int {
	q := a / b
	r := a % b
	if r != 0 && ((r < 0) != (b < 0)) {
		q--
	}
	return q
}

func printProgress(stage string, done, total int64) {
	if total <= 0 {
		return
	}
	pct := float64(done) / float64(total)
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	barW := 30
	filled := int(math.Round(pct * float64(barW)))
	if filled < 0 {
		filled = 0
	}
	if filled > barW {
		filled = barW
	}
	bar := make([]byte, 0, barW)
	for i := 0; i < barW; i++ {
		if i < filled {
			bar = append(bar, '=')
		} else {
			bar = append(bar, '-')
		}
	}
	fmt.Printf("\r[%s] %s %6.2f%% (%d/%d)", string(bar), stage, pct*100, done, total)
}

// --------------------
// TopK heap (min-heap)
// --------------------

type pairHeap []pair

func (h pairHeap) Len() int { return len(h) }

// Less implements min-heap: the "worst" (lowest weight, then higher dist) is on top.
func (h pairHeap) Less(i, j int) bool {
	if h[i].Weight != h[j].Weight {
		return h[i].Weight < h[j].Weight
	}
	return h[i].Dist > h[j].Dist
}

func (h pairHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *pairHeap) Push(x any) {
	*h = append(*h, x.(pair))
}

func (h *pairHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func pushTopK(h *pairHeap, p pair, k int) {
	if k <= 0 {
		return
	}
	if h.Len() < k {
		heap.Push(h, p)
		return
	}
	worst := (*h)[0]
	if p.Weight > worst.Weight || (p.Weight == worst.Weight && p.Dist < worst.Dist) {
		(*h)[0] = p
		heap.Fix(h, 0)
	}
}

func reversePairs(a []pair) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}
