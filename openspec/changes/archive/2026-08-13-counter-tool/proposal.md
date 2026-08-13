## Why

用户在日常生活中需要记录重复性事件（喝了几杯水、上了几次厕所）。倒数日验证了"多实例工具"扩展位后，计数器是第二个纯数据工具：dashboard 卡片上加号一键 +1，按周期（天/月/年/永不清零）自动清零重计，可选目标值，并保留各周期历史用于回顾。

## What Changes

- 新增 `counter` 工具模块，登记进前端工具注册表（`tools/index.js`），不改动平台本体。
- 新增计数器数据模型：名字 + 重置周期 + 可选目标值 + 按周期的次数桶 map。
- 重置周期四种：**天 / 月 / 年 / 永不清零**；次数按"今天日期 + 周期"算出的桶 key 存取，无 tick 滚动逻辑，换周期自然开新桶。
- dashboard 卡片新增 **+** 按钮：点击一次当前周期次数 +1，不触发进详情、不干扰拖拽排序；卡片展示当前周期大数字与状态（设目标显示"还差 N / 已达成"，未设只显示次数）。
- 详情页：计数器列表（增删改、重命名、改周期/目标），每条支持 **− / 可编辑数字 / +** 直接修改计数，并提供"最近 N 个周期"的历史列表。
- 后端新增：`domain/counter.go`（模型 + 校验 + 周期桶 key）、`store/counter.go`（`counters.json`）、`App` 方法（列表 / 新增 / 修改 / 删除 / 增减 / 设置次数）。
- 浏览器预览模式（无 Wails 桥）下显示空状态，不造假数据。

## Capabilities

### New Capabilities
- `counter`: 计数器实例的配置（名字、重置周期、可选目标值）、按周期计数、dashboard 卡片 + 按钮与详情页增删改及计数编辑

### Modified Capabilities
- 无。平台（`tool-dashboard`）不需要改动：注册表、卡片网格、详情页路由、卡片契约均为通用机制。

## Impact

- `frontend/src/tools/counter/index.js`（新）：工具模块，`refreshInterval: 60000`
- `frontend/src/tools/index.js`：注册 counter
- `domain/counter.go`（新）：`Counter` 类型、`Validate`、周期桶 key 计算
- `domain/counter_test.go`（新）：周期桶 key 与校验单测
- `store/counter.go`（新）：照抄 `countdown.go` 的 IO 模式
- `model/model.go`：新增 `CounterView`
- `app.go`：追加 `Counters` / `AddCounter` / `UpdateCounter` / `DeleteCounter` / `IncrementCounter` / `DecrementCounter` / `SetCounterCount`
- `frontend/src/app.css`：+ 按钮与历史列表样式
- 不新增第三方依赖；Go 端继续使用肥 App 结构，不引入工具接口。
