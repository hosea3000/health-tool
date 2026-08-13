# Agent Instructions

## 语言

- 使用中文沟通、编写说明和注释。

## 项目结构

- 这是一个 Wails v2 桌面应用：Go 后端入口是 `main.go`，业务协调在 `app.go`，领域状态机在 `domain/monitor.go`、倒数日规则在 `domain/countdown.go`。
- 前端是无框架 Vite 应用，源代码在 `frontend/src`；`main.js` 通过 `views/dashboard.js` / `views/detail.js` 切换视图，每个工具是 `frontend/src/tools/<tool>/index.js` 导出的模块，在 `frontend/src/tools/index.js` 的 `registry` 里注册。
- `input_*`、`tray_*`、`notification_*`、`lock_*` 按平台提供 Windows 实现（`//go:build windows`）和非 Windows stub（`//go:build !windows`）；修改输入监听或托盘行为时同时检查两类文件。托盘是原生 Win32 API 实现，无 systray 依赖。
- `CONTEXT.md` 是领域词汇表；状态名称和“有效活动”“工作段”“提醒休息期”等术语应与其保持一致。
- 业务按 spec 驱动开发：需求/设计归档在 `openspec/`，流程为 propose → apply → archive（见 `.opencode/commands/opsx-*.md` 与 openspec skills）。

## 命令

- 安装前端依赖：`cd frontend && npm install`。
- 前端构建：`cd frontend && npm run build`。
- Go 测试：`go test ./...`；单包测试可用 `go test .` 或 `go test ./domain`，单测试用 `-run TestName`。
- 开发运行：`wails dev`。
- 发布构建：先 `cd frontend && npm run build`，再 `wails build`。
- 无 lint / typecheck / Makefile 配置；`go test ./...` 与前端构建是唯一校验手段。CI 无测试步骤，发布见 `.github/workflows/release.yml`（打 `v*` tag 触发，Linux 交叉编译 `windows/amd64` 产物并 `gh release create`）。

## 关键约束

- 发布产物是 Windows exe（CI 在 ubuntu 上 `wails build -platform windows/amd64`）；本地 `wails build` 默认构建当前平台。开发在 Linux/macOS 上运行时，输入/托盘/通知/锁屏均为 stub。
- `main.go` 使用 `//go:embed all:frontend/dist`；`frontend/dist` 被 `.gitignore` 忽略，修改前端后必须重新构建前端才能进行 Go/Wails 构建。
- 不直接编辑 `frontend/dist` 或 `frontend/wailsjs`；两者都是构建生成产物。`frontend/wailsjs/go/main/App.js` 由 Wails 根据 `App` 上导出的方法自动生成，改动 `App` 的方法签名或 `AppStatus`/`Settings` 等结构体字段后，跑一次 `wails dev` 或 `wails build` 重新生成。
- `App` 的绑定方法是前后端数据契约：`Status`、`Timeline`、`GetSettings`、`SaveSettings`、`CountdownEvents`、`AddCountdown`、`UpdateCountdown`、`DeleteCountdown`、`GetCardOrder`、`SaveCardOrder`。
- 用户数据持久化到 `os.UserConfigDir()/health-tool/`：`settings.json`、`timeline.json`、`countdowns.json`、`card_order.json`；各 store 文件（`*_store.go`）负责读写，`App` 在 `startup`/`shutdown` 及每次变更时读写。
- 关闭主窗口会隐藏窗口而不是退出应用；输入监控、托盘和提醒仍会继续运行（单实例，二次启动只唤起窗口）。
- `domain.Monitor` 的状态转换和计时规则优先于 UI 文案；UI 只展示 `Status()` 返回的状态，不要在前端重建状态机。
