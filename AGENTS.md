# Agent Instructions

## 语言

- 使用中文沟通、编写说明和注释。

## 项目结构

- 这是一个 Wails v2 桌面应用：Go 后端入口是 `main.go`，业务协调在 `app.go`，领域状态机在 `domain/monitor.go`。
- 前端是无框架 Vite 应用，源代码在 `frontend/src`；`frontend/src/main.js` 通过 Wails 绑定调用 `Status`、`GetSettings` 和 `SaveSettings`。
- `input_*`、`tray_*`、`notification_*`、`lock_*` 按平台提供 Windows 实现和非 Windows stub；修改输入监听或托盘行为时同时检查两类文件。
- `CONTEXT.md` 是领域词汇表；状态名称和“有效活动”“工作段”“提醒休息期”等术语应与其保持一致。

## 命令

- 安装前端依赖：`cd frontend && npm install`。
- 前端构建：`cd frontend && npm run build`。
- Go 测试：`go test ./...`；单包测试可用 `go test .` 或 `go test ./domain`，单测试用 `-run TestName`。
- 开发运行：`wails dev`。
- 发布构建：先运行 `cd frontend && npm run build`，再运行 `wails build`。

## 关键约束

- `main.go` 使用 `//go:embed all:frontend/dist`；`frontend/dist` 被 `.gitignore` 忽略，修改前端后必须重新构建前端才能进行 Go/Wails 构建。
- 不直接编辑 `frontend/dist`；前端源文件改动应在 `frontend/src` 完成。
- Go 的 `AppStatus` 字段是前后端数据契约；增删字段时同步检查 `frontend/wailsjs/go/models.ts` 及前端消费逻辑。
- 关闭主窗口会隐藏窗口而不是退出应用；输入监控、托盘和提醒仍会继续运行。
- `domain.Monitor` 的状态转换和计时规则优先于 UI 文案；UI 只展示 `Status()` 返回的状态，不要在前端重建状态机。
