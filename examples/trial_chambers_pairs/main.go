package main

import (
	"container/heap"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/scriptlinestudios/gobiomes"
	"github.com/scriptlinestudios/gobiomes/constants"
)

// trial_chambers_pairs
// - 搜索 40w*40w 区域内（默认 [-200000,200000] x [-200000,200000]）的试炼密室(TrialChambers)
// - 找“二联”：两处结构生成尝试点的距离 <= maxd（默认 150）
// - 权重机制：weight = maxd - dist（越近越靠上）
// - 输出：JSON 或 Markdown（避免 HTML 打开卡顿）
//
// 重要说明：
// 试炼密室的生成尝试点由“区域网格”决定，最小理论间距大概率远大于 150。
// 程序会打印 regionSize/chunkRange 推导的 minAxisDist，帮助判断 pairs=0 是否正常。

type pos struct {
	X int `json:"x"`
	Z int `json:"z"`
}

type pair struct {
	A      pos     `json:"a"`
	B      pos     `json:"b"`
	Dist   float64 `json:"dist"`
	Weight float64 `json:"weight"`
}

type outJSON struct {
	Seed        uint64  `json:"seed"`
	MC          int     `json:"mc"`
	Area        [4]int  `json:"area"` // minx,maxx,minz,maxz
	MaxDist     float64 `json:"maxDist"`
	MinAxisDist float64 `json:"minAxisDist"`
	Chambers    int     `json:"chambers"`
	PairsTotal  int64   `json:"pairsTotal"`
	TopK        int     `json:"topK"`
	TopPairs    []pair  `json:"topPairs"`
	GeneratedAt string  `json:"generatedAt"`
}

func main() {
	var (
		seed    = flag.Uint64("seed", 20251223, "世界种子(64-bit)")
		mc      = flag.Int("mc", constants.MC_1_21_1, "版本常量，见 constants/versions.go")
		minX    = flag.Int("minx", -200000, "最小 X(block)")
		maxX    = flag.Int("maxx", 200000, "最大 X(block)")
		minZ    = flag.Int("minz", -200000, "最小 Z(block)")
		maxZ    = flag.Int("maxz", 200000, "最大 Z(block)")
		maxD    = flag.Float64("maxd", 150, "二联间隔最大距离(方块)")
		format  = flag.String("format", "json", "输出格式: json 或 md")
		outPath = flag.String("out", "trial_chambers_pairs.json", "输出文件路径")
		topK    = flag.Int("top", 5000, "只输出 TopK 对（按权重排序）；计算时仍统计 pairsTotal")
		quiet   = flag.Bool("quiet", false, "安静模式（不打印进度条/提示）")
	)
	flag.Parse()

	if *minX > *maxX {
		*minX, *maxX = *maxX, *minX
	}
	if *minZ > *maxZ {
		*minZ, *maxZ = *maxZ, *minZ
	}
	if *maxD <= 0 {
		panic("maxd must be > 0")
	}
	if *topK < 1 {
		*topK = 1
	}

	finder := gobiomes.NewFinder(*mc)
	gen := gobiomes.NewGenerator(*mc, 0)
	gen.ApplySeed(*seed, int(constants.DimOverworld))

	sc, err := finder.GetStructureConfig(int(constants.TrialChambers))
	if err != nil {
		panic(err)
	}
	regionBlocks := int(sc.RegionSize) * 16
	if regionBlocks <= 0 {
		panic("invalid region size")
	}

	// 理论最小轴向间距（blocks）≈ (regionSize - (chunkRange-1)) * 16
	minAxisDist := float64(int(sc.RegionSize)-int(sc.ChunkRange)+1) * 16
	if !*quiet {
		fmt.Printf("TrialChambers config: regionSize=%d chunks, chunkRange=%d chunks, minAxisDist≈%.0f blocks\n",
			sc.RegionSize, sc.ChunkRange, minAxisDist)
		if minAxisDist > *maxD {
			fmt.Printf("提示：minAxisDist(≈%.0f) > maxd(%.0f)，理论上很难/不可能存在二联，因此 pairs=0 可能是正常结果。\n",
				minAxisDist, *maxD)
		}
	}

	rx0 := floorDiv(*minX, regionBlocks)
	rx1 := floorDiv(*maxX, regionBlocks)
	rz0 := floorDiv(*minZ, regionBlocks)
	rz1 := floorDiv(*maxZ, regionBlocks)

	// 1) 扫描区域内所有“可行”的试炼密室位置
	positions := make([]pos, 0, 1024)
	regionsTotal := int64(rx1-rx0+1) * int64(rz1-rz0+1)
	var regionsDone int64
	lastPrint := time.Now()

	for rz := rz0; rz <= rz1; rz++ {
		for rx := rx0; rx <= rx1; rx++ {
			p, err := finder.GetStructurePos(int(constants.TrialChambers), *seed, rx, rz)
			if err != nil {
				panic(err)
			}
			if p != nil {
				if p.X >= *minX && p.X <= *maxX && p.Z >= *minZ && p.Z <= *maxZ {
					if gen.IsViableStructurePos(int(constants.TrialChambers), p.X, p.Z, 0) {
						positions = append(positions, pos{X: p.X, Z: p.Z})
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

	if len(positions) == 0 {
		fmt.Println("no Trial Chambers found in area")
		return
	}

	// 2) 二联查找：网格加速 + TopK 小顶堆（避免输出/内存爆炸）
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

	pq := make(pairHeap, 0, *topK)
	var checked int64
	var pairsTotal int64
	lastPrint = time.Now()

	for i, a := range positions {
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
					d := dist(a, b)
					if d <= *maxD {
						pairsTotal++
						w := (*maxD - d)
						pushTopK(&pq, pair{A: a, B: b, Dist: d, Weight: w}, *topK)
					}
					checked++
				}
			}
		}

		if !*quiet && time.Since(lastPrint) > 300*time.Millisecond {
			fmt.Printf("\r[calc pairs] checked=%d pairsTotal=%d topK=%d", checked, pairsTotal, len(pq))
			lastPrint = time.Now()
		}
	}
	if !*quiet {
		fmt.Printf("\r[calc pairs] checked=%d pairsTotal=%d topK=%d\n", checked, pairsTotal, len(pq))
	}

	// 堆 -> 切片 -> 按权重降序输出
	topPairs := make([]pair, 0, len(pq))
	for pq.Len() > 0 {
		p := heap.Pop(&pq).(pair)
		topPairs = append(topPairs, p)
	}
	// pq 是小顶堆，pop 出来从“最差”到“最好”，这里反转
	reversePairs(topPairs)

	// 再做一次稳定排序（确保同权重按距离升序）
	sort.SliceStable(topPairs, func(i, j int) bool {
		if topPairs[i].Weight != topPairs[j].Weight {
			return topPairs[i].Weight > topPairs[j].Weight
		}
		return topPairs[i].Dist < topPairs[j].Dist
	})

	if err := writeOut(*format, *outPath, outJSON{
		Seed:        *seed,
		MC:          *mc,
		Area:        [4]int{*minX, *maxX, *minZ, *maxZ},
		MaxDist:     *maxD,
		MinAxisDist: minAxisDist,
		Chambers:    len(positions),
		PairsTotal:  pairsTotal,
		TopK:        *topK,
		TopPairs:    topPairs,
		GeneratedAt: time.Now().Format(time.RFC3339),
	}); err != nil {
		panic(err)
	}

	fmt.Printf("seed=%d mc=%d area=[%d,%d]x[%d,%d] chambers=%d pairs<=%.0f total=%d top=%d\n",
		*seed, *mc, *minX, *maxX, *minZ, *maxZ, len(positions), *maxD, pairsTotal, len(topPairs))
	fmt.Printf("已输出: %s\n", *outPath)
}

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
	s := "# 试炼密室二联（Trial Chambers Pairs）\n\n"
	s += fmt.Sprintf("- seed: `%d`\n", o.Seed)
	s += fmt.Sprintf("- mc: `%d`\n", o.MC)
	s += fmt.Sprintf("- area: `[%d,%d]×[%d,%d]`\n", o.Area[0], o.Area[1], o.Area[2], o.Area[3])
	s += fmt.Sprintf("- maxDist: `%.0f`\n", o.MaxDist)
	s += fmt.Sprintf("- minAxisDist(estimated): `%.0f`\n", o.MinAxisDist)
	s += fmt.Sprintf("- chambers: `%d`\n", o.Chambers)
	s += fmt.Sprintf("- pairsTotal: `%d`\n", o.PairsTotal)
	s += fmt.Sprintf("- topK: `%d`\n", o.TopK)
	s += fmt.Sprintf("- generatedAt: `%s`\n\n", o.GeneratedAt)

	s += "## Top Pairs\n\n"
	s += "| # | A(x,z) | B(x,z) | dist | weight |\n"
	s += "|---:|---|---|---:|---:|\n"
	for i, p := range o.TopPairs {
		s += fmt.Sprintf("| %d | %d,%d | %d,%d | %.2f | %.2f |\n",
			i+1, p.A.X, p.A.Z, p.B.X, p.B.Z, p.Dist, p.Weight)
	}
	return s
}

func dist(a, b pos) float64 {
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
	// if p is better than the worst (root), replace
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
