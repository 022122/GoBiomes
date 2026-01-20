package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"math"
	"os"
	"sort"
	"time"

	"github.com/scriptlinestudios/gobiomes"
	"github.com/scriptlinestudios/gobiomes/constants"
)

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

type htmlModel struct {
	Seed        uint64
	MC          int
	Area        [4]int
	MaxDist     float64
	Total       int
	Pairs       []pair
	PairsJSON   template.JS
	GeneratedAt string
}

func main() {
	var (
		seed  = flag.Uint64("seed", 20251223, "世界种子(64-bit)")
		mc    = flag.Int("mc", constants.MC_1_21_1, "版本常量，见 constants/versions.go")
		minX  = flag.Int("minx", -200000, "最小 X(block)")
		maxX  = flag.Int("maxx", 200000, "最大 X(block)")
		minZ  = flag.Int("minz", -200000, "最小 Z(block)")
		maxZ  = flag.Int("maxz", 200000, "最大 Z(block)")
		maxD  = flag.Float64("maxd", 150, "二联间隔最大距离(方块)")
		out   = flag.String("out", "trial_chambers_pairs.html", "输出 HTML 路径")
		quiet = flag.Bool("quiet", false, "安静模式（不打印进度条）")
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

	finder := gobiomes.NewFinder(*mc)
	gen := gobiomes.NewGenerator(*mc, 0)
	gen.ApplySeed(*seed, int(constants.DimOverworld))

	// Trial Chambers 属于 overworld 结构
	sc, err := finder.GetStructureConfig(int(constants.TrialChambers))
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

	// 1) 扫描区域内所有“可行”的试炼密室位置
	var (
		positions    = make([]pos, 0, 1024)
		totalRegions = int64(rx1-rx0+1) * int64(rz1-rz0+1)
		doneRegions  int64
		lastPrint    = time.Now()
	)

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

			doneRegions++
			if !*quiet && time.Since(lastPrint) > 200*time.Millisecond {
				printProgress("scan regions", doneRegions, totalRegions)
				lastPrint = time.Now()
			}
		}
	}
	if !*quiet {
		printProgress("scan regions", totalRegions, totalRegions)
		fmt.Println()
	}

	if len(positions) == 0 {
		fmt.Println("no Trial Chambers found in area")
		return
	}

	// 2) 二联查找：网格加速
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

	pairs := make([]pair, 0, 1024)
	var checked int64
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
						// 权重机制：越近越大。这里用 (maxD - d) 线性权重。
						w := (*maxD - d)
						pairs = append(pairs, pair{A: a, B: b, Dist: d, Weight: w})
					}
					checked++
				}
			}
		}

		if !*quiet && time.Since(lastPrint) > 300*time.Millisecond {
			// 这里 total 用 positions^2/2 的近似会很大，不太准，改成“已检查候选数”。
			fmt.Printf("\r[calc pairs] checked=%d pairs=%d", checked, len(pairs))
			lastPrint = time.Now()
		}
	}
	if !*quiet {
		fmt.Printf("\r[calc pairs] checked=%d pairs=%d\n", checked, len(pairs))
	}

	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Weight != pairs[j].Weight {
			return pairs[i].Weight > pairs[j].Weight
		}
		return pairs[i].Dist < pairs[j].Dist
	})

	pairsJSONBytes, _ := json.Marshal(pairs)

	m := htmlModel{
		Seed:        *seed,
		MC:          *mc,
		Area:        [4]int{*minX, *maxX, *minZ, *maxZ},
		MaxDist:     *maxD,
		Total:       len(positions),
		Pairs:       pairs,
		PairsJSON:   template.JS(pairsJSONBytes),
		GeneratedAt: time.Now().Format(time.RFC3339),
	}

	if err := writeHTML(*out, m); err != nil {
		panic(err)
	}

	fmt.Printf("seed=%d mc=%d area=[%d,%d]x[%d,%d] chambers=%d pairs<=%.0f=%d\n",
		m.Seed, m.MC, m.Area[0], m.Area[1], m.Area[2], m.Area[3], m.Total, m.MaxDist, len(m.Pairs))
	fmt.Printf("已输出 HTML: %s\n", *out)
}

func writeHTML(path string, m htmlModel) error {
	tpl := template.Must(template.New("page").Parse(pageHTML))
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return tpl.Execute(f, m)
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

const pageHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Trial Chambers Pairs</title>
  <style>
    body { font-family: ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, Arial; margin: 20px; }
    table { border-collapse: collapse; width: 100%; }
    th, td { border: 1px solid #ddd; padding: 8px; }
    th { background: #f5f5f5; text-align: left; }
    .meta { color: #444; margin-bottom: 12px; }
    .small { color: #666; font-size: 12px; }
    input { padding: 6px; width: 220px; }
    .row { display:flex; gap: 12px; align-items: center; flex-wrap: wrap; }
    .pill { padding: 2px 8px; border-radius: 999px; background: #eef; display: inline-block; }
    button { padding: 6px 10px; }
  </style>
</head>
<body>
  <h1>试炼密室二联（间隔≤{{printf "%.0f" .MaxDist}}）</h1>
  <div class="meta">
    <div>seed=<span class="pill">{{.Seed}}</span> mc=<span class="pill">{{.MC}}</span> area=<span class="pill">[{{index .Area 0}},{{index .Area 1}}]×[{{index .Area 2}},{{index .Area 3}}]</span></div>
    <div class="small">generatedAt={{.GeneratedAt}} | chambers={{.Total}} | pairs={{len .Pairs}}</div>
  </div>

  <div class="row">
    <label>最小权重(越大越近)：<input id="minW" type="number" step="1" value="0"></label>
    <label>最大距离：<input id="maxD" type="number" step="1" value="{{printf "%.0f" .MaxDist}}"></label>
    <button onclick="render()">筛选</button>
    <span class="small">提示：权重=MaxDist-距离；越近越靠上</span>
  </div>

  <table>
    <thead>
      <tr>
        <th>#</th>
        <th>A (x,z)</th>
        <th>B (x,z)</th>
        <th>距离</th>
        <th>权重</th>
      </tr>
    </thead>
    <tbody id="tbody"></tbody>
  </table>

  <script>
    const data = {{.PairsJSON}};

    function render() {
      const minW = Number(document.getElementById('minW').value || 0);
      const maxD = Number(document.getElementById('maxD').value || 0);

      const tbody = document.getElementById('tbody');
      tbody.innerHTML = '';

      let idx = 0;
      for (const p of data) {
        if (p.weight < minW) continue;
        if (maxD > 0 && p.dist > maxD) continue;
        idx++;
        const tr = document.createElement('tr');
        tr.innerHTML =
          '<td>' + idx + '</td>' +
          '<td>' + p.a.x + ', ' + p.a.z + '</td>' +
          '<td>' + p.b.x + ', ' + p.b.z + '</td>' +
          '<td>' + p.dist.toFixed(2) + '</td>' +
          '<td>' + p.weight.toFixed(2) + '</td>';
        tbody.appendChild(tr);
      }
    }

    render();
  </script>
</body>
</html>`
