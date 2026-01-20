# Gobiomes (Pure Go)

这是一个完全使用 Go 语言重写的 Minecraft 生物群系和结构查找库，移植自著名的 C 语言库 [cubiomes](https://github.com/Cubitect/cubiomes)。

## 特性

- **不依赖 CGO**: 纯 Go 实现，易于跨平台编译和集成。
- **全版本支持**: 支持从 Beta 1.8 到最新版本 (1.21+) 的生物群系生成逻辑。
- **结构查找**: 支持村庄、要塞、神庙、女巫小屋等所有原版结构的生成位置计算。
- **高性能**: 针对 Go 进行了优化，支持并发搜索。

## 安装

```bash
go get github.com/scriptlinestudios/gobiomes
```

## 快速开始

```go
package main

import (
	"fmt"
	"github.com/scriptlinestudios/gobiomes"
)

func main() {
	mc := gobiomes.MC_1_21_1
	seed := uint64(12345)

	// 初始化查找器和生成器
	finder := gobiomes.NewFinder(mc)
	gen := gobiomes.NewGenerator(mc, 0)
	gen.ApplySeed(seed, gobiomes.DimOverworld)

	// 查找最近的村庄
	rx, rz := 0, 0 // 区域坐标
	pos, _ := finder.GetStructurePos(gobiomes.Village, seed, rx, rz)
	
	if pos != nil && gen.IsViableStructurePos(gobiomes.Village, pos.X, pos.Z, 0) {
		fmt.Printf("在 (%d, %d) 找到村庄\n", pos.X, pos.Z)
	}

	// 采样生物群系
	biome := gen.GetBiomeAt(1, 100, 64, 100)
	fmt.Printf("坐标 (100, 100) 的生物群系 ID: %d\n", biome)
}
```

## 目录结构

- `GO/`: 核心库代码。
- `GO/examples/`: 示例程序。
- `GO/constants/`: 生物群系、结构、版本等常量定义。

## 许可证

本项目遵循 MIT 许可证。