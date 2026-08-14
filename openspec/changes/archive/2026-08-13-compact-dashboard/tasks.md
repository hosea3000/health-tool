## 1. 压缩 dashboard 容器与标题

- [x] 1.1 `.dashboard` padding 48px → 32px
- [x] 1.2 `.dashboard-heading` margin-bottom 32px → 20px，`h1` 字号 `clamp(40px,5vw,56px)` → `clamp(32px,4vw,44px)`

## 2. 压缩网格与卡片

- [x] 2.1 `.card-grid` gap 16px → 12px
- [x] 2.2 `.tool-card` padding 24px → 18px
- [x] 2.3 `.tool-card-value` margin `20px 0 24px` → `12px 0 14px`，字号 `clamp(48px,7vw,72px)` → `clamp(36px,5vw,52px)`
- [x] 2.4 `.tool-card-top` padding-bottom 14px → 12px，`.tool-card-meta` padding-top 12px → 10px

## 3. 校验

- [x] 3.1 `cd frontend && npm run build` 构建通过
- [x] 3.2 `wails dev` 人工确认：卡片 4 列等宽、拖拽/点击/刷新行为不变、一屏可见卡片增多
