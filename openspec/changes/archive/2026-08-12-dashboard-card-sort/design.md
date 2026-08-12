## Context

dashboard 由 `views/dashboard.js` 渲染：每个工具一个 `card-group`（`display: contents`）包装自己的卡片，卡片在网格内按"组"分块流动。组边界固定为注册顺序，因此倒数日事件卡永远无法排到久坐提醒卡之前。

顺序对账的前提是网格完全扁平（卡片是直接子元素）。拖拽排序是平台层（tool-dashboard）能力，与刚归档的 `dashboard-card-grid`（4 列等宽扁平化）一脉相承。

## Goals / Non-Goals

**Goals:**
- dashboard 卡片跨工具自由拖拽排序，顺序持久化，重启保持。
- 新增卡片自动出现在末尾。
- 渲染/刷新时卡片位置稳定，不因工具刷新而跳动。
- 顺序存储于 Go 后端 `card_order.json`，与 settings/countdowns 同级。

**Non-Goals:**
- 卡片增删的排序 UI（如"置顶/置底"按钮）——拖拽已覆盖，不做按钮。
- 键盘可达的排序操作（HTML5 DnD 仅鼠标）；卡片点击（Enter 进详情）仍键盘可达。
- 拖动时实时跨工具动画重排的精细过渡效果——先做基础 drop 重排。
- dashboard 其他配置化（隐藏/缩放卡片）仍后续再做。

## Decisions

### D1. 网格扁平化：卡片为直接子元素
弃用 `card-group`（`display: contents`）分组结构，工具产出的卡片直接 append 进 `#card-grid`。每张卡片是独立的网格单元，DOM 顺序即展示顺序，跨工具任意排列。
- 代价：工具刷新不再"替换整个组"，改为按 `data-card` key 对单个节点原位替换/追加/移除。
- 备选（保留分组 + 仅组内排序）：无法跨工具排列，不满足"随便排序"，弃用。

### D2. 卡片身份 key 与工具契约
每张卡片带 `data-card` 稳定 key：`reminder`、`countdown:<eventId>`、占位卡用工具 id（如 `countdown`）。`renderCards()` 契约新增要求：返回的卡片必须设置 `data-card`。
- 两个既有工具模块同步补 key；未来新工具按契约遵守。

### D3. 顺序持久化于 Go 后端
新增 `card_order_store.go`：`card_order.json` 存 `{ "order": ["countdown:456", "reminder", ...] }`，读写照抄 `timeline_store.go` 模式（用户目录 `health-tool/` 下）。
```
App.GetCardOrder() []string      // 读，空文件返回 nil
App.SaveCardOrder(order) bool    // 写，全量覆盖
```
启动时加载进 `App` 内存，供前端 `GetCardOrder` 查询。前端在 drop 后调 `SaveCardOrder`。
- 理由：项目所有配置（settings/countdowns/timeline）都经 Go 持久化，位置确定（`AppData\Roaming\health-tool\`）、可备份；localStorage 位于 WebView2 浏览器数据目录，位置不透明、有被清理风险。

### D4. 渲染对账
```
挂载:  Promise.all(各工具 renderCards())
       → 收集全部卡片 key（当前有效集合）
       → 读存储顺序，过滤失效 key
       → 按顺序渲染：命中存储 key 的按序插入，新 key 追加末尾
       → 无存储时默认注册顺序 + 工具内部顺序

刷新:  某工具 renderCards() 返回当前集合
       → 已有 data-card 节点 → 原位替换内容
       → 新 data-card → append 到网格末尾
       → 该工具消失的 key → 移除节点
```
只有拖拽 drop 显式 `SaveCardOrder`；增删卡片由对账自然处理（新卡末尾、删卡移除），加载时存储中的失效 key 丢弃。

### D5. 拖拽交互：HTML5 drag & drop，整卡可拖
卡片 `draggable="true"`：
- `dragstart`：记录源 key，加 `.dragging` 类（半透明）。
- `dragover`：`preventDefault` 允许 drop；按拖过目标卡片中线决定"插到前面/后面"，加插入指示。
- `drop`：把源节点移动到目标位置；序列化 DOM 卡片顺序 → `SaveCardOrder`。
- `dragend`：清除拖拽态样式。
- 点击不动 = click 进详情（HTML5 DnD 靠移动阈值区分点击与拖动，互不干扰）。

### D6. 拖拽与刷新互斥
拖拽开始（dragstart）时设 `dragging=true`，dashboard 的定时刷新检查该标志：正在拖拽的工具跳过本次刷新，dragend 后恢复。防止 1s 刷新替换节点打断拖动。

### D7. 占位卡参与排序
工具无卡片时 dashboard 生成的占位卡 key = 工具 id（如 `countdown`），同样可拖、可被排到任意位置。事件出现后占位卡被对账移除，新事件卡追加末尾。

## Risks / Trade-offs

- [WebKitGTK/WebView2 的 HTML5 DnD 差异] → dragover 需统一 `preventDefault`；真机验证 Linux/Windows 双端拖拽。
- [扁平化后工具刷新逻辑重写，回归风险] → 对账逻辑独立成函数并保持工具模块接口不变；真机回归验证刷新位置稳定。
- [拖拽期间节点被刷新替换] → D6 互斥标志兜底。
- [存储顺序与当前卡片集合长期不一致] → 每次挂载/刷新对账，失效 key 自动丢弃。

## Migration Plan

1. 后端先行：`card_order_store.go` + `app.go` 方法 + 启动加载 → `wails generate module`。
2. 前端：dashboard 扁平化（对账渲染）→ 工具模块补 `data-card` → 拖拽事件 → 样式。
3. 验证：`go test ./...`、`npm run build`、真机拖拽/重启保持/新增末尾。
4. 兼容：无存储文件时按默认顺序渲染（现状行为），旧用户无缝升级；无数据迁移。

## Open Questions

- 跨工具拖拽时的插入指示细节（整卡高亮 vs 中线分割线）实现时定。
- 是否需要在详情页也支持排序（超出范围，不做）。
