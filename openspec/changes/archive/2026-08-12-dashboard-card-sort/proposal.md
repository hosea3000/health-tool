## Why

dashboard 卡片目前只能按工具注册顺序固定排列，用户无法调整自己喜欢的布局。作为个人桌面工具，卡片顺序是用户配置的一部分，需要支持跨工具自由拖拽排序，并在下次启动保持；新增卡片（如新建倒数日事件）应出现在末尾。

## What Changes

- dashboard 卡片网格**扁平化**：卡片直接作为网格子元素，跨工具任意拖拽排序（弃用按工具分组的 `card-group` 结构）。
- 工具契约新增卡片身份：`renderCards()` 产出的每张卡片带 `data-card` 稳定 key（如 `reminder`、`countdown:<eventId>`、占位卡 `countdown`）。
- **顺序持久化到 Go 后端**：新增 `card_order.json`（与 settings/countdowns 同级），提供 `GetCardOrder` / `SaveCardOrder` 方法。
- 卡片渲染**按存储顺序对账**：已有 key 原位刷新、新 key 追加末尾、消失 key 移除、失效 key 加载时丢弃。
- 拖拽交互：HTML5 原生 drag & drop，整卡可拖（点击不动仍进详情页）；drop 后立即持久化新顺序。
- 拖拽进行中暂停对应卡片刷新，避免节点替换打断拖动。
- 空工具占位卡参与排序（key = 工具 id）。

## Capabilities

### New Capabilities
- 无

### Modified Capabilities
- `tool-dashboard`: 新增"卡片拖动排序与顺序持久化"需求（跨工具扁平排序、持久化、新增卡片在末尾）

## Impact

- `frontend/src/views/dashboard.js`：扁平化渲染 + 顺序对账 + 拖拽事件
- `frontend/src/tools/reminder/index.js`、`frontend/src/tools/countdown/index.js`：卡片补 `data-card` key
- `frontend/src/app.css`：拖拽态样式（cursor、dragging、drop 指示）
- `card_order_store.go`（新）：`card_order.json` 读写（照抄 timeline_store 模式）
- `app.go`：追加 `GetCardOrder` / `SaveCardOrder`；启动时加载顺序
- `wailsjs` 绑定重新生成
- 无新第三方依赖
