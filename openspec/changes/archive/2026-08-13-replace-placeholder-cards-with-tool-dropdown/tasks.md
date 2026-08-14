## 1. 顶栏下拉导航

- [x] 1.1 在 `dashboard.js` 的 `renderDashboard` 顶栏 HTML 中新增"工具"下拉（`<details>`/`<summary>` + `<ul role="menu">`），遍历 `Object.values(registry)` 渲染工具名按钮
- [x] 1.2 为下拉项绑定点击，调用既有 `onOpenTool(tool.id)` 进入详情页
- [x] 1.3 在 `app.css` 新增 `.tool-menu` 相关样式（绝对定位、阴影、覆盖 `<details>` 默认 marker，与 topbar 视觉一致）

## 2. 移除占位卡

- [x] 2.1 删除 `dashboard.js` 中 `makePlaceholder` 函数
- [x] 2.2 删除 `mountCards` 与 `renderToolCards` 中 `if (cards.length === 0) cards = [makePlaceholder(tool)]` 两处占位注入
- [x] 2.3 删除 `app.css` 中 `.tool-card-empty` 相关三条样式规则

## 3. 验证

- [x] 3.1 `cd frontend && npm run build` 前端构建通过
- [x] 3.2 手动验证：空工具不出现在网格、无占位卡，且经顶栏下拉可进入详情页
- [x] 3.3 手动验证：配置倒数日/计数器后返回 dashboard，新卡片正常出现并追加末尾
- [x] 3.4 手动验证：下拉键盘可达（Tab/Enter），`aria-expanded` 正确切换
- [x] 3.5 手动验证：`card_order.json` 中残留的占位 key 被忽略，不影响正常卡片排序
