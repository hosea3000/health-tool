## Context

应用是 Wails v2 桌面应用：Go 后端（`main.go` 入口、`app.go` 协调、`domain/monitor.go` 状态机）负责久坐提醒的输入监听、计时、提醒与时间线持久化；前端是无框架 Vite 应用，`frontend/src/main.js` 以一段硬编码模板渲染单工具界面，通过 Wails 绑定调用 `Status`/`Timeline`/`GetSettings`/`SaveSettings`。

现状中已有多工具伏笔：topbar wordmark 为"健康工具箱"，工具名旁带"01"编号。但前端没有任何路由或组件化结构，全部内容（hero、insight 卡、时间线、设置弹窗）都写死在 `#app` 的 innerHTML 里。

目标场景：应用需要承载多个工具（久坐提醒 + 未来的倒数日等），且倒数日一类工具"一个工具多张卡片"（每个事件一张卡片）。需要先立起一个平台骨架，再逐个接入工具。

## Goals / Non-Goals

**Goals:**
- dashboard 为应用首页，以卡片网格陈列各工具的实例预览。
- 建立工具/卡片分层：工具是类型，卡片是实例；一个工具可产出 1..N 张卡片。
- 点击卡片进入该工具详情页（工具级聚合），详情页可返回 dashboard。
- 卡片刷新频率由工具自声明，dashboard 统一调度。
- 新增工具只需在前端注册表登记，不改 dashboard 本体。
- 保持无框架技术栈，不引入新依赖。

**Non-Goals:**
- 本次不实现倒数日工具本身，只为其预留接入位。
- dashboard 可配置化（卡片排序/隐藏/缩放）后续再做。
- 不引入 Go 端工具接口（Tool interface）——第三个工具出现前 YAGNI。
- 不改 Go 后端结构与领域状态机。

## Decisions

### D1. 工具=类型，卡片=实例
dashboard 的展示单元是**卡片**而非工具。工具是编译进程序的类型，负责产出一张或多张卡片：久坐提醒产出 1 张运行时派生的固定卡片（无持久化实例），倒数日产出 N 张卡片（每张对应一个持久化事件）。
- 备选：工具列表 + 每工具子页面，无法表达"一工具多卡片"。
- 备选：每张卡片绑定完整独立详情页，用户已否决（直接进入工具详情页即可）。

### D2. 前端工具注册表 + 视图路由，无框架
前端结构拆为两层：
```
main.js                视图路由：state = {view:'dashboard'} | {view:'detail', tool}
views/dashboard.js     收集各工具 renderCards() 产物，按 interval 调度刷新，渲染网格
views/detail.js        路由到 registry[activeTool].renderDetail()，提供返回入口
tools/reminder/        久坐提醒工具模块 { renderCards, renderDetail, refreshInterval }
tools/countdown/       （未来）同构接入
```
- 卡片数据来源：各工具模块自行调用其 Go API（reminder 沿用 `Status`/`Timeline`），不做后端统一 `DashboardCards()` 聚合。理由：工具边界清晰、后端零改动。
- 备选：Go 端工具接口 + 统一卡片 JSON 契约。两个工具就上接口，过度抽象。

### D3. 详情页按工具聚合
点击同一工具的多张卡片都进入同一个工具详情页。`instanceId` 作为可选 focus 参数保留在路由里，v1 可不传（用户已确认直接进工具详情页即可）。

### D4. 卡片刷新频率由工具声明
工具模块声明 `refreshInterval`，dashboard 调度器按间隔分组刷新：`reminder: 1000ms`（沿用现状每秒刷新），未来 `countdown: 60000ms`。调度器实现为"每个间隔一组工具的定时器"，不做复杂度。

### D5. 设置下沉到工具详情页
topbar 的全局"设置"按钮移除；久坐提醒的提醒/休息时长设置弹窗迁入其详情页。每个工具在自己的详情页内提供设置/管理入口。这是对现有 UI 唯一的破坏性改动。

### D6. topbar 调整
dashboard：仅显示"健康工具箱"，去掉"01"编号与全局设置。detail：提供"‹ 返回" + 面包屑"健康工具箱 / <工具名>"。

### D7. Go 端保持肥 App
本次不引入任何 Go 结构性改动。新工具（倒数日）实现时若确实需要共享生命周期，再考虑抽取。

### D8. 卡片契约：结构化最小字段
平台为卡片定义最小结构化字段（title / value / unit / accent 等），工具只填数据，平台统一排版。富样式留给详情页。为后续 dashboard 配置化（排序/隐藏）预留基础。

## Risks / Trade-offs

- [单文件 main.js 拆分为多模块，改动面大] → 拆分后逐项回归验证现有行为（提醒状态、时间线、设置保存）；Go 端零改动，回滚只需 revert 前端。
- [设置入口从 topbar 移到详情页，用户可能一时找不到] → 详情页内提供显眼的设置入口（沿用现有设置卡片视觉）。
- [统一卡片模板限制卡片富样式] → 卡片只做预览；需要富展示时进详情页。
- [每工具自管刷新，dashboard 调度逻辑分散] → 调度器只按 interval 分组，契约简单；倒数日接入时若发现不足再演进。

## Migration Plan

1. 拆分前端模块：路由 + dashboard + detail + reminder 工具模块，行为保持与现状一致（先让 reminder 以新结构跑起来）。
2. 迁移设置弹窗到 reminder 详情页，移除 topbar 全局设置与"01"编号。
3. 验证：`go test ./...`、`cd frontend && npm run build`、手动走通 dashboard→详情→返回。
4. 无 Go 端改动，无数据迁移；回滚 = 恢复前端旧版。

## Open Questions

- 卡片契约的最终字段集合（如 accent 是否需要）在实现时确定。
- 路由中的 `instanceId` focus 参数 v1 是否需要，可延后。
