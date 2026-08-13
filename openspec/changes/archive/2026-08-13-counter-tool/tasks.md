## 1. 后端领域层

- [x] 1.1 新增 `domain/counter.go`：`Counter{ID, Name, Period, Goal, Counts}` 与 `PeriodKey(period, t)`、`CurrentCount(t)`、`Validate()`
- [x] 1.2 新增 `domain/counter_test.go`：周期桶 key（天/月/年/永不清零）、校验、当前次数单测
- [x] 1.3 在 `model/model.go` 新增 `CounterView`（含当前次数、周期文案、目标状态、历史）

## 2. 后端存储与 App

- [x] 2.1 新增 `store/counter.go`：`counters.json` 的 Load/Save/路径，照抄 `store/countdown.go` 模式
- [x] 2.2 在 `app.go` 增加字段 `counters`、`countersPath` 与 `loadCountersLocked`/`persistCountersLocked`
- [x] 2.3 在 `app.go` 增加 `Counters` / `AddCounter` / `UpdateCounter` / `DeleteCounter`
- [x] 2.4 在 `app.go` 增加 `IncrementCounter` / `DecrementCounter` / `SetCounterCount`

## 3. 前端工具模块

- [x] 3.1 新增 `frontend/src/tools/counter/index.js`：`renderCards()`（每计数器一卡，含 + 按钮）、`renderDetail()`（列表 + 增删改表单 + 计数编辑 + 历史）
- [x] 3.2 在 `tools/index.js` 注册 `counterTool`
- [x] 3.3 在 `frontend/src/app.css` 增加 + 按钮、可编辑数字、历史列表样式

## 4. 验证与文档

- [x] 4.1 跑 `go test ./...` 通过
- [x] 4.2 `cd frontend && npm run build` 通过
- [x] 4.3 在 CONTEXT.md 补充"计数器 / 重置周期 / 计数桶"词条
