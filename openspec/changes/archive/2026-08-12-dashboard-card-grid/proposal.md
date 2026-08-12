## Why

当前 dashboard 卡片网格的"格"是工具槽而非卡片：一个工具的多张卡片被锁死在它自己的窄列里竖向堆叠（倒数日多个事件即如此），不够"表格化"。需要把所有工具的卡片扁平混排，以统一的 4 列等宽表格呈现，不按工具分组。

## What Changes

- dashboard 卡片网格扁平化：工具槽改为 `display: contents`，每张卡片直接成为网格单元，所有工具的卡片混排、自动换行。
- 网格布局改为 **4 列等宽**，窄窗口响应式降列（2 列 / 1 列）。
- 卡片出场顺序 = 工具注册顺序（久坐提醒在前，倒数日事件随后）；不做按内容排序。
- 空工具占位卡保持现状（虚线、可点击进入详情页），混排进表格。
- 保留每工具独立刷新间隔、DOM 位置稳定、点击进详情页路由等既有行为。

## Capabilities

### New Capabilities
- 无

### Modified Capabilities
- `tool-dashboard`: 新增"卡片网格扁平排列"需求（4 列等宽表格、按注册顺序混排、窄窗口降列）

## Impact

- `frontend/src/views/dashboard.js`：工具槽结构改为 `display: contents` 包装
- `frontend/src/app.css`：`.card-grid` 改为 4 列等宽 + 响应式降列；调整 `.card-slot` 相关样式
- 无后端改动、无数据迁移、无新依赖
