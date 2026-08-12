## 1. 前端模块拆分

- [x] 1.1 拆分 `frontend/src/main.js`：抽出视图路由（`dashboard` / `detail` 两态切换）
- [x] 1.2 建立工具注册表结构 `frontend/src/tools/reminder/index.js`，封装 `renderCards` / `renderDetail` / `refreshInterval`
- [x] 1.3 实现 dashboard 视图（`frontend/src/views/dashboard.js`）：收集各工具 `renderCards()` 产物，按各自间隔调度刷新，渲染卡片网格
- [x] 1.4 实现 detail 视图（`frontend/src/views/detail.js`）：路由到 `registry[activeTool].renderDetail()`，提供"‹ 返回"入口
- [x] 1.5 改造 topbar：dashboard 显示"健康工具箱"，去掉"01"编号与全局设置按钮；detail 显示"返回 + 面包屑"

## 2. 久坐提醒工具适配

- [x] 2.1 将现有 hero / insight / timeline 界面迁入久坐提醒 `renderDetail()`
- [x] 2.2 将设置弹窗迁入久坐提醒详情页（沿用现有设置卡片视觉与校验逻辑），移除 topbar 全局设置入口
- [x] 2.3 实现久坐提醒 `renderCards()`：卡片展示当前状态 + 计时预览，1 秒刷新（沿用 `Status()` 轮询）
- [x] 2.4 回归验证现有行为：提醒状态、时间线、设置保存（构建通过 + 逻辑直接移植，运行时行为待真机确认）

## 3. 验证

- [x] 3.1 `cd frontend && npm run build` 构建通过
- [x] 3.2 手动走通 dashboard → 详情 → 返回流程
- [x] 3.3 `go test ./...` 通过
