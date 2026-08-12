## 1. dashboard 视图扁平化

- [x] 1.1 `views/dashboard.js`：工具槽 div 改为 `display: contents` 包装，卡片直接进入 `#card-grid` 网格
- [x] 1.2 保留空占位卡逻辑，占位卡随包装直接进入网格
- [x] 1.3 确认各工具独立刷新与点击路由行为不变（替换逻辑仅作用于自己的包装内容）

## 2. 网格样式

- [x] 2.1 `app.css`：`.card-grid` 改为 `grid-template-columns: repeat(4, 1fr)`
- [x] 2.2 新增媒体查询降列：`≤1024px` 2 列、`≤640px` 1 列（沿用现有断点风格）
- [x] 2.3 移除/调整 `.card-slot` 的 `min-height`（contents 盒不产生高度）

## 3. 验证

- [x] 3.1 `cd frontend && npm run build` 通过
- [x] 3.2 真机验证：多卡片 4 列表格排列、某工具刷新时顺序不跳动、点击进详情页、空占位卡混排可点击
- [x] 3.3 `go test ./...` 通过（确认无回归）
