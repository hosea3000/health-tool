## Why

`health-tool` 目前的发布产物只有裸 exe，用户拿到手需要自行放置、无可卸载入口、也不会自动补齐 WebView2 运行时；同时内置的一键更新检查/下载在国内网络访问 GitHub 时经常失败，缺少加速通道。本次改动把发布形态升级为「installer + 裸 exe 双产物」，并为一键更新补上可配置的代理加速，使安装、更新两条链路在国内网络下都能闭环。

## What Changes

- 发布产物新增 NSIS installer（`health-tool-amd64-installer.exe`），与裸 `health-tool.exe` 一并挂到 GitHub Release。
- CI 流水线新增 NSIS 安装步骤，`wails build` 加 `-nsis -installscope user`，安装到 `%LOCALAPPDATA%\Programs\health-tool\`（可写目录，保证一键更新能覆盖 exe）。
- installer 中文化（`MUI_LANGUAGE SimpChinese`），并在卸载前检测应用是否仍在运行（按托盘窗口类名 `HealthToolTrayClass`），运行中拒绝卸载并提示先退出。
- installer/卸载流程不触碰用户数据目录（`os.UserConfigDir()/health-tool/`），保证升级与重装后设置、打卡记录、倒数日不丢失。
- 一键更新继续下载裸 exe 覆盖二进制（不下载 installer），与 installer 首次安装形态共存。
- 一键更新新增「代理加速」：设置持久化 `UpdateProxy` 字段，检查更新与下载资产均可经代理前缀（gh-proxy 系列）请求，代理不可用时检查更新回退直连。
- 设置页「更新」段下方新增「网络加速」小段，提供代理地址输入与保存。
- **BREAKING**：`App.SaveSettings` 参数由三个位置参数（`reminderMinutes, restMinutes, notificationsEnabled`）改为单个 `model.Settings` 结构体，返回值保持 `bool`。前端 `reminder/index.js` 调用点与 wailsjs 生成绑定需同步更新。

## Capabilities

### New Capabilities
- `packaging-installer`: 定义 Windows installer 的发布形态、安装目录、中文化、卸载进程检测、用户数据保留，以及一键更新与 installer 的共存关系。

### Modified Capabilities
- `ci-autorelease`: 发布产物完整性从「仅 exe」扩展为「exe + installer」，构建流程新增 NSIS 安装与 `-nsis -installscope user` 参数。
- `version-update-check`: 新增「更新代理加速」requirement（代理配置持久化、检查更新经代理并回退直连），并修改设置持久化结构以容纳 `UpdateProxy`。
- `one-click-auto-update`: 下载资产时应用代理前缀；明确与 installer 共存时仍下载裸 exe。

## Impact

- **CI**：`.github/workflows/release.yml`（新增 NSIS 安装、构建参数、上传 installer）。
- **打包**：`build/windows/installer/project.nsi`（中文化、卸载进程检测）、`.gitignore`（忽略 `wails_tools.nsh` 与 `tmp/`）。
- **后端**：`updater.go`（新增 `proxiedURL`、`checkForUpdates` 加代理参数）、`model/model.go`（`Settings` 新增 `UpdateProxy`）、`app.go`（`SaveSettings` 签名变更、`CheckForUpdates`/`DownloadAndApplyUpdate` 读代理）。
- **前端**：`frontend/src/views/settings.js`（网络加速输入区）、`frontend/src/tools/reminder/index.js`（`SaveSettings` 结构体调用）、`frontend/wailsjs/**`（wails build 重新生成）。
- **测试**：`updater_test.go` 补代理用例，`app_test.go` 适配 `SaveSettings` 新签名。
- **发布流程**：本地开发仍需 `cd frontend && npm run build` 后再 `wails build`；installer 仅在 Windows 目标构建下产出。
