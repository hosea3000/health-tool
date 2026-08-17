## 1. 后端：自启注册表能力

- [x] 1.1 新建 `autostart.go`：定义包内接口约定与 Run key 命令行值构造纯函数 `buildAutoStartCommand(exePath string) string`（exe 路径含空格时整体加引号，追加 ` --hidden`），并编写单测覆盖普通路径与含空格路径
- [x] 1.2 新建 `autostart_windows.go`（`//go:build windows`）：用 `golang.org/x/sys/windows/registry` 实现 `setAutoStart(enabled bool) error` 与 `autoStartEnabled() (bool, error)`，读写 HKCU `Software\Microsoft\Windows\CurrentVersion\Run` 下名为 `health-tool` 的值
- [x] 1.3 新建 `autostart_stub.go`（`//go:build !windows`）：两个函数均返回平台不支持错误

## 2. 后端：绑定方法与启动参数

- [x] 2.1 `app.go` 新增绑定方法 `SetAutoStart(enabled bool) error`（薄封装 `setAutoStart`）与 `AutoStartEnabled() (bool, error)`；`Settings` 结构体新增 `AutoStart bool` 字段（JSON 标签 `autoStart`），仅由 `GetSettings` 实时从注册表读取填充
- [x] 2.2 确认 `SaveSettings` 链路忽略 `AutoStart` 字段（`settings.json` 不持久化），`store/settings.go` 校验逻辑不因新字段受影响，补充/调整相关单测
- [x] 2.3 `main.go` 用 `flag` 包解析 `--hidden`（纯函数化便于单测），映射到 Wails 启动隐藏：优先验证 `options.App` 的 `Hidden` 字段在 Windows 的行为，不可用则 fallback 到 `OnStartup` 中 `runtime.WindowHide`
- [x] 2.4 运行 `go test ./...` 通过

## 3. 前端：设置页「通用」section

- [x] 3.1 `frontend/src/views/settings.js` 在「关于」下新增「通用」section，放置开机自启 toggle；用 `GetSettings().autoStart` 初始化开关状态
- [x] 3.2 toggle 变更调用 `SetAutoStart`：成功后更新开关，失败时回滚开关并展示错误提示；状态读取失败或非 Wails bridge 环境下隐藏/禁用开关
- [x] 3.3 `app.css` 补充 toggle 与提示样式，与现有设置页视觉一致

## 4. 联调与验证

- [x] 4.1 运行 `wails dev` 重新生成 `frontend/wailsjs` 绑定（`SetAutoStart`、`Settings.autoStart`），确认前端 import 正常
- [x] 4.2 `cd frontend && npm run build` 通过
- [x] 4.3 人工验证清单（Windows）：开关写入/删除 Run key 值正确（regedit 查看）；重启系统后应用静默进托盘、无主窗口；自启实例运行时手动双击 exe 唤起窗口；手动启动（无参数）正常显示窗口；任务管理器「启动应用」中出现 health-tool 条目
