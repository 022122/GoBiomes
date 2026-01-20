# GoBiomes（Go 版本）

本目录 [`GO/`](GO/:1) 是对原项目（Python + C 绑定）的 Go 语言重构实现：

- 保留核心能力：Biome 生成、结构位置计算、结构可行性判定、Java Random（MC RNG）
- **不依赖原项目代码布局**：cubiomes C 源码已 vendored 到 [`GO/cubiomes/`](GO/cubiomes/:1)，Go 封装在 [`GO/`](GO/:1)

## 功能对照

| 原 Python API | Go API |
|---|---|
| `Generator(version, flags)` | [`gobiomes.NewGenerator()`](GO/generator.go:33) |
| `generator.apply_seed(seed, dim)` | [`(*gobiomes.Generator).ApplySeed()`](GO/generator.go:45) |
| `generator.get_biome_at(scale,x,y,z)` | [`(*gobiomes.Generator).GetBiomeAt()`](GO/generator.go:52) |
| `generator.gen_biomes(range)` | [`(*gobiomes.Generator).GenBiomes()`](GO/generator.go:59) |
| `generator.is_viable_structure_pos(st, x, z, flags)` | [`(*gobiomes.Generator).IsViableStructurePos()`](GO/generator.go:81) |
| `Finder(version)` | [`gobiomes.NewFinder()`](GO/finder.go:55) |
| `finder.get_structure_config(st)` | [`(*gobiomes.Finder).GetStructureConfig()`](GO/finder.go:61) |
| `finder.chunk_generate_rnd(seed,cx,cz)` | [`(*gobiomes.Finder).ChunkGenerateRnd()`](GO/finder.go:72) |
| `finder.get_structure_pos(st, seed, rx, rz)` | [`(*gobiomes.Finder).GetStructurePos()`](GO/finder.go:80) |
| `Rng(seed)` | [`gobiomes.NewRng()`](GO/rng.go:18) |

## 构建要求

- Go 1.21+
- **需要启用 CGO**（因为底层使用 vendored 的 C 版 cubiomes）：
  - Windows：需要安装可用的 C 编译器（例如 mingw-w64）并确保在 PATH

## 快速验证

在仓库根目录执行：

```powershell
cd .\GO
# 仅编译/检查
go test ./...
```

cubiomes C 源码通过 [`GO/cubiomes_all.c`](GO/cubiomes_all.c:1) 作为单个 translation unit 编译；
部分 `static inline` 函数通过 [`GO/helpers.c`](GO/helpers.c:1) / [`GO/helpers.h`](GO/helpers.h:1) 提供可链接包装。

## 示例

示例代码位于 [`GO/cmd/`](GO/cmd/:1)。

### 1) 搜索蘑菇岛（mushroom_fields）

运行：

```powershell
cd .\GO
go run .\cmd\mushroom\main.go
```

对应示例中会使用：
- 版本常量：[`constants.MC_1_21_1`](GO/constants/versions.go:78)
- 维度常量：[`constants.DimOverworld`](GO/constants/dimensions.go:6)
- 生物群系常量：[`constants.MushroomFields`](GO/constants/biomes.go:30)

### 2) 搜索前哨站（Outpost）种子

运行：

```powershell
cd .\GO
go run .\cmd\outpost\main.go
```

对应示例中会使用：
- 结构常量：[`constants.Outpost`](GO/constants/structures.go:17)

## 包结构

- Go 主包：[`package gobiomes`](GO/generator.go:3)
- 常量：[`package constants`](GO/constants/versions.go:1)

关键文件：
- [`GO/generator.go`](GO/generator.go:1)
- [`GO/finder.go`](GO/finder.go:1)
- [`GO/rng.go`](GO/rng.go:1)
- [`GO/types.go`](GO/types.go:1)
- [`GO/cgo.go`](GO/cgo.go:1)
- [`GO/cubiomes_all.c`](GO/cubiomes_all.c:1)

