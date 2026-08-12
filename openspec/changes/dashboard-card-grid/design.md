## Context

dashboard 由 `views/dashboard.js` 渲染：为每个注册工具创建一个 `card-slot` div（作为网格单元格）追加进 `#card-grid`，再把工具产出的卡片塞进该 slot。网格 `repeat(auto-fill, minmax(280px, 1fr))` 的"格"是工具槽，因此单个工具的多张卡片（如倒数日事件）竖向堆在一格里。

需要改为所有工具的卡片扁平混排成统一的等宽表格。

## Goals / Non-Goals

**Goals:**
- 所有工具卡片扁平混排，不按工具分组，自动换行成表格状。
- 4 列等宽，窄窗口响应式降列。
- 顺序 = 工具注册顺序，稳定且不随刷新跳动。
- 保留：每工具独立刷新、点击进详情页、空工具占位卡。

**Non-Goals:**
- 跨工具内容排序（久坐提醒"已工作时长"与倒数日"剩余天数"语义不可比）。
- dashboard 自由拖动排序 / 配置化（后续单独做，本次只把顺序稳定下来）。
- 空占位卡的交互方案演进（本次保持现状）。

## Decisions

### D1. 工具槽改 `display: contents`
每个工具的包装 div 从"网格单元格"改为 `display: contents`：包装盒不产生自身格子，其子元素（卡片）直接参与父网格布局。效果：所有卡片扁平化进 `#card-grid`，同时——
- 各工具 `renderCards()` 仍只替换自己的包装内容，**独立刷新**（久坐提醒 1s 不拖累其他卡片）；
- 包装节点在 DOM 中位置固定，卡片**不会因其他工具刷新而重排跳动**；
- 空占位卡逻辑不变，占位卡同样成为网格单元。

### D2. 顺序 = 工具注册顺序
DOM 中工具包装的先后即注册顺序（久坐提醒 → 倒数日），天然稳定。工具内部保持各自原有顺序（倒数日按剩余天数升序，由后端 `CountdownEvents` 排序保证）。顺序来源是 DOM 隐式的；未来做自由拖动排序时，再把顺序提升为显式状态。

### D3. 4 列等宽 + 响应式降列
`.card-grid` 改为 `grid-template-columns: repeat(4, 1fr)`（等宽 4 列）。窗口变窄时用媒体查询降列，避免卡片被压成无法阅读的窄条：`≤1024px → 2 列`，`≤640px → 1 列`（沿用现有 `@media` 断点风格）。
- 备选 auto-fill：列数随宽度浮动，表格感弱于固定列数，弃用。

### D4. 卡片同行不等高
倒数日卡无进度条行，比久坐提醒卡矮。网格默认 `align-items: stretch` 使同行的矮卡撑到等高，表格感更强，接受该行为。

## Risks / Trade-offs

- [`display: contents` 兼容性] → WebKitGTK（Wails Linux）与 WebView2（Windows）自 2018 年起均支持；无降级风险。
- [窄窗口 4 列挤压] → 媒体查询降列兜底。
- [固定 4 列在极宽窗口下卡片过宽] → dashboard 容器已有 `max-width: 1240px`，单卡宽度有上界，可接受。

## Migration Plan

1. `dashboard.js`：slot 类名/样式改为 contents 包装。
2. `app.css`：网格 4 列 + 媒体查询；移除 `.card-slot` 的 `min-height`（contents 盒不产生高度，由卡片自身决定）。
3. 验证：`npm run build`；真机确认表格排列、独立刷新不跳动、点击路由、空占位卡。
4. 纯前端改动，无数据迁移；回滚 = 恢复 dashboard 视图与网格样式。

## Open Questions

- 空占位卡的呈现方案（更宽的虚线位等）后续演进。
- dashboard 自由拖动排序：把顺序提升为显式状态，另开 change。
