## Context

dashboard 首页由 `frontend/src/views/dashboard.js` 渲染，样式集中在 `frontend/src/app.css`。当前密度相关样式：

| 选择器 | 现状 |
| --- | --- |
| `.dashboard` | `padding: 48px` |
| `.dashboard-heading` | `margin-bottom: 32px` |
| `.dashboard-heading h1` | `font-size: clamp(40px, 5vw, 56px)` |
| `.card-grid` | `gap: 16px`（4 列） |
| `.tool-card` | `padding: 24px` |
| `.tool-card-value` | `margin: 20px 0 24px; font-size: clamp(48px, 7vw, 72px)` |
| `.tool-card-top` | `padding-bottom: 14px` |
| `.tool-card-meta` | `padding-top: 12px` |

只改 CSS，不动 `dashboard.js`、卡片 DOM 结构、拖拽排序与刷新逻辑。

## Goals / Non-Goals

**Goals:**
- 在不增加列数、不改卡片结构的前提下，让一屏可见的卡片明显增多。
- 保持视觉层级：数值仍是卡片主角，紧凑但不可读性受损。

**Non-Goals:**
- 不改 4 列等宽布局与响应式断点（1024px/640px）。
- 不改 topbar、footer、详情页、设置弹窗等非 dashboard 区域。
- 不引入新的布局系统或依赖。

## Decisions

### D1：只调 CSS 数值，不动结构
改为直接修改 `app.css` 中 dashboard 相关声明的具体值，而非引入 CSS 变量或重构。原因：一次性密度调整，无多处复用需求，变量化属过度设计。跳过 `ponytail` 的"先加变量"诱惑，直接改字面量。

### D2：各值的压缩幅度
- `.dashboard` padding `48px` → `32px`
- `.dashboard-heading` margin-bottom `32px` → `20px`；`h1` `clamp(40px,5vw,56px)` → `clamp(32px,4vw,44px)`
- `.card-grid` gap `16px` → `12px`
- `.tool-card` padding `24px` → `18px`
- `.tool-card-value` margin `20px 0 24px` → `12px 0 14px`；字号 `clamp(48px,7vw,72px)` → `clamp(36px,5vw,52px)`
- `.tool-card-top` padding-bottom `14px` → `12px`
- `.tool-card-meta` padding-top `12px` → `10px`

理由：约 1/3 的间距压缩，数字字号降到原约 72% 仍清晰；`tabular-nums` 与 `letter-spacing` 保持不变，保证数值可读。

### D3：保持移动端断点不变
`@media (max-width: 800px)` 里 `.dashboard { padding: 32px 24px }` 已相对紧凑，不重复调整；只动桌面断点的默认值。窄屏降列逻辑（`.card-grid` 1024px/640px）保持原样。

## Risks / Trade-offs

- [卡片过挤、数值与文案贴边] → 通过保留卡片 `border-radius: 16px`、`border` 与 18px 内边距维持呼吸感；若观感仍挤，回退 padding 到 20px。
- [数值缩小后老花/高 DPI 用户可读性下降] → 保留最小 `clamp` 下限（36px）与 tabular-nums，不无限缩小。
- [回归无自动化视觉测试] → 前端构建（`npm run build`）作为唯一校验，配合 `wails dev` 人工确认。

## Migration Plan

无数据迁移。改动仅影响渲染样式，构建前端后即时生效；不满意可回滚单个 CSS 提交。
