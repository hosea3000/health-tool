## Why

当前「检查更新」只做到发现新版本后跳转 GitHub,用户必须手动下载 exe、退出应用、覆盖文件、重新启动,过程繁琐且容易出错。需要把「下载 → 替换 → 重启」闭环做成界面一键操作。

## What Changes

- 新增一键更新全链路:点击「立即更新」后自动下载新版本 exe(带进度反馈),下载完成后弹窗确认,确认后自动退出、替换可执行文件、重启为新版本。
- `CheckForUpdates` 增强:解析 release 资产列表,缓存 `health-tool.exe` 的下载地址,供一键更新直接使用。
- 设置页 UI 增强:`update-available` 时提供「立即更新」按钮与下载进度条;下载完成后弹确认框(原生对话框,文案含新版本号);取消后按钮变为「重启更新」,可直接跳过下载进入确认。
- 替换机制:应用退出后由一次性批处理脚本将新文件覆盖为正式 exe 并重启应用;不保留旧版本备份。
- 平台与版本约束:非 Windows 平台与 dev 版本不提供一键更新;exe 目录不可写时降级为跳转 GitHub 手动更新。
- 失败兜底:替换失败时启动旧版本并保留新文件;下载中断的 `.part` 残渣在启动时清理。

## Capabilities

### New Capabilities

- `one-click-auto-update`:一键更新全链路能力——资产下载与进度、目录可写探测与降级、确认重启弹窗、退出后替换并重启、失败兜底与残渣清理、平台 stub 与 dev 短路。

### Modified Capabilities

- `version-update-check`:手动检查更新需解析并返回资产下载地址(下载链接随检查结果缓存);设置页更新区在 `update-available` 时提供「立即更新」入口与下载进度、确认重启的交互,取消后提供「重启更新」入口。

## Impact

- `updater.go`:检查逻辑增加 assets 解析;新增下载与替换编排逻辑。
- 新增 `updater_apply_windows.go`(Windows 实现)与 `updater_apply_stub.go`(非 Windows stub),沿用现有按平台双文件模式。
- `app.go`:新增绑定方法(如 `DownloadAndApplyUpdate`),复用现有 Quit 链路(`requestQuit` + `runtime.Quit`);`model` 增加下载状态/进度相关结构。
- 前端 `frontend/src/views/settings.js`:更新区新增「立即更新」「重启更新」按钮、进度条与确认弹窗交互;通过 Wails events 接收下载进度。
- `.github/workflows/release.yml`:可选新增 sha256 checksum 资产(不阻塞,可延后)。
- 无新增第三方依赖;bat 脚本为构建期生成的临时文件,不纳入版本库。
