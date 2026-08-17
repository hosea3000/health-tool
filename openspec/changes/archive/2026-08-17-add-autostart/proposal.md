## Why

健康工具必须在运行中才有价值：久坐提醒依赖持续的输入监控，托盘驻留设计也以"一直在跑"为前提。但应用目前只能手动启动，重启电脑后往往忘记打开，久坐提醒整天失效。开机自启动是补上这个缺口的关键一步。

## What Changes

- 新增 Windows 开机自启动开关：写入 / 删除 HKCU `Software\Microsoft\Windows\CurrentVersion\Run` 注册表键（用户级权限，无需管理员）。
- 新增 `--hidden` 启动参数：自启时以隐藏窗口方式启动，直接进入托盘驻留，不弹窗打扰。
- 设置页新增「通用」section，放置开机自启 toggle；这是设置页第一个全局设置项。
- 新增 App 绑定方法 `SetAutoStart(bool)`（注册表写入可能失败，独立方法以便把错误反馈给 UI）；`GetSettings` 返回值附带 `autoStart` 实际状态供 UI 初始化。
- 非 Windows 平台提供 stub（返回不支持），沿用 `tray_*` / `input_*` 的平台拆分模式。

## Capabilities

### New Capabilities
- `autostart`: 开机自启动能力——Run key 的读写与状态查询、`--hidden` 静默启动、设置页开关交互与状态展示。

### Modified Capabilities
<!-- 无：本变更不改动现有 spec 的 requirement。settings 相关行为属于新 capability 的 UI 面。 -->

## Impact

- **后端**：新增 `autostart_windows.go` / `autostart_stub.go`（平台拆分）；`main.go` 解析 `--hidden` 并映射到 Wails 启动选项；`app.go` 新增绑定方法、`GetSettings` 返回扩容。
- **前端契约**：`App` 绑定新增 `SetAutoStart`，`Settings` 结构体新增字段——`frontend/wailsjs` 需要重新生成（`wails dev` 或 `wails build` 时自动完成）。
- **设置页**：`frontend/src/views/settings.js` 新增「通用」section 与 toggle 交互。
- **已知边界（明确不做）**：
  - 不感知任务管理器「启动应用」中用户对 `StartupApproved` 的禁用操作，UI 可能显示与实际不符；spec 中记录为已接受边界。
  - 不做 exe 路径变更后的自愈重写（现有自动更新用 `move /y` 原地覆盖，路径不变，自启不受影响）。
  - 托盘菜单不放自启开关（保持三项菜单不变）。
- **测试**：CI 在 ubuntu 构建，注册表逻辑无法集成测试；可测部分为纯函数（如 Run key 命令行值的构造，含路径带空格的引号处理）。
