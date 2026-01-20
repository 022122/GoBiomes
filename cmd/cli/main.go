package main

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gobiomes"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("========================================")
	fmt.Println("      Gobiomes 交互式搜索工具 (Pure Go) ")
	fmt.Println("========================================")

	// 1. 版本选择
	fmt.Println("\n[1] 选择 Minecraft 版本:")
	versions := []struct {
		name string
		id   int
	}{
		{"1.21.1", gobiomes.MC_1_21_1},
		{"1.20.6", gobiomes.MC_1_20},
		{"1.19.2", gobiomes.MC_1_19_2},
		{"1.18.2", gobiomes.MC_1_18_2},
		{"1.17.1", gobiomes.MC_1_17_1},
		{"1.16.5", gobiomes.MC_1_16_5},
		{"1.12.2", gobiomes.MC_1_12_2},
		{"1.7.10", gobiomes.MC_1_7_10},
	}
	for i, v := range versions {
		fmt.Printf("%d) %s\n", i+1, v.name)
	}
	versionIdx := readInt(reader, "请选择版本序号 (默认 1): ", 1) - 1
	if versionIdx < 0 || versionIdx >= len(versions) {
		versionIdx = 0
	}
	mc := versions[versionIdx].id

	// 2. 种子输入
	seedStr := readString(reader, "\n[2] 输入世界种子 (默认 0): ", "0")
	seed, _ := strconv.ParseUint(seedStr, 10, 64)

	// 3. 搜索内容选择
	fmt.Println("\n[3] 选择搜索的结构:")
	structs := []struct {
		name string
		id   gobiomes.StructureType
	}{
		{"村庄 (Village)", gobiomes.Village},
		{"海底神殿 (Monument)", gobiomes.Monument},
		{"女巫小屋 (Swamp Hut)", gobiomes.SwampHut},
		{"沙漠神殿 (Desert Pyramid)", gobiomes.DesertPyramid},
		{"丛林神庙 (Jungle Pyramid)", gobiomes.JunglePyramid},
		{"试炼密室 (Trial Chambers)", gobiomes.TrialChambers},
		{"林地府邸 (Mansion)", gobiomes.Mansion},
		{"前哨站 (Outpost)", gobiomes.Outpost},
		{"要塞 (Stronghold) [仅限 1.18 前]", gobiomes.Stronghold},
	}
	for i, s := range structs {
		fmt.Printf("%d) %s\n", i+1, s.name)
	}
	structIdx := readInt(reader, "请选择结构序号 (默认 1): ", 1) - 1
	if structIdx < 0 || structIdx >= len(structs) {
		structIdx = 0
	}
	structID := structs[structIdx].id

	// 4. 搜索模式
	fmt.Println("\n[4] 选择搜索模式:")
	fmt.Println("1) 搜索最近的一个")
	fmt.Println("2) 搜索所有二联 (距离 < 128)")
	fmt.Println("3) 搜索所有三联 (距离 < 128)")
	fmt.Println("4) 搜索最近的一个二联")
	fmt.Println("5) 搜索最近的一个三联")
	mode := readInt(reader, "请选择模式序号 (默认 1): ", 1)

	// 5. 聚类距离 (仅模式 2, 3)
	clusterDist := 128.0
	if mode > 1 {
		clusterDist = float64(readInt(reader, "\n[5] 输入聚类最大距离 (方块, 默认 128): ", 128))
	}

	// 6. 搜索范围
	radius := readInt(reader, "\n[6] 输入搜索半径 (方块, 默认 5000): ", 5000)

	// 7. 线程设定
	workers := readInt(reader, fmt.Sprintf("\n[7] 设定搜索线程数 (默认 %d): ", runtime.NumCPU()), runtime.NumCPU())

	fmt.Printf("\n开始搜索: 版本=%s, 种子=%d, 结构=%s, 模式=%d, 半径=%d, 线程=%d\n",
		versions[versionIdx].name, seed, structs[structIdx].name, mode, radius, workers)

	doSearch(mc, seed, structID, mode, radius, workers, clusterDist)
}

type pos struct{ x, z int }

type cluster struct {
	ps   []pos
	dist float64 // 距离中心的距离
}

func doSearch(mc int, seed uint64, structID gobiomes.StructureType, mode int, radius int, workers int, clusterDist float64) {
	startTime := time.Now()
	finder := gobiomes.NewFinder(mc)
	sc, err := finder.GetStructureConfig(structID)
	if err != nil {
		fmt.Printf("错误: 无法获取结构配置: %v\n", err)
		return
	}

	regionSize := int(sc.RegionSize) * 16
	r := (radius + regionSize - 1) / regionSize
	if r < 0 {
		r = 0
	}

	foundChan := make(chan pos, 1024)
	var found []pos
	var foundCount int64
	var foundMu sync.Mutex
	var wgCollect sync.WaitGroup
	wgCollect.Add(1)
	go func() {
		defer wgCollect.Done()
		for p := range foundChan {
			foundMu.Lock()
			found = append(found, p)
			foundMu.Unlock()
			atomic.AddInt64(&foundCount, 1)
		}
	}()

	type task struct{ rx, rz int }
	taskChan := make(chan task, 1024)
	var wg sync.WaitGroup
	var regionsDone int64
	regionsTotal := int64((2*r + 1) * (2*r + 1))

	// 进度条显示
	stopProgress := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				done := atomic.LoadInt64(&regionsDone)
				pct := float64(done) * 100.0 / float64(regionsTotal)
				fCount := atomic.LoadInt64(&foundCount)
				fmt.Printf("\r正在扫描区域: %d/%d (%.2f%%) 找到: %d", done, regionsTotal, pct, fCount)
			case <-stopProgress:
				done := atomic.LoadInt64(&regionsDone)
				fCount := atomic.LoadInt64(&foundCount)
				fmt.Printf("\r正在扫描区域: %d/%d (%.2f%%) 找到: %d\n", done, regionsTotal, float64(done)*100.0/float64(regionsTotal), fCount)
				return
			}
		}
	}()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wFinder := gobiomes.NewFinder(mc)
			wGen := gobiomes.NewGenerator(mc, 0)
			wGen.ApplySeed(seed, gobiomes.DimOverworld)

			for {
				select {
				case <-ctx.Done():
					return
				case t, ok := <-taskChan:
					if !ok {
						return
					}
					p, err := wFinder.GetStructurePos(structID, seed, t.rx, t.rz)
					if err == nil && p != nil {
						distSq := int64(p.X)*int64(p.X) + int64(p.Z)*int64(p.Z)
						if distSq <= int64(radius)*int64(radius) {
							if wGen.IsViableStructurePos(structID, p.X, p.Z, 0) {
								foundChan <- pos{p.X, p.Z}
								if mode == 1 {
									cancel() // 找到一个就停止
									return
								}
							}
						}
					}
					atomic.AddInt64(&regionsDone, 1)
				}
			}
		}()
	}

	for rz := -r; rz <= r; rz++ {
		for rx := -r; rx <= r; rx++ {
			taskChan <- task{rx, rz}
		}
	}
	close(taskChan)
	wg.Wait()
	close(foundChan)
	wgCollect.Wait()
	close(stopProgress)

	if len(found) == 0 {
		fmt.Println("未找到结构。")
		return
	}

	// 排序，按距离中心由近到远
	sort.Slice(found, func(i, j int) bool {
		di := found[i].x*found[i].x + found[i].z*found[i].z
		dj := found[j].x*found[j].x + found[j].z*found[j].z
		return di < dj
	})

	if mode == 1 {
		p := found[0]
		fmt.Printf("找到最近的结构: x=%d, z=%d (距离中心: %.1f)\n", p.x, p.z, math.Sqrt(float64(p.x*p.x+p.z*p.z)))
		return
	}

	// 搜索多联
	var clusters []cluster

	if mode == 2 || mode == 4 {
		for i := 0; i < len(found); i++ {
			for j := i + 1; j < len(found); j++ {
				d := dist(found[i], found[j])
				if d < clusterDist {
					avgX := (found[i].x + found[j].x) / 2
					avgZ := (found[i].z + found[j].z) / 2
					centerDist := math.Sqrt(float64(avgX*avgX + avgZ*avgZ))
					clusters = append(clusters, cluster{
						ps:   []pos{found[i], found[j]},
						dist: centerDist,
					})
					if mode == 4 {
						goto foundCluster
					}
				}
			}
		}
	} else if mode == 3 || mode == 5 {
		for i := 0; i < len(found); i++ {
			for j := i + 1; j < len(found); j++ {
				for k := j + 1; k < len(found); k++ {
					d1 := dist(found[i], found[j])
					d2 := dist(found[j], found[k])
					d3 := dist(found[i], found[k])
					if d1 < clusterDist && d2 < clusterDist && d3 < clusterDist {
						avgX := (found[i].x + found[j].x + found[k].x) / 3
						avgZ := (found[i].z + found[j].z + found[k].z) / 3
						centerDist := math.Sqrt(float64(avgX*avgX + avgZ*avgZ))
						clusters = append(clusters, cluster{
							ps:   []pos{found[i], found[j], found[k]},
							dist: centerDist,
						})
						if mode == 5 {
							goto foundCluster
						}
					}
				}
			}
		}
	}

foundCluster:
	if len(clusters) == 0 {
		fmt.Println("\n未找到符合条件的聚类。")
		return
	}

	// 按距离中心排序
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].dist < clusters[j].dist
	})

	if mode == 4 || mode == 5 {
		c := clusters[0]
		fmt.Printf("\n找到最近的聚类 (距离中心: %.1f):\n", c.dist)
		for i, p := range c.ps {
			fmt.Printf("  点 %d: x=%d, z=%d\n", i+1, p.x, p.z)
		}
	} else {
		fmt.Printf("\n搜索完成，共找到 %d 处聚类。正在导出到 HTML...\n", len(clusters))
		exportToHTML(clusters, mc, seed, structID, startTime)
	}
}

func exportToHTML(clusters []cluster, mc int, seed uint64, structID gobiomes.StructureType, startTime time.Time) {
	filename := fmt.Sprintf("search_results_%d.html", time.Now().Unix())
	f, err := os.Create(filename)
	if err != nil {
		fmt.Printf("导出 HTML 失败: %v\n", err)
		return
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	fmt.Fprintln(w, `<!DOCTYPE html>
<html>
<head>
	   <meta charset="UTF-8">
	   <title>Gobiomes 搜索结果</title>
	   <style>
	       body { font-family: sans-serif; margin: 20px; background: #f0f0f0; }
	       .container { max-width: 1000px; margin: auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
	       h1 { color: #333; border-bottom: 2px solid #eee; padding-bottom: 10px; }
	       .meta { color: #666; margin-bottom: 20px; }
	       table { width: 100%; border-collapse: collapse; margin-top: 20px; }
	       th, td { padding: 12px; text-align: left; border-bottom: 1px solid #eee; }
	       th { background: #f8f8f8; }
	       tr:hover { background: #fafafa; }
	       .coords { font-family: monospace; background: #eee; padding: 2px 5px; border-radius: 3px; }
	       .dist { color: #888; font-size: 0.9em; }
	   </style>
</head>
<body>
<div class="container">
	   <h1>Gobiomes 搜索结果</h1>`)

	fmt.Fprintf(w, `<div class="meta">
	       <p>种子: <b>%d</b> | 版本 ID: %d | 结构类型: %v</p>
	       <p>搜索耗时: %v | 找到聚类总数: %d</p>
	   </div>`, seed, mc, structID, time.Since(startTime).Truncate(time.Millisecond), len(clusters))

	fmt.Fprintln(w, `<table>
	       <tr>
	           <th>#</th>
	           <th>坐标点</th>
	           <th>距离中心</th>
	       </tr>`)

	for i, c := range clusters {
		var coordStrs []string
		for _, p := range c.ps {
			coordStrs = append(coordStrs, fmt.Sprintf("<span class='coords'>/tp @s %d ~ %d</span>", p.x, p.z))
		}
		fmt.Fprintf(w, `<tr>
	           <td>%d</td>
	           <td>%s</td>
	           <td><span class="dist">%.1f</span></td>
	       </tr>`, i+1, strings.Join(coordStrs, " | "), c.dist)
	}

	fmt.Fprintln(w, `    </table>
</div>
</body>
</html>`)
	w.Flush()
	fmt.Printf("结果已保存至: %s\n", filename)
}

func dist(a, b struct{ x, z int }) float64 {
	dx := float64(a.x - b.x)
	dz := float64(a.z - b.z)
	return math.Sqrt(dx*dx + dz*dz)
}

func readString(reader *bufio.Reader, prompt string, def string) string {
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "" {
		return def
	}
	return text
}

func readInt(reader *bufio.Reader, prompt string, def int) int {
	str := readString(reader, prompt, strconv.Itoa(def))
	val, err := strconv.Atoi(str)
	if err != nil {
		return def
	}
	return val
}
