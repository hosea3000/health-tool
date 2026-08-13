## Context

当前根目录 19 个 Go 文件全部属于 `package main`。数据契约类型定义在绑定层文件里（`AppStatus`/`TimelineEntry`/`CountdownView` 在 `app.go`，`Settings` 在 `settings.go`），持久化逻辑分散在 4 个文件（`settings.go` + 3 个 `*_store.go`）。存储函数引用这些契约类型，`App` 又引用存储函数，导致类型、存储、绑定三者盘根错节在同一包内。

关键约束（来自 `AGENTS.md`）：

- Wails v2 应用，`main.go` 必须为 `package main` 且 `//go:embed all:frontend/dist` 的路径相对 main 包目录（根目录），因此根目录永远至少保留 `main.go`。
- `App` 的导出方法是前后端绑定契约，改动返回类型会影响 `frontend/wailsjs` 生成产物。
- `domain` 包已干净分离（零外部依赖），保持不变。
- 平台层（`input`/`lock`/`notification`/`tray`）中 `tray` 的 `startTray` 是 `*App` 方法（`tray.init(a *App)` 持有 App），与 `App` 存在循环耦合，本次**不触碰**。

## Goals / Non-Goals

**Goals:**

- 将数据契约类型抽取到 `model` 包，持久化逻辑抽取到 `store` 包。
- 建立单向无环依赖：`main → store → model → domain`，且 `main → domain`。
- 保持所有对外行为不变（绑定方法签名语义、持久化格式、状态机逻辑均不变）。

**Non-Goals:**

- 不改动 `domain` 状态机逻辑（`monitor.go` / `countdown.go`）。
- 不拆分 `app.go`（452 行）的绑定方法与 tick 循环。
- 不重构平台层（`input`/`lock`/`notification`/`tray`），不解决 `tray ↔ App` 循环依赖。
- 不合并 store 文件或消除其 JSON 读写样板重复（那是独立的后续优化）。

## Decisions

### D1：包边界划分

| 包 | 职责 | 内容 |
|----|------|------|
| `model` | 数据契约类型 | `AppStatus`、`Settings`、`TimelineEntry`、`CountdownView` |
| `store` | 持久化 | settings/timeline/countdown/card_order 的 load/save/path 及文件结构体类型 |
| `domain` | 领域状态机（已有） | `monitor.go`、`countdown.go`，新增 `DurationFromMinutes` |
| `main` | 入口 + 绑定 + 平台层 | `main.go`、`app.go`、平台层 10 文件 |

依赖方向：

```
main ──▶ store ──▶ model ──▶ domain
  │                        ▲
  └────────── domain ──────┘
```

**理由**：`model` 只依赖 `domain`（因 `CountdownView.Rule` 引用 `domain.Rule`、`DefaultSettings` 引用 `domain.DefaultReminderDuration`），`store` 依赖 `model` + `domain`，`main` 依赖三者。无环。

### D2：符号迁移表

| 原（package main） | 新 |
|---|---|
| `Settings` | `model.Settings` |
| `defaultSettings()` | `model.DefaultSettings()` |
| `AppStatus` | `model.AppStatus` |
| `TimelineEntry` | `model.TimelineEntry` |
| `CountdownView` | `model.CountdownView` |
| `durationFromMinutes()` | `domain.DurationFromMinutes()` |
| `loadSettings`/`saveSettings`/`userSettingsPath` | `store.LoadSettings`/`SaveSettings`/`UserSettingsPath` |
| `timelineFile`/`loadTimelineFile`/`saveTimelineFile`/`userTimelinePath` | `store.TimelineFile`/`LoadTimelineFile`/`SaveTimelineFile`/`UserTimelinePath` |
| `countdownFile`/`loadCountdownFile`/`saveCountdownFile`/`userCountdownPath` | `store.CountdownFile`/`LoadCountdownFile`/`SaveCountdownFile`/`UserCountdownPath` |
| `cardOrderFile`/`loadCardOrderFile`/`saveCardOrderFile`/`userCardOrderPath` | `store.CardOrderFile`/`LoadCardOrderFile`/`SaveCardOrderFile`/`UserCardOrderPath` |

**注意**：`timelineFile`/`countdownFile`/`cardOrderFile` 三个结构体也必须导出，因 `app.go`（如 `saveCountdownFile(path, countdownFile{...})`）与 `app_test.go` 直接构造/读取它们。

### D3：`durationFromMinutes` 归属 → `domain`

该函数被 `app.go`（3 处：构造 monitor、`SetReminderDuration`/`SetRestDuration`）与 `store`（校验 settings）共同引用，必须放在 main 与 store 都能 import 的公共底层。

**候选与取舍**：

- `domain.DurationFromMinutes()`（**采用**）：语义正确（时间单位换算），domain 是零依赖公共底层，main/store 均已 import domain。
- `model.DurationFromMinutes()`：与 `Settings` 放一起，但 `model` 承载的是契约类型，混入工具函数略杂。

### D4：平台层与 tray 保持不动

`tray_windows.go` 的 `startTray` 是 `*App` 方法，`tray.init(a *App)` 持有 App，与 `App` 循环耦合。将其抽到独立包需引入接口解耦（`interface{ ShowMainWindow(); RequestQuit() }`），改动面大且无行为收益，本次明确排除。

### D5：Wails 绑定产物重生成

绑定方法返回类型从 `main.AppStatus` 变为 `model.AppStatus` 等。落地后必须运行 `wails dev`（或 `wails build`）重新生成 `frontend/wailsjs`，并核对前端对模型类型的 import 路径（如 `frontend/src/tools/reminder/index.js`）。

## Risks / Trade-offs

- **[风险] 前端 import 路径断裂**：wailsjs 类型命名空间由 `main` 变为 `model`，前端 import 若不更新会导致构建失败 → **缓解**：落地后运行 `cd frontend && npm run build` 验证，并全局搜索 `wailsjs` 相关 import。
- **[风险] 测试文件迁移遗漏**：`settings_test.go` 测的是 store 逻辑，须随函数迁入 `store` 包并改包名；`app_test.go` 大量引用被迁移符号 → **缓解**：以 `go test ./...` 全量编译为验收标准。
- **[风险] 导出后命名冲突/拼写错误**：手动迁移易漏改引用 → **缓解**：以 `go build ./...` 编译错误驱动，逐包迁移（先 model，再 store，最后 main），每步可编译。
- **[权衡] 增加包间跳转成本**：拆包后 `Settings` 等类型需跨包 import → 接受，换取清晰的单向依赖与根目录整洁。

## Open Questions

无未决问题。`durationFromMinutes` 归属已按 D3 决定为 `domain`。
