## 1. Domain：新增工具函数

- [x] 1.1 在 `domain` 包新增 `DurationFromMinutes(minutes int) time.Duration`（等价于 `time.Duration(minutes) * time.Minute`），并删除根目录 `settings.go` 中的旧 `durationFromMinutes`

## 2. Model 包：数据契约类型

- [x] 2.1 新建 `model/model.go`，迁入 `AppStatus`、`TimelineEntry`、`CountdownView`（自 `app.go`）与 `Settings`、`DefaultSettings`（自 `settings.go`），package 为 `model`，将 `DefaultSettings` 内对 `domain.Default*Duration` 的引用改为 `model` 包内 import

## 3. Store 包：持久化逻辑

- [x] 3.1 新建 `store/settings.go`，迁入 `LoadSettings`、`SaveSettings`、`UserSettingsPath`，内部校验改用 `domain.DurationFromMinutes`
- [x] 3.2 新建 `store/timeline.go`，迁入 `TimelineFile`、`LoadTimelineFile`、`SaveTimelineFile`、`UserTimelinePath`
- [x] 3.3 新建 `store/countdown.go`，迁入 `CountdownFile`、`LoadCountdownFile`、`SaveCountdownFile`、`UserCountdownPath`
- [x] 3.4 新建 `store/card_order.go`，迁入 `CardOrderFile`、`LoadCardOrderFile`、`SaveCardOrderFile`、`UserCardOrderPath`
- [x] 3.5 删除根目录 `timeline_store.go`、`countdown_store.go`、`card_order_store.go`、`settings.go`

## 4. Main 包：更新引用

- [x] 4.1 更新 `app.go` 的 import 与引用：`Settings`/`AppStatus`/`TimelineEntry`/`CountdownView` → `model.*`，存储函数与文件类型 → `store.*`，`durationFromMinutes` → `domain.DurationFromMinutes`

## 5. 测试迁移

- [x] 5.1 将 `settings_test.go` 迁入 `store/` 包（包名改为 `store`，更新 import 与类型引用）
- [x] 5.2 更新 `app_test.go` 中被迁移符号的引用（`Settings`、`timelineFile`、`loadTimelineFile`、`saveTimelineFile` 等）
- [x] 5.3 确认 `main_test.go`、`input_test.go` 无被迁移符号引用，如有则一并更新

## 6. 验证

- [x] 6.1 运行 `go build ./...` 确认无编译错误、无 import cycle
- [x] 6.2 运行 `go test ./...` 确认全部测试通过
- [x] 6.3 运行 `wails dev`（或 `wails build`）重新生成 `frontend/wailsjs`
- [x] 6.4 全局搜索 `wailsjs` 相关 import，核对前端模型类型命名空间由 `main` 变为 `model` 后的引用路径，并运行 `cd frontend && npm run build` 确认前端构建通过
