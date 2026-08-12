## Why

目前项目没有任何发布流水线，打 tag 后需要手动 `wails build` 并手动把产物挂到 GitHub Release，过程繁琐且版本号靠手工维护。需要一条"打 tag 即自动发布"的流水线。

## What Changes

- 新增 `.github/workflows/release.yml`：打 `v*` tag 时在 `windows-latest` runner 上自动构建并发布；main 分支 push 触发构建验证。
- 构建流程：checkout（全历史）→ setup Go/Node → 安装 Wails CLI → 从 tag 提取版本写入 `wails.json` 的 `info.productVersion` → `wails build -clean -nsis` 产出 exe 与 NSIS 安装器。
- 发布：用 GitHub 预装的 `gh` CLI 的 `gh release create --generate-notes` 创建 Release 并上传产物；changelog 由 GitHub 原生生成，每个资产自动附带 SHA256。
- `wails.json` 补充 `info` 段（`productVersion` 默认 `1.0.0`，CI 每次用 tag 覆盖）。
- 发布产物：`health-tool.exe` + NSIS `*-installer.exe`，由 `gh release create` 一并上传。
- 版本号完全由 tag 驱动，不手工维护。

## Capabilities

### New Capabilities

- `ci-autorelease`: 打 `v*` tag 时自动构建 Wails 应用并发布 Windows 产物到 GitHub Release，changelog 由 GitHub 原生生成。

### Modified Capabilities

- 无（纯构建/发布基础设施，不改运行时行为）。

## Impact

- 新增文件：`.github/workflows/release.yml`。
- 删除文件：`.goreleaser.yaml`（改用纯 GitHub Actions + `gh` CLI，不引入 GoReleaser）。
- 修改文件：`wails.json`（增加 `info` 段）。
- 依赖：CI 环境需要 Go、Node.js、Wails CLI；`gh` CLI 由 windows-latest 预装。
- 仓库：需要 `contents: write` 权限的 GitHub Actions token 以创建 Release。
- 不触碰运行时代码与前端逻辑；`frontend/dist` 由 CI 中的 `wails build` 现场构建（gitignore 保持不变）。
