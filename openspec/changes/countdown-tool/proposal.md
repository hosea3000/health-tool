## Why

多工具平台骨架（dashboard + 工具/卡片分层 + 详情页路由）已就绪并归档。倒数日是第一个"参考型、多实例"工具，用于验证平台的扩展位：一个工具产出多张卡片、详情页内配置增删改。它不涉及后台常驻行为，是纯数据展示工具，实现简单且能立即兑现平台价值。

## What Changes

- 新增 `countdown` 工具模块，登记进前端工具注册表（`tools/index.js`），不改动平台本体。
- 新增事件数据模型：标题 + 到期规则，四种规则形态——**一次性日期 / 每月几号 / 每周周几 / 大小周周几**。
- 后端新增：`domain/countdown.go` 规则引擎（下次到期日计算 + 校验）、`countdown_store.go` 持久化、`App` 四个方法（列表/新增/修改/删除）。
- 前端新增 `tools/countdown/index.js`：`renderCards()` 每个事件产出一张 dashboard 卡片，`renderDetail()` 提供事件列表与新增/编辑/删除。
- 卡片展示剩余天数：未来显示"还剩 N 天"，当天显示"今天"，已过的一次性日期显示"已经 N 天"。
- 事件排序：按剩余天数升序（最紧急在前）；已过事件排在未来事件之后，按最近优先。
- 标题限 20 字；删除需二次确认。
- 存储为独立 `countdowns.json`，事件长期存在（无按天滚动）。
- 浏览器预览模式（无 Wails 桥）下显示空状态，不造假数据。

## Capabilities

### New Capabilities
- `countdown`: 倒数日事件的配置（四种到期规则）、持久化、dashboard 卡片展示与详情页增删改

### Modified Capabilities
- 无。平台（`tool-dashboard`）不需要改动：注册表、卡片网格、详情页路由、卡片契约均为通用机制。

## Impact

- `frontend/src/tools/countdown/index.js`（新）：工具模块，`refreshInterval: 60000`
- `frontend/src/tools/index.js`：注册 countdown
- `domain/countdown.go`（新）：`Rule` 类型、`NextOccurrence`、`Validate`
- `domain/countdown_test.go`（新）：规则引擎单测
- `countdown_store.go`（新）：照抄 `timeline_store.go` 的 IO 模式
- `app.go`：追加 `CountdownEvents` / `AddCountdown` / `UpdateCountdown` / `DeleteCountdown`
- 不新增第三方依赖；Go 端继续使用肥 App 结构，不引入工具接口。
