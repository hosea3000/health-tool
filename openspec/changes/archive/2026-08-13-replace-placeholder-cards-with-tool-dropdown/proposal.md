## Why

当前 dashboard 为每个空工具渲染一张"暂无内容"占位卡，把「工具存在」（导航入口）和「工具无数据」（空态）两件事绑在了一起：占位卡的空态文案与工具详情页已有的空态重复；占位卡的 key 用工具 id（如 `countdown`），一旦工具产出真实卡片（key 变为 `countdown:<id>`），`card_order.json` 里残留的占位 key 就变成静默失效的孤儿 key。需要把"进入工具"这一职责从卡片剥离到顶栏导航，让卡片回归纯状态组件。

## What Changes

- 顶栏新增"工具"下拉列表，列出注册表**全部工具**（固定导航枢纽，不随空态变化），点击任意项进入对应工具详情页；键盘可达并带 `aria-expanded`。
- dashboard 卡片网格只渲染**有数据的工具卡片**；空工具不再渲染占位卡，也不再参与网格与排序。
- 删除 `makePlaceholder` 及 `tool-card-empty` 相关样式。
- **BREAKING**：占位卡不再存在，`card_order.json` 中残留的占位 key（等于工具 id）在加载时按失效 key 丢弃。
- 下拉仅在 dashboard 顶栏提供；详情页顶栏保持不变（保留「‹ 返回」按钮），不带状态高亮。

## Capabilities

### New Capabilities

- 无

### Modified Capabilities

- `tool-dashboard`: 移除空工具占位卡，新增顶栏"工具"下拉导航（固定列表、进入详情、键盘可达）。

## Impact

- `frontend/src/views/dashboard.js`：移除 `makePlaceholder` 与空工具占位逻辑，新增顶栏下拉渲染与交互。
- `frontend/src/app.css`：新增下拉样式，移除 `tool-card-empty` 相关样式。
- 纯前端改动：注册表 `tools/index.js` 无需变更，无 Go 绑定变更，无需重新生成 `wailsjs`。
- 无新第三方依赖。
