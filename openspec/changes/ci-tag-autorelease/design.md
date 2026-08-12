## Context

项目是 Wails v2 桌面应用（Windows 单平台发布：`input_*`/`tray_*`/`notification_*`/`lock_*` 均为 Windows 实现 + 其余平台 stub）。`wails build` 是唯一正确构建入口：编译前端、`go:embed` 嵌入 `frontend/dist`、注入图标/清单/版本信息、可选生成 NSIS 安装器。GoReleaser 原生 `go build` 无法完成这些，因此不能走默认 builder。

版本信息链路已确认：`wails.json` 的 `info.productVersion` → Wails 构建时模板化写入 `build/windows/info.json`（exe 文件属性）与 NSIS `project.nsi`（安装器版本）。tag 只需写入这一处即全链路生效。

## Goals / Non-Goals

**Goals:**
- 打 `v*` tag 自动发布：构建、版本注入、GitHub Release 全部自动化。
- 版本号唯一来源是 git tag。
- 发布产物为单 exe。
- 分支 push 不触发流水线。

**Non-Goals:**
- 多平台矩阵发布（macOS/Linux 不在范围）。
- 代码签名（Authenticode）与公证。
- 分发到 Homebrew/Scoop/Docker 等额外渠道。
- 生成 NSIS 安装器（`wails build -nsis` 依赖 makensis 不在 runner PATH，且单 exe 已满足使用场景）。
- 生成独立 checksums.txt（GitHub 对每个 Release 资产自动附带 SHA256）。
- main 分支 push 触发构建验证。
- 修改运行时代码或前端逻辑。

## Decisions

**D1. 单 `windows-latest` job，不跨平台交叉编译**
Wails 官方不支持从 Linux 交叉编译 Windows exe；本项目也只发布 Windows。→ 在 Windows runner 上原生构建。

**D2. 不使用 GoReleaser，纯 GitHub Actions + `gh` CLI**
曾尝试 GoReleaser，但本项目是单平台、单产物，用不上其核心能力（多平台矩阵、多渠道分发），反而引入 `builder: prebuilt` 为 Pro 专属、go builder + post hook 覆盖 hack、`wails build` 污染 git 状态需恢复等额外复杂度。放弃 GoReleaser，改用 windows-latest **预装的 `gh` CLI**：
- `gh release create <tag> 产物... --generate-notes` 创建 Release 并上传产物；
- changelog 由 GitHub 原生生成（基于两次 release 之间的 PR/commit）；
- 每个资产 GitHub 自动附带 SHA256。

**D3. 版本从 tag 写入 `wails.json` 的 `info.productVersion`**
`{{.Info.ProductVersion}}` 模板唯一数据源是 `wails.json`。CI 中用 node 脚本 `TrimStart('v')` 后写入并序列化回 JSON，随后 `wails build` 自动注入 exe 与安装器。替代方案（`-ldflags -X main.version`）无法写入 Windows 文件属性，弃用。

**D4. 只发布单 exe，不生成 NSIS 安装器**
`wails build -nsis` 依赖 makensis（NSIS 编译工具），而 windows-latest runner 未将其加入 PATH，wails 静默跳过安装器生成导致发布失败。本项目是单平台小工具，Windows 10/11 普遍自带 WebView2 运行时，单 exe 已满足使用场景。直接 `wails build -clean` 产出 `build/bin/health-tool.exe`，`gh release create` 上传单个 exe。

**D5. 仅 tag 触发，分支 push 不触发流水线**
`on.push.tags` 只监听 `v*` tag；分支 push 不触发，避免每次提交都空跑构建。`gh release create` 只上传单 exe。

## Risks / Trade-offs

- [`wails build` 会修改已跟踪文件（wailsjs 绑定、NSIS 模板、go.mod）] → 本方案不检查 git 状态（`gh` CLI 不校验 dirty），无需恢复；构建后直接上传产物。
- [Wails CLI 版本漂移导致构建失败] → `go install` 固定 `@v2.13.0`，与项目依赖一致。
- [`frontend/dist` 被 gitignore，CI 需要网络安装 npm 依赖] → wails.json 已配置 `frontend:build`，`wails build` 自动执行；npm 源需在 CI 可达。
- [changelog 依赖 GitHub 的 PR/commit 生成] → 首次发布会列出 tag 前全部提交，后续只列两次 release 之间；属 GitHub 原生行为，可接受。

## Migration Plan

- 首次发布：打 `v0.1.0` tag，观察 GitHub Actions 首个 run 与 Release 产物。
- 回滚：删除 tag 与对应 Release 即可，不涉及线上运行代码。

## Open Questions

- 是否需要 `-nsis` 安装器之外的应用内升级检查？当前不涉及。
- 后续是否要接代码签名？是则需另购证书并作为 secrets 注入，属 Non-Goal。
