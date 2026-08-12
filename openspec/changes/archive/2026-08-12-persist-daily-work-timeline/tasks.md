## 1. 持久化文件读写

- [x] 1.1 定义持久化文件结构（`date`、`savedAt`、`entries`），新增 `loadTimeline(path)` 与 `saveTimeline(path, ...)`，沿用 `settings.go` 的目录与权限约定。
- [x] 1.2 在 `App` 新增 `timelinePath` 字段，`newApp`/`newAppWithSettings` 中为空（内存模式）。

## 2. 生命周期集成

- [x] 2.1 启动时若 `timelinePath` 非空则加载当天记录，过期文件忽略；未闭合记录以 `savedAt` 闭合。
- [x] 2.2 状态转换时（`recordTimelineTransitionLocked` 有变化）重写持久化文件。
- [x] 2.3 `shutdown()` 时以当前时刻闭合进行中的开放记录并写盘。
- [x] 2.4 tick 中检测跨午夜，日期变化时清空记录并写空文件。

## 3. 验证

- [x] 3.1 新增 Go 测试：当天重启恢复记录、跨天忽略旧文件、未闭合记录以 `savedAt` 闭合、状态转换落盘。
- [x] 3.2 前端文案由"仅保留本次运行记录"改为"仅保留当天记录"。
- [x] 3.3 运行 `cd frontend && npm run build` 与 `go test ./...` 均通过。
