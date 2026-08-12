## Context

`Timeline()` 当前返回 `App` 内存中的 `[]TimelineEntry`（`app.go`），每条记录含 `kind`、`startedAt`、可空 `endedAt` 与按查询时刻计算的 `durationSeconds`；记录仅在本次运行周期有效，应用退出即丢失。`settings.go` 已确立本地持久化模式：`os.UserConfigDir()/health-tool/` 目录 0700、JSON 文件 0600，且无文件时安全返回默认值。

## Goals

- 当天记录跨应用重启保留，`Timeline()` 与累计工作时长不因重启归零。
- 新的一天（重启跨天或运行跨午夜）从空记录开始。
- 持久化对 `TimelineEntry` 前端契约零改动。

## Non-Goals

- 不支持多天历史查询、导出或统计分析。
- 不持久化当前进行中工作段的秒级进度（时长始终按查询时刻计算，落盘只存区间起止）。
- 不记录任何输入内容、坐标或窗口信息，沿用隐私边界。

## Decisions

- **单文件 `timeline.json` + `date` 字段。** 文件形如 `{date, savedAt, entries[]}`；日期与当天不符即整文件作废。相比按天分文件，免去旧文件堆积与清理策略，贴合"只要当天"。
- **复用设置存储目录与权限。** 位于 `~/.config/health-tool/timeline.json`，沿用 0700 目录、0600 文件，与 `settings.go` 一致的 JSON 读写封装。
- **仅在状态转换时写盘。** 在 `recordTimelineTransitionLocked` 有实际变化后重写文件；不参与每秒轮询，避免无谓 IO。
- **优雅退出闭合开放记录。** `shutdown()` 时以当前时刻闭合进行中的记录并写盘；崩溃恢复时用 `savedAt` 闭合未闭合记录，最多损失最后一次写入后的尾巴。
- **跨午夜在 tick 中检测。** 每秒轮询时比较本地日期，日期变化即清空内存记录并写空文件。
- **持久化路径可注入。** `App` 新增 `timelinePath` 字段，`newApp`/`newAppWithSettings` 中为空（内存模式，兼容现有测试）；测试与真实路径通过构造函数或字段注入，参照 `settingsPath` 现有模式。
- **前端文案同步。** section note 由"仅保留本次运行记录"改为"仅保留当天记录"，空状态文案保持现有语义。

## Risks and Open Questions

- [崩溃丢尾] 强杀进程时最后一段未闭合记录以 `savedAt` 闭合，可能少计最后数分钟；优雅退出不受影响。可接受。
- [写盘失败] 目录不可写时静默跳过持久化，不影响内存行为与提醒功能（与设置保存失败策略一致）。
- [旧文件兼容] 首次升级时旧版本无 `timeline.json`，按无文件处理，行为安全。
