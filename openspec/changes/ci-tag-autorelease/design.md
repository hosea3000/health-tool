## Context

项目是 Wails v2 桌面应用（Windows 单平台发布：`input_*`/`tray_*`/`notification_*`/`lock_*` 均为 Windows 实现 + 其余平台 stub）。`wails build` 是唯一正确构建入口：编译前端、`go:embed` 嵌入 `frontend/dist`、注入图标/清单/版本信息、可选生成 NSIS 安装器。GoReleaser 原生 `go build` 无法完成这些，因此不能走默认 builder。

版本信息链路已确认：`wails.json` 的 `info.productVersion` → Wails 构建时模板化写入 `build/windows/info.json`（exe 文件属性）与 NSIS `project.nsi`（安装器版本）。tag 只需写入这一处即全链路生效。

## Goals / Non-Goals

**Goals:**
- 打 `v*` tag 自动发布：构建、版本注入、changelog、checksums、GitHub Release 全部自动化。
- 版本号唯一来源是 git tag。
- 发布产物含 exe 归档与 NSIS 安装器。

**Non-Goals:**
- 多平台矩阵发布（macOS/Linux 不在范围）。
- 代码签名（Authenticode）与公证。
- 分发到 Homebrew/Scoop/Docker 等额外渠道。
- 修改运行时代码或前端逻辑。

## Decisions

**D1. 单 `windows-latest` job，不跨平台交叉编译**
Wails 官方不支持从 Linux 交叉编译 Windows exe；本项目也只发布 Windows。→ 在 Windows runner 上原生构建。

**D2. GoReleaser 使用 Go builder + post hook 覆盖产物**
`builder: prebuilt` 是 GoReleaser **Pro 专属**功能，开源版不可用（YAML 直接报 `field prebuilt not found`），弃用。
开源版采用：`builder: go`（`goos: windows`、`goarch: amd64`）正常编译，随后 **post hook** 用 `wails build` 的真实产物覆盖 GoReleaser 的输出：`cp build/bin/health-tool.exe "{{ .Path }}"`。已实测验证：覆盖后的 exe 正确进入 zip 与 `checksums.txt`。
前提：真实项目 `main.go` 有 `//go:embed all:frontend/dist`，go build 需 `frontend/dist` 存在 —— CI 中先跑 `wails build` 生成它，再跑 goreleaser。

**D3. 版本从 tag 写入 `wails.json` 的 `info.productVersion`**
`{{.Info.ProductVersion}}` 模板唯一数据源是 `wails.json`。CI 中用 node 脚本 `TrimStart('v')` 后写入并序列化回 JSON，随后 `wails build` 自动注入 exe 与安装器。替代方案（`-ldflags -X main.version`）无法写入 Windows 文件属性，弃用。

**D4. NSIS 安装器作为 `extra_files` 上传，不进 archive**
安装器由 `wails build -nsis` 产出为 `build/bin/*-installer.exe`（如 `health-tool-amd64-installer.exe`）。GoReleaser 的 archive 只归档构建二进制；安装器用 `release.extra_files` glob 附带上传，避免多文件冲突。

**D5. 归档格式用 zip**
单 exe 直接作为 Release 资产会被浏览器/杀软误伤，zip 更稳妥，且为 checksums 提供清晰入口。

## Risks / Trade-offs

- [`goreleaser release --clean` 清空 `dist/` 但不影响 `build/bin/`] → prebuilt 产物路径在 `build/bin/`，生命周期无冲突。
- [Wails CLI 版本漂移导致构建失败] → 不固定 CLI 安装版本，失败即红；后续可按需 pin。
- [`frontend/dist` 被 gitignore，CI 需要网络安装 npm 依赖] → wails.json 已配置 `frontend:build`，`wails build` 自动执行；npm 源需在 CI 可达。
- [Changelog 质量依赖 commit message 规范] → 当前仓库 commit 为中文自由格式，changelog 将按 commit 原样聚合，后续可引入 conventional commits 增强。
- [prebuilt builder 的二进制须与 `goos`/`goarch` 声明一致] → 本方案改用 go builder + post hook 覆盖，配置固定 `windows/amd64`，与 `wails build` 默认产物一致。

## Migration Plan

- 首次发布：打 `v0.1.0` tag，观察 GitHub Actions 首个 run 与 Release 产物。
- 回滚：删除 tag 与对应 Release 即可，不涉及线上运行代码。

## Open Questions

- 是否需要 `-nsis` 安装器之外的应用内升级检查？当前不涉及。
- 后续是否要接代码签名？是则需另购证书并作为 secrets 注入，属 Non-Goal。
