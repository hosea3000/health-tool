## Why

目前项目没有任何发布流水线，打 tag 后需要手动 `wails build` 并手动把产物挂到 GitHub Release，过程繁琐且版本号靠手工维护。需要一条"打 tag 即自动发布"的流水线。

## What Changes

- 新增 `.github/workflows/release.yml`：打 `v*` tag 时在 `windows-latest` runner 上自动构建并发布。
- 新增 `.goreleaser.yaml`：以 `prebuilt` builder 模式对 `wails build` 产物做归档、生成 changelog 与 checksums，并创建 GitHub Release。
- 构建流程：checkout（全历史）→ setup Go/Node → 从 tag 提取版本写入 `wails.json` 的 `info.productVersion` → `wails build -clean -nsis` 产出 exe 与 NSIS 安装器 → `goreleaser release`。
- `wails.json` 补充 `info` 段（`productVersion` 默认 `1.0.0`，CI 每次用 tag 覆盖）。
- 发布产物：`health-tool.zip`（含 exe）+ NSIS `setup.exe`（extra_files）+ `checksums.txt` + 基于 git log 的 changelog。
- 版本号完全由 tag 驱动，不手工维护。

## Capabilities

### New Capabilities

- `ci-autorelease`: 打 `v*` tag 时自动构建 Wails 应用并通过 GoReleaser 发布 Windows 产物到 GitHub Release，包含 changelog 与 checksums。

### Modified Capabilities

- 无（纯构建/发布基础设施，不改运行时行为）。

## Impact

- 新增文件：`.github/workflows/release.yml`、`.goreleaser.yaml`。
- 修改文件：`wails.json`（增加 `info` 段）。
- 依赖：CI 环境需要 Go、Node.js、Wails CLI；引入 GoReleaser v2 作为发布工具。
- 仓库：需要 `contents: write` 权限的 GitHub Actions token 以创建 Release。
- 不触碰运行时代码与前端逻辑；`frontend/dist` 由 CI 中的 `wails build` 现场构建（gitignore 保持不变）。
