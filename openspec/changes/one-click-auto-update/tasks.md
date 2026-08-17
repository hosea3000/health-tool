## 1. 模型与检查结果扩展

- [x] 1.1 `model.UpdateCheckResult` 增加 `DownloadURL` 字段；新增 `UpdateDownloadEvent` 结构（阶段：下载中/完成/错误，已下载字节、总字节、百分比、消息）
- [x] 1.2 `updater.go` 的 `githubRelease` 扩展 `assets` 字段与资产结构（`name`、`browser_download_url`）；`checkForUpdates` 按文件名精确匹配 `health-tool.exe` 并填充 `DownloadURL`
- [x] 1.3 更新 `updater_test.go`：资产匹配（存在/缺失/多资产）、`DownloadURL` 填充与空值场景

## 2. 下载与替换（Windows 实现）

- [x] 2.1 新增 `updater_apply_windows.go`：实现 exe 目录可写探测（尝试创建临时文件）、`health-tool.exe.part` / `.new` 路径常量与清理函数
- [x] 2.2 实现下载：请求 `DownloadURL`，流式写入 exe 同目录 `.part`，通过 `runtime.EventsEmit("update:progress", …)` 推送进度，完成后 rename 为 `.new` 并推送完成事件；失败清理 `.part` 并推送错误事件；下载前删除残留 `.part`
- [x] 2.3 实现一次性批处理脚本生成：等待进程消失（`tasklist` 轮询）→ `move /y .new → exe`（失败重试至多 3 次）→ 启动 exe（失败则启动旧版并保留 `.new`）→ 删除自身；路径用 `%~dp0` 引用且全部加引号、处理空格/`&`/`!` 转义
- [x] 2.4 实现 `ApplyUpdateAndRestart`：校验 `.new` 存在 → 写 bat → `cmd /c start /b` 启动脚本 → 置位退出请求并调用 `runtime.Quit`（复用现有 shutdown 持久化链路）
- [x] 2.5 实现启动清理：应用启动时删除 exe 同目录的 `.part` 残渣，保留 `.new`

## 3. 非 Windows stub

- [x] 3.1 新增 `updater_apply_stub.go`：下载、可写探测、重启等方法返回"当前平台不支持自动更新"提示；启动清理为空操作

## 4. App 集成

- [x] 4.1 `app.go` 新增绑定方法 `DownloadAndApplyUpdate()`（goroutine 异步执行：dev 短路 → 平台检查 → 可写探测 → 下载 → 事件推送）与 `ApplyUpdateAndRestart()`；dev 版本与非 Windows 平台不发起网络请求
- [x] 4.2 `App.startup` 调用 `.part` 残渣清理；`CheckForUpdates` 返回结果携带 `DownloadURL`（dev 短路行为不变）
- [x] 4.3 运行 `wails build` 或 `wails dev` 重新生成 `frontend/wailsjs` 绑定（新增两个绑定方法）

## 5. 前端设置页

- [x] 5.1 `settings.js` 更新区：`update-available` 且 `downloadUrl` 非空时显示「立即更新」按钮（空则仅保留「前往 GitHub 查看」）；点击调用 `DownloadAndApplyUpdate` 并进入下载态（禁用按钮防重复点击）
- [x] 5.2 订阅 `update:progress` 事件渲染进度条（阶段文案：下载中百分比 / 下载完成）
- [x] 5.3 下载完成：调用前端 runtime `MessageDialog` 弹出原生确认框（文案含新版本号）；确认 → 调用 `ApplyUpdateAndRestart`；取消 → 按钮变为「重启更新」
- [x] 5.4 「重启更新」点击 → 直接弹确认框，确认后调用 `ApplyUpdateAndRestart`（不重新下载）
- [x] 5.5 错误事件 → 反馈区显示失败文案，按钮恢复「立即更新」；无 Wails 桥接的开发预览环境不显示更新按钮

## 6. 测试与构建验证

- [x] 6.1 `updater_test.go` 补充 bat 生成测试：内容包含等待/move/start 关键片段，路径转义（空格、`&`、`!`）与 `%~dp0` 引用正确
- [x] 6.2 `app_test.go` 补充：dev 版本 `DownloadAndApplyUpdate` 短路不发起网络请求；stub 平台返回不支持提示
- [x] 6.3 验证 `go test ./...` 全绿且 `cd frontend && npm run build` 成功

## 7. 待更新状态持久化（评审后追加）

- [x] 7.1 `model` 新增 `PendingUpdateInfo`（exists + version）；`exeUpdatePaths` 增加 `.new.version` 路径
- [x] 7.2 下载完成写入 `.new.version`；确认框版本一律从 `pendingUpdateVersion` 读取（消除内存依赖）
- [x] 7.3 新增 `PendingUpdateInfo()` 绑定方法，设置页加载时查询并渲染「重启更新」入口
- [x] 7.4 `CheckForUpdates` 返回 `up-to-date` 时清理 `.new` 与 `.new.version`；检查失败保留
- [x] 7.5 移除 `DownloadAndApplyUpdate` 的 `.new` 短路；bat 在 move 成功后删除 `.new.version`；启动清理孤儿 `.new.version`
- [x] 7.6 测试：版本元数据读写、清理函数、`PendingUpdateInfo`、up-to-date 清理；前端构建与 Windows 交叉编译通过
