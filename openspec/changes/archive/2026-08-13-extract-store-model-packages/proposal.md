## Why

根目录 19 个 Go 文件全部属于 `package main`，存储层（4 个 `*_store.go`/`settings.go`）与数据契约类型（`Settings`、`TimelineEntry`、`AppStatus`、`CountdownView`）与 `App` 绑定层纠缠在同一包内，目录既平又乱，难以看出分层。将存储层与数据契约类型抽取为独立包，建立单向无环的依赖方向，是纯代码整理，不改变任何行为。

## What Changes

- 新建 `model` 包，承载数据契约类型：`AppStatus`、`Settings`、`TimelineEntry`、`CountdownView`（原定义在 `app.go` / `settings.go`）。
- 新建 `store` 包，承载持久化逻辑：`settings`、`timeline`、`countdown`、`card_order` 的 load/save/path 函数及其文件结构体类型（原 `timeline_store.go`、`countdown_store.go`、`card_order_store.go`、`settings.go`）。
- 删除根目录 `settings.go`、`settings_test.go`、`timeline_store.go`、`countdown_store.go`、`card_order_store.go`。
- 修改 `app.go`、`app_test.go`：更新 import 与类型/函数引用为新的包限定名。
- `durationFromMinutes` 工具函数迁移至 `domain` 包（被 `app.go` 与 `store` 共同引用，需放在两者均可 import 的公共底层）。
- 不触碰平台层（`input`/`lock`/`notification`/`tray`）、`domain` 状态机逻辑、`main.go` 入口与 embed。
- 重新生成 Wails 绑定产物（`frontend/wailsjs`），因绑定方法返回类型从 `main.*` 变为 `model.*`。

## Capabilities

### New Capabilities

无。本次为纯内部重构，不引入新的可观察能力。

### Modified Capabilities

无。现有能力（`work-rest-timeline`、`countdown`、`tool-dashboard`、`single-instance-window-activation`、`ci-autorelease`）的需求均未变化，仅实现代码的物理位置发生移动。

## Impact

- **代码文件**：新增 `model/`、`store/` 两个包；删除根目录 5 个文件；修改 `app.go`、`app_test.go`。
- **依赖方向**：建立 `main → store → model → domain` 与 `main → domain` 的单向无环依赖。
- **生成产物**：`frontend/wailsjs` 需重新生成（类型命名空间从 `main` 变为 `model`），前端对绑定类型的 import 需核对。
- **无** API 行为变化、无数据库/持久化格式变化、无新增外部依赖。
