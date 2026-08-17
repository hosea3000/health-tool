# Proposal: 版本更新检查

## Why

应用以 Windows exe 形式发布，用户无法获知 GitHub Release 上是否有新版本。当前版本号只存在于 `wails.json`（运行时不可读），也没有任何更新提示入口。需要提供手动检查更新的能力，发现新版本后引导用户前往 GitHub Release 页面查看。

## What Changes

- 主包新增运行时版本号变量（默认 `dev`），CI 发布构建时通过 `-ldflags -X main.version` 注入 tag 版本号（去 `v` 前缀）；本地开发构建保持 `dev` 不变。
- 后端新增 `CheckForUpdates()` 绑定方法：请求 GitHub API `repos/hosea3000/health-tool/releases/latest`，解析最新 tag 与当前版本做语义化比较，返回结构化结果（`up-to-date` / `update-available` / `error` 三态）。
- `dev` 版本跳过检查，不请求网络。
- dashboard 顶栏「健康工具箱」旁新增设置入口（齿轮图标，内联 SVG），点击进入新的设置页。
- 新增设置页 `frontend/src/views/settings.js`（复用 detail 视图的 shell 结构）：含「关于」区（应用名 + 当前版本号）与「更新」区（检查更新按钮 + 结果反馈）。
- 发现新版本时，页面内反馈区显示最新版本号并提供「前往 GitHub 查看」按钮，点击调用 `runtime.BrowserOpenURL` 打开 release 页面。
- 不引入自动检查；无「忽略此版本」功能；失败时页面内提示原因。

## Capabilities

### New Capabilities
- `version-update-check`: 手动版本更新检查能力——运行时当前版本号的获取、GitHub Release 查询、语义化版本比较、结果三态反馈，以及设置页入口与跳转行为。

### Modified Capabilities
- `ci-autorelease`: 发布流程新增运行时版本号注入——`wails build` 步骤通过 ldflags 从 tag 派生并注入 `main.version`。

## Impact

- 代码：`main.go`（版本号变量）、新 `updater.go`（检查逻辑 + semver 比较）、`app.go`（`CheckForUpdates` 绑定）、`model`（`UpdateCheckResult` 数据契约）、前端 `main.js` / `views/dashboard.js`（顶栏入口）/ 新 `views/settings.js` / `style.css`。
- CI：`.github/workflows/release.yml` 的 `Wails build` 步骤追加 ldflags。
- 依赖：无新增第三方依赖（`net/http` 标准库）。
- 数据持久化：无新增。
- 平台差异：检查与跳转均跨平台；发布目标仍为 Windows。