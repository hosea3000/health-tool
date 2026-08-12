## Context

`work-rest-timeline` 能力已提供 `Timeline()` 查询：返回按 `startedAt` 升序的记录，每条记录带 `kind`、`startedAt`、可空 `endedAt` 和 `durationSeconds`（进行中的记录按查询时刻计算）。前端 `refreshStatus()` 每秒并发拉取 `Status()` 与 `Timeline()`，`renderTimeline(entries)` 已持有全量记录。

## Goals

- 在时间线区域展示累计工作时长，口径仅统计 `working` 类型记录。
- 汇总在纯前端完成，不引入后端改动或数据契约变化。
- 展示实时更新，无记录时给出明确的全零展示。

## Non-Goals

- 不在后端增加汇总字段或新的 Wails 方法。
- 不改变现有 `formatDuration` 的时间线条目格式。
- 不统计 `idle-paused`、`resting` 或其他类型记录。
- 不做持久化或跨运行周期汇总。

## Decisions

- **由前端汇总工作段。** `Timeline()` 返回的 `durationSeconds` 已足够计算总和，前端对 `kind === 'working'` 的记录求和即可。相比后端新增字段或方法，避免改动 `AppStatus`/`TimelineEntry` 契约、`wailsjs/go/models.ts` 与 Go 测试，符合现有"前端只消费后端只读快照"的边界。
- **展示在时间线标题行。** 在 `工作与休息记录` 标题旁新增"工作时长"文本，语义上归属于记录区域，与 hero 面板的当前工作段计时区分开。
- **格式恒为三位。** 使用 `xxx小时xx分xx秒` 全量格式（含前导零），如 `0小时23分45秒`、`1小时05分09秒`；新增独立的格式化函数，不动现有时间线条目的 `formatDuration`。
- **无记录时显示全零。** 空时间线不额外改变空状态文案，累计工作时长显示 `0小时0分0秒`，避免误导。
- **沿用每秒刷新。** 累计时长随现有 `refreshStatus()` 刷新自然更新，无需新定时器。

## Risks and Open Questions

- [累计时长精确到秒的截断误差] 现有 `durationSeconds` 按秒取整求和，跨多个工作段时最多累计零点几秒的截断；对展示型汇总可接受，无需后端改为毫秒精度。
- 浏览器预览模式（无 Wails 桥）时间线返回空数组，累计工作时长显示 `0小时0分0秒`，行为一致。
