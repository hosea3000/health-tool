## Why

dashboard 首页当前留白偏多：页容器 48px 内边距、卡片 24px 内边距、网格 16px 间距、卡片数值字号高达 clamp(48px,7vw,72px)。随着各工具产出的卡片越来越多，一屏可见的卡片数量偏少，信息密度低，需要在不改变卡片结构与 4 列布局的前提下压缩间距与字号，让更多卡片同时进入视野。

## What Changes

- 压缩 dashboard 页容器内边距（48px → 更小）与 heading 区块的外边距、字号。
- 压缩卡片网格间距与卡片内边距（24px → 更小）。
- 缩小卡片数值字号（clamp(48px,7vw,72px) → 更小）及其上下外边距。
- 保持 4 列等宽布局与响应式断点降列行为不变。
- 保持卡片结构、拖拽排序、刷新机制不变。

## Capabilities

### New Capabilities

<!-- 无 -->

### Modified Capabilities

<!-- 无：本次为纯视觉/实现层面的间距与字号调整，不改变任何 spec 级行为。 -->

## Impact

- 仅改动 `frontend/src/app.css` 中 dashboard 相关样式（`.dashboard`、`.dashboard-heading`、`.card-grid`、`.tool-card`、`.tool-card-value` 等）。
- 无后端、无数据契约、无依赖变更。
