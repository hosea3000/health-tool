## Context

平台已有多工具骨架：dashboard 卡片网格 + 详情页路由（`views/dashboard.js` / `views/detail.js`），工具注册表 `tools/index.js`。倒数日（countdown）验证了"多实例工具"的完整链路：`renderCards()` 每实例一卡、`renderDetail()` 详情页增删改、后端 `domain`（业务逻辑）+ `store`（JSON 持久化）+ 肥 App 绑定方法。

计数器与倒数日同构，但无规则引擎：数据是"名字 + 周期 + 目标 + 次数"，逻辑集中在**周期桶 key 计算**。这是纯数据展示工具，不涉及后台常驻行为。

## Goals / Non-Goals

**Goals:**
- 多实例计数器：用户可新增/删除/重命名多个计数器（如"喝水""上厕所"）。
- 按周期计数：天 / 月 / 年 / 永不清零，周期切换后当前次数从 0 重新累计，历史保留。
- dashboard 卡片 + 按钮一键 +1；详情页可精确修改计数。
- 可选目标值：设置了显示"还差 N / 已达成"，未设置无限累加。
- 详情页展示"最近 N 个周期"历史列表。

**Non-Goals:**
- 不做跨周期统计图表（折线图/柱状图）——历史数据已存，之后按需加。
- 不做提醒/通知（例如"目标达成提醒"）。
- 不做计数器分组或排序（沿用平台统一卡片拖拽排序）。

## Decisions

### 1. 周期桶 map 存取，替代 tick 滚动

计数器存 `counts map[string]int`，key 由 `periodKey(period, today)` 计算：

```
period: day   → "2026-08-13"
period: month → "2026-08"
period: year  → "2026"
period: never → "all"
```

读当前次数 = `counts[key] ?? 0`；+1 即 `counts[key]++`。应用跨天/跨月/跨年重开时，算出的 key 自然指向新桶，显示 0。**不在 `tick()` 里加任何 rollover 逻辑**（对比 timeline 的 `rolloverLocked`）。

替代方案：存"日期 + 单次数"并在 tick 里清零——需要在 App 常驻周期里维护状态，且没有历史，弃用。

### 2. 数据模型与分层

- `domain/counter.go`：`Counter{ID, Name, Period, Goal, Counts}` + `Validate()`（名字非空且 ≤20 字、周期合法、目标 ≥0）+ `PeriodKey(period, t)` + `CurrentCount(t)`。逻辑放 domain，沿用 countdown 的 `CountdownEvent.Validate()` 模式。
- `model/model.go`：`CounterView{ID, Name, Period, Goal, Count, PeriodLabel, GoalReached, History}` —— 供前端展示的只读视图，当前次数在读取时按 `now()` 实时计算。
- `store/counter.go`：`counters.json`，格式 `{savedAt, counters: []domain.Counter}`，照抄 `countdown.go` 的 Load/Save 模式。

### 3. App 绑定方法

| 方法 | 签名 | 说明 |
|------|------|------|
| `Counters` | `() []model.CounterView` | 列表，含当前周期次数与历史 |
| `AddCounter` | `(name, period, goal) string` | 返回错误消息，空串成功 |
| `UpdateCounter` | `(id, name, period, goal) string` | 更新名字/周期/目标，保留已有次数桶 |
| `DeleteCounter` | `(id) bool` | 删除 |
| `IncrementCounter` | `(id) int` | 当前周期 +1，返回新次数 |
| `DecrementCounter` | `(id) int` | 当前周期 −1（下限 0），返回新次数 |
| `SetCounterCount` | `(id, count) int` | 详情页直接设置当前周期次数（下限 0） |

所有写操作 `persistCountersLocked()` 落盘。周期桶是幂等的（按 key 读写），跨方法天然一致。

### 4. dashboard 卡片 + 按钮

卡片结构沿用 `tool-card`：kicker「计数器」、status 计数器名、大数字当前次数、meta 周期文案与目标状态。右上角新增圆形 + 按钮：

- 按钮是 `renderCards()` 里创建的真实 `<button>`，直接绑定 click 监听：`event.stopPropagation()` + 调用 `IncrementCounter(id)` → 用返回的新次数**直接更新卡片大数字节点**（不整卡重渲染，避免拖拽态丢失）。
- `stopPropagation` 阻止冒泡到 grid 的 `[data-tool]` 委托（否则会进详情页）；按钮是独立 button，Enter/Space 原生触发其 click 且不命中 grid keydown 的 `data-tool` 判断。
- 卡片本身 `draggable`，按钮在卡片内；点击不会触发拖拽。
- dashboard 按 `refreshInterval` 轮询重建卡片，与后端次数保持一致（跨周期清零也由此刷新，60s 内可见）。

### 5. 详情页

- 新增/编辑表单复用 `settings-backdrop` 面板：名字（≤20 字）、周期下拉（天/月/年/永不清零）、目标值数字输入（可留空 = 0 表示不设）。
- 每条计数器：`[−] 可编辑数字 [+]` 三个操作 + `编辑`/`删除`。数字可点开直接输入（`input type=number`，失焦/回车提交 `SetCounterCount`）。
- 历史列表：取最近 7 个非零周期桶，按时间倒序展示"周期标签 + 次数"（如 `08-11 · 4 次`）。周期标签复用 `PeriodKey` 反解或按周期格式化日期。

### 6. 文案与语义

- meta 周期文案：天→「今日」、月→「本月」、年→「今年」、永不清零→「累计」。
- 目标：`goal > 0` 且未达成 →「还差 N」；达成 →「已达成」；`goal = 0` → 只显示次数。
- 术语与 CONTEXT.md 对齐，新增"计数器 / 重置周期 / 计数桶"词条。

## Risks / Trade-offs

- **桶 map 无限增长**（天周期每天一个新桶，几字节/天）→ 数据量可忽略，不做清理；若长期运行显规模再按需裁剪。
- **卡片 + 按钮与卡片点击/拖拽冲突** → 按钮独立监听 + stopPropagation；轮询重建可能替换掉正在交互的节点，故 + 点击走直接 DOM 更新而非重渲染。
- **60s 刷新延迟**导致跨周期清零在卡片上最多晚 1 分钟可见 → 可接受；详情页操作即时刷新。
