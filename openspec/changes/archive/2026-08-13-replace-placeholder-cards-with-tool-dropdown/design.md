## Context

dashboard 目前通过 `renderCards()` 收集各工具卡片；空工具被 `dashboard.js` 用 `makePlaceholder()` 补一张"暂无内容"占位卡，既充当导航入口又充当空态。占位卡 key 用工具 id（`countdown`），与真实卡片 key（`countdown:<id>`）不一致，会在 `card_order.json` 里留下失效 key。详情页自身已有完整空态文案（"还没有倒数日…"），占位卡属于重复。

工具注册表 `frontend/src/tools/index.js` 是前端唯一事实来源，每个工具对象已有 `id` 与 `name`，下拉导航无需扩展任何契约。整项改动纯前端，无 Go 绑定变更。

## Goals / Non-Goals

**Goals:**

- 把"进入工具"从卡片剥离到顶栏下拉，卡片回归纯状态组件。
- 空工具不再渲染占位卡，不再参与网格与排序。
- 新增工具仍只需改 `tools/index.js` 一行注册，即可同时出现在下拉。

**Non-Goals:**

- 详情页不做工具切换下拉（保留「‹ 返回」按钮），也不做"当前工具"高亮。
- 下拉不携带各工具的状态提示（避免把空态职责又背回导航）。
- 不清理历史 `card_order.json` 文件（失效 key 由既有加载逻辑丢弃，无需迁移脚本）。

## Decisions

### 1. 下拉列全部工具，固定不随空态变化

下拉直接遍历 `Object.values(registry)` 渲染 `tool.name`，点击调用既有 `onOpenTool(tool.id)`。不区分"有数据/无数据"：菜单是稳定的导航枢纽，内容不随数据增删而跳动。

替代方案（只列空工具）被否决——菜单内容动态变化会让用户找不到工具，也增加状态判断逻辑。

### 2. 用原生 `<details>/<summary>` 承载下拉

```html
<details class="tool-menu">
  <summary class="tool-menu-toggle">工具 ▾</summary>
  <ul class="tool-menu-list" role="menu">
    <li><button role="menuitem" data-tool="reminder">久坐提醒</button></li>
    ...
  </ul>
</details>
```

理由：原生 `<details>` 免费获得 Enter/Space 展开、`aria-expanded` 同步，免去手写开关状态。选中某项后进入详情页，整个 dashboard 视图随之卸载重建，下拉的展开态无需额外清理。

替代方案（button + `aria-expanded` + 手写 open/close + 点击外部关闭 + Escape）被否决——引入更多状态管理，收益为零。

### 3. 删除占位卡逻辑，空工具自然不渲染

- 删除 `dashboard.js` 的 `makePlaceholder()`，及 `mountCards`/`renderToolCards` 里两处 `if (cards.length === 0) cards = [makePlaceholder(tool)]`。
- `renderCards()` 返回空数组时，既有 `for (const card of cards)` 循环自然不产出节点，网格与顺序对账无需改动。
- 删除 `app.css` 的 `.tool-card-empty*` 三条规则。

### 4. 占位 key 无需迁移

`card_order.json` 中残留的占位 key（等于工具 id）在 `mountCards` 的 `keyToElement` 里已不存在对应元素，走既有"失效 key 丢弃"分支自动忽略，无需额外代码。

## Risks / Trade-offs

- **[可发现性] 空工具首次安装时只在顶栏可见** → 下拉固定在 topbar 右侧常驻，位置稳定；注册表新增工具自动入列，不依赖用户配置。
- **[样式] `<details>` 自带 marker 与默认样式** → 用 CSS 覆盖 marker 并给菜单绝对定位 + 阴影，与现有 `.topbar`/`.back-button` 视觉一致。
- **[键盘] 下拉项需 `role="menuitem"` 的 button 承载** → 用真 `<button>`（可 Tab 聚焦、Enter 触发），`<ul role="menu">` 提供语义。
