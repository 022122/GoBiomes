package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"gobiomes"
)

// 说明
// - 搜索范围：默认 [-200000,200000] x [-200000,200000]（即 40w*40w）
// - 结构：女巫小屋（SwampHut）
// - 输出：双联/三联/四联
// - 四联额外条件：存在点 P
//     1) 以 P 为圆心、半径 128 的圆可以包含 4 个小屋“刷怪范围”
//     2) 以 P 为圆心、半径 24 的圆不与 4 个小屋“刷怪范围”相交
//
// 注意：这里将“刷怪范围”近似为以女巫小屋中心点为圆心、半径 spawnR 的圆。
// spawnR 默认取 8（略保守）。如果你希望更严格/更宽松，可用 -spawnR 调整。

type hut struct {
	X int
	Z int
}

type pt struct {
	X float64
	Z float64
}

type result struct {
	size   int
	idx    []int
	center *pt // 仅四联时非 nil
}

func main() {
	var (
		seed   = flag.Uint64("seed", 0, "世界种子(64-bit)")
		mc     = flag.Int("mc", gobiomes.MC_1_21_1, "版本常量")
		minX   = flag.Int("minx", -200000, "最小 X(block)")
		maxX   = flag.Int("maxx", 200000, "最大 X(block)")
		minZ   = flag.Int("minz", -200000, "最小 Z(block)")
		maxZ   = flag.Int("maxz", 200000, "最大 Z(block)")
		spawnR = flag.Float64("spawnR", 8, "刷怪范围半径近似(方块)，默认 8")
		step   = flag.Float64("step", 2, "寻找中心点时的网格步长(方块)，越小越严格但越慢")
		out    = flag.String("out", "witch_hut_clusters.json", "输出 JSON 路径")
		quiet  = flag.Bool("quiet", false, "安静模式（不打印进度条）")
	)
	flag.Parse()

	if *minX > *maxX {
		*minX, *maxX = *maxX, *minX
	}
	if *minZ > *maxZ {
		*minZ, *maxZ = *maxZ, *minZ
	}

	finder := gobiomes.NewFinder(*mc)
	gen := gobiomes.NewGenerator(*mc, 0)
	gen.ApplySeed(*seed, gobiomes.DimOverworld)

	sc, err := finder.GetStructureConfig(gobiomes.SwampHut)
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

	// 1) 收集范围内所有可生成（biome viable）的女巫小屋位置
	huts := make([]hut, 0, 1024)
	regionsTotal := int64(rx1-rx0+1) * int64(rz1-rz0+1)
	var regionsDone int64
	lastPrint := time.Now()

	for rz := rz0; rz <= rz1; rz++ {
		for rx := rx0; rx <= rx1; rx++ {
			p, err := finder.GetStructurePos(gobiomes.SwampHut, *seed, rx, rz)
			if err != nil {
				panic(err)
			}
			if p != nil {
				if p.X >= *minX && p.X <= *maxX && p.Z >= *minZ && p.Z <= *maxZ {
					if gen.IsViableStructurePos(gobiomes.SwampHut, p.X, p.Z, 0) {
						huts = append(huts, hut{X: p.X, Z: p.Z})
					}
				}
			}

			regionsDone++
			if !*quiet && time.Since(lastPrint) > 200*time.Millisecond {
				printProgress("scan regions", regionsDone, regionsTotal)
				lastPrint = time.Now()
			}
		}
	}
	if !*quiet {
		printProgress("scan regions", regionsTotal, regionsTotal)
		fmt.Println()
	}

	if len(huts) == 0 {
		fmt.Println("no huts found")
		return
	}

	// 2) 构建“可能同圈”邻接：若两小屋中心距离 <= 2*(128-spawnR) 则存在半径 128 圆可覆盖二者刷怪范围
	outerR := 128.0
	innerR := 24.0
	bigR := outerR - *spawnR
	if bigR <= 0 {
		panic("spawnR too large")
	}
	neighborDist := 2 * bigR
	cellSize := int(math.Ceil(neighborDist))
	if cellSize < 1 {
		cellSize = 1
	}

	cells := make(map[[2]int][]int, len(huts))
	for i, h := range huts {
		cx := floorDiv(h.X, cellSize)
		cz := floorDiv(h.Z, cellSize)
		cells[[2]int{cx, cz}] = append(cells[[2]int{cx, cz}], i)
	}

	uf := newUF(len(huts))
	for i, h := range huts {
		cx := floorDiv(h.X, cellSize)
		cz := floorDiv(h.Z, cellSize)
		for dz := -1; dz <= 1; dz++ {
			for dx := -1; dx <= 1; dx++ {
				lst := cells[[2]int{cx + dx, cz + dz}]
				for _, j := range lst {
					if j <= i {
						continue
					}
					if dist2(huts[i], huts[j]) <= neighborDist*neighborDist {
						uf.union(i, j)
					}
				}
			}
		}
	}

	// 3) 分组
	groups := map[int][]int{}
	for i := range huts {
		root := uf.find(i)
		groups[root] = append(groups[root], i)
	}

	// 4) 输出：对每个 connected component：
	//    - size 2/3/4 直接考虑
	//    - size >4：枚举子集（优先 4 -> 3 -> 2）
	res := make([]result, 0, 256)

	for _, idxs := range groups {
		sort.Ints(idxs)
		sz := len(idxs)
		switch {
		case sz < 2:
			continue
		case sz <= 4:
			appendBySize(&res, huts, idxs, outerR, innerR, *spawnR, *step)
		default:
			// 超过 4 的簇：枚举子集
			appendCombos(&res, huts, idxs, 4, outerR, innerR, *spawnR, *step)
			appendCombos(&res, huts, idxs, 3, outerR, innerR, *spawnR, *step)
			appendCombos(&res, huts, idxs, 2, outerR, innerR, *spawnR, *step)
		}
	}

	// 去重（因为大簇枚举会重复）
	uniq := map[string]result{}
	for _, r := range res {
		key := fmt.Sprintf("%d:%v", r.size, r.idx)
		if _, ok := uniq[key]; ok {
			continue
		}
		uniq[key] = r
	}

	res = res[:0]
	for _, r := range uniq {
		res = append(res, r)
	}

	sort.Slice(res, func(i, j int) bool {
		if res[i].size != res[j].size {
			return res[i].size > res[j].size
		}
		// 稳定输出
		ai, aj := res[i].idx, res[j].idx
		for k := 0; k < len(ai) && k < len(aj); k++ {
			if ai[k] != aj[k] {
				return ai[k] < aj[k]
			}
		}
		return len(ai) < len(aj)
	})

	// 输出 JSON
	type jsonHut struct {
		X int `json:"x"`
		Z int `json:"z"`
	}
	type jsonCenter struct {
		X float64 `json:"x"`
		Z float64 `json:"z"`
	}
	type jsonCluster struct {
		Size   int         `json:"size"`
		Huts   []jsonHut   `json:"huts"`
		Center *jsonCenter `json:"center,omitempty"`
	}
	type jsonOut struct {
		Seed     uint64        `json:"seed"`
		MC       int           `json:"mc"`
		Area     [4]int        `json:"area"` // minx,maxx,minz,maxz
		SpawnR   float64       `json:"spawnR"`
		OuterR   float64       `json:"outerR"`
		InnerR   float64       `json:"innerR"`
		HutCount int           `json:"hutCount"`
		Clusters []jsonCluster `json:"clusters"`
	}

	outObj := jsonOut{
		Seed:     *seed,
		MC:       *mc,
		Area:     [4]int{*minX, *maxX, *minZ, *maxZ},
		SpawnR:   *spawnR,
		OuterR:   outerR,
		InnerR:   innerR,
		HutCount: len(huts),
		Clusters: make([]jsonCluster, 0, len(res)),
	}

	for _, r := range res {
		c := jsonCluster{Size: r.size, Huts: make([]jsonHut, 0, len(r.idx))}
		if r.size == 4 && r.center != nil {
			c.Center = &jsonCenter{X: r.center.X, Z: r.center.Z}
		}
		for _, id := range r.idx {
			h := huts[id]
			c.Huts = append(c.Huts, jsonHut{X: h.X, Z: h.Z})
		}
		outObj.Clusters = append(outObj.Clusters, c)
	}

	data, err := json.MarshalIndent(outObj, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(*out, data, 0o644); err != nil {
		panic(err)
	}

	fmt.Printf("seed=%d mc=%d area=[%d,%d]x[%d,%d] huts=%d clusters=%d\n", *seed, *mc, *minX, *maxX, *minZ, *maxZ, len(huts), len(outObj.Clusters))
	fmt.Printf("已输出 JSON: %s\n", *out)
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
	// \r 覆盖当前行
	fmt.Printf("\r[%s] %s %6.2f%% (%d/%d)", string(bar), stage, pct*100, done, total)
}

func appendBySize(out *[]result, huts []hut, idxs []int, outerR, innerR, spawnR, step float64) {
	sz := len(idxs)
	switch sz {
	case 2, 3:
		// 对双联/三联：只要求存在中心点能覆盖刷怪范围(outerR)
		if ok, _ := existsCenter(idxs, huts, outerR, -1, spawnR, step); ok {
			*out = append(*out, result{size: sz, idx: append([]int(nil), idxs...)})
		}
	case 4:
		// 对四联：同时满足 outerR 覆盖 + innerR 不相交
		if ok, c := existsCenter(idxs, huts, outerR, innerR, spawnR, step); ok {
			*out = append(*out, result{size: 4, idx: append([]int(nil), idxs...), center: &c})
		}
	}
}

func appendCombos(out *[]result, huts []hut, idxs []int, k int, outerR, innerR, spawnR, step float64) {
	if k < 2 || k > len(idxs) {
		return
	}

	// 简单组合枚举（idxs 通常很小）
	comb := make([]int, 0, k)
	var dfs func(start int)
	seen := map[string]struct{}{}

	dfs = func(start int) {
		if len(comb) == k {
			key := fmt.Sprint(comb)
			if _, ok := seen[key]; ok {
				return
			}
			seen[key] = struct{}{}

			// k==4 用双条件，否则只用覆盖条件
			inR := -1.0
			if k == 4 {
				inR = innerR
			}
			if ok, c := existsCenter(comb, huts, outerR, inR, spawnR, step); ok {
				var cp *pt
				if k == 4 {
					cp = &c
				}
				*out = append(*out, result{size: k, idx: append([]int(nil), comb...), center: cp})
			}
			return
		}
		for i := start; i < len(idxs); i++ {
			comb = append(comb, idxs[i])
			dfs(i + 1)
			comb = comb[:len(comb)-1]
		}
	}

	dfs(0)
}

// existsCenter 判断是否存在点 P 满足：
//   - 对每个 hut：dist(P, hutCenter) <= outerR - spawnR （使得半径 outerR 的圆覆盖 hut 刷怪范围）
//   - 若 innerR >= 0：对每个 hut：dist(P, hutCenter) >= innerR + spawnR （使得半径 innerR 的圆不与刷怪范围相交）
//
// 返回找到的一个中心点（近似）。
func existsCenter(idxs []int, huts []hut, outerR, innerR, spawnR, step float64) (bool, pt) {
	bigR := outerR - spawnR
	smallR := innerR + spawnR

	// 先求所有 big circle 的 bbox 交集
	xmin := math.Inf(-1)
	xmax := math.Inf(+1)
	zmin := math.Inf(-1)
	zmax := math.Inf(+1)

	for _, id := range idxs {
		h := huts[id]
		cx, cz := float64(h.X), float64(h.Z)
		xmin = math.Max(xmin, cx-bigR)
		xmax = math.Min(xmax, cx+bigR)
		zmin = math.Max(zmin, cz-bigR)
		zmax = math.Min(zmax, cz+bigR)
	}

	if xmin > xmax || zmin > zmax {
		return false, pt{}
	}
	if step <= 0 {
		step = 2
	}

	// 网格扫描：bbox 尺寸 <= 2*bigR，bigR<=128，因此最多 ~256/step 的粒度
	for x := xmin; x <= xmax; x += step {
		for z := zmin; z <= zmax; z += step {
			p := pt{X: x, Z: z}
			if !withinAll(p, idxs, huts, bigR) {
				continue
			}
			if innerR >= 0 {
				if intersectsAny(p, idxs, huts, smallR) {
					continue
				}
			}
			return true, p
		}
	}

	// 最后尝试 bbox 中心点
	p := pt{X: (xmin + xmax) / 2, Z: (zmin + zmax) / 2}
	if withinAll(p, idxs, huts, bigR) {
		if innerR < 0 || !intersectsAny(p, idxs, huts, smallR) {
			return true, p
		}
	}
	return false, pt{}
}

func withinAll(p pt, idxs []int, huts []hut, r float64) bool {
	r2 := r * r
	for _, id := range idxs {
		h := huts[id]
		dx := p.X - float64(h.X)
		dz := p.Z - float64(h.Z)
		if dx*dx+dz*dz > r2 {
			return false
		}
	}
	return true
}

func intersectsAny(p pt, idxs []int, huts []hut, r float64) bool {
	r2 := r * r
	for _, id := range idxs {
		h := huts[id]
		dx := p.X - float64(h.X)
		dz := p.Z - float64(h.Z)
		if dx*dx+dz*dz < r2 {
			return true
		}
	}
	return false
}

func dist2(a, b hut) float64 {
	dx := float64(a.X - b.X)
	dz := float64(a.Z - b.Z)
	return dx*dx + dz*dz
}

func floorDiv(a, b int) int {
	q := a / b
	r := a % b
	if r != 0 && ((r < 0) != (b < 0)) {
		q--
	}
	return q
}

// ---------- union-find ----------

type uf struct {
	p []int
	r []uint8
}

func newUF(n int) *uf {
	p := make([]int, n)
	r := make([]uint8, n)
	for i := range p {
		p[i] = i
	}
	return &uf{p: p, r: r}
}

func (u *uf) find(x int) int {
	for u.p[x] != x {
		u.p[x] = u.p[u.p[x]]
		x = u.p[x]
	}
	return x
}

func (u *uf) union(a, b int) {
	ra, rb := u.find(a), u.find(b)
	if ra == rb {
		return
	}
	if u.r[ra] < u.r[rb] {
		u.p[ra] = rb
		return
	}
	if u.r[ra] > u.r[rb] {
		u.p[rb] = ra
		return
	}
	u.p[rb] = ra
	u.r[ra]++
}
