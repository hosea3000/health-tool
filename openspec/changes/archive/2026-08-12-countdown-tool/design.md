## Context

平台骨架（`multi-tool-dashboard`）已归档：dashboard 首页按工具注册表收集 `renderCards()` 产物渲染卡片网格，卡片按工具声明的 `refreshInterval` 调度刷新；点击卡片路由到该工具详情页。工具模块契约 = `{ id, name, refreshInterval, renderCards, renderDetail }`，新增工具只需登记 `tools/index.js`。

倒数日是其上第一个多实例参考型工具：一个事件 = 一张卡片，用户可配置多个事件；无后台常驻行为，纯数据展示。持久化可完全沿用 `timeline_store.go`/`settings.go` 的本地 JSON 模式。

## Goals / Non-Goals

**Goals:**
- 四种到期规则：一次性日期、每月几号、每周周几、大小周周几。
- 一个事件一张 dashboard 卡片，可配置多个。
- 剩余天数语义三段式：还剩 N 天 / 今天 / 已经 N 天（一次性过期）。
- 规则引擎为纯 Go 逻辑，置于 `domain` 包并有单元测试。
- 配置（新增/编辑/删除）全部在倒数日详情页内完成。

**Non-Goals:**
- 到期的系统通知或提醒（后续再议）。
- dashboard "+" 添加瓦片（后续做，本次从详情页新增）。
- 浏览器预览模式的假数据（无 Wails 桥时显示空状态）。
- 事件倒计时到小时粒度（只做日期粒度）。

## Decisions

### D1. 统一规则模型
```
Rule  { Type: date | monthly | weekly | biweekly }
  date     → Target: "2006-01-02"            // 一次性
  monthly  → Day: 1..31                      // 每月几号
  weekly   → Weekday: 0..6（周一=0）          // 每周周几
  biweekly → Weekday: 0..6, Phase: big|small, Anchor: "2006-01-02"  // 大小周
```
`NextOccurrence(rule, now)` 返回 `>= now` 的下一次到期日（date 类型返回固定 Target）；`Validate(rule)` 校验字段。

### D2. 大小周相位锚定（方案 C，修订）
大小周 = 每隔两周的周几。**锚周（创建日所在周）固定为大周**，与选择无关；表单的"大周/小周"开关只决定事件落在哪种周上：选"大周"从本周开始，选"小周"从下周开始，之后每两周一次。存 `Phase` + `Anchor=创建日`。相位判定：与锚周周数差为偶数=大周，奇数=小周。这样用户可以选择让事件从"下周"开始。
- 原方案把锚周相位绑定到开关（选小周则本周即小周），导致事件永远落在创建周，无法设成下周开始，已修订。
- 备选 A（ISO 周奇偶）：与用户真实周期不一定对齐，弃用。

### D3. 每月短月钳位
"每月 31 号"在短月取该月最后一天（2月→28/29，4/6/9/11→30），而不是跳月。

### D4. 剩余天数语义
`remainingDays = NextOccurrence - today`（日期粒度，Local 时区）。
- `> 0` → 卡片大数字 = N，文案"还剩 N 天"
- `== 0` → 文案"今天"
- `< 0`（仅一次性过期）→ 大数字 = |N|，文案"已经 N 天"
周期型事件的 NextOccurrence 恒 `>= today`，不会出现负数。

### D5. 排序
列表与卡片按剩余天数升序：未来事件（今天=0 最先）在前，已过事件排在所有未来事件之后、按最近优先。这是对"剩余天数升序"的解释：负数按字面会排最前，但已过事件更适合作为归档区殿后。

### D6. 标题上限 20 字
表单校验 + 卡片展示截断，保证 `minmax(280px,1fr)` 网格内排版不破。

### D7. 存储与 API
`~/.config/health-tool/countdowns.json`：`{ "events": [CountdownEvent] }`，无日期滚动（与 timeline 按天滚动不同）。ID 用纳秒时间戳字符串，不引入 uuid 依赖。
```
App.CountdownEvents() []CountdownView     // 事件 + NextDate + RemainingDays，已排序
App.AddCountdown(title, rule) bool
App.UpdateCountdown(id, title, rule) bool
App.DeleteCountdown(id) bool
```
`CountdownView` 返回排序后的展示数据，前端只渲染。

### D8. 删除二次确认
详情页删除前弹确认，防误删。

### D9. 刷新与预览
`refreshInterval: 60000`（分钟级，跨午夜由轮询自然覆盖）。无 Wails 桥时 `CountdownEvents` 返回空列表，前端渲染空状态。

### D10. 空工具的可点击占位卡（实现时补充）
实现中发现：事件为零时 dashboard 无卡片可点，无法进入详情页新增。平台层在工具 `renderCards()` 返回空数组时，渲染一张可点击的占位卡（kicker=工具名、value=—、点击进入详情页），保证每个工具的详情页始终可达。这是对"dashboard 入口"类问题（原方案 B 推迟项）的轻量替代。

## Risks / Trade-offs

- [大小周相位默认"大周"若与用户实际不符，一半时间倒计时偏差一周] → 创建/编辑表单里该开关醒目展示；编辑随时可改。
- [排序对"已过殿后"的解释与"剩余天数升序"字面不完全一致] → 在 specs 中明确写死排序规则，避免实现歧义。
- [浏览器预览看不到真实数据] → 接受；`wails dev` 下开发验证，预览模式仅保证不报错。
- [每月钳位与"跳过"语义的用户预期可能不同] → 已确认钳位；文案"每月15号"对月末事件的用户心智一致。

## Migration Plan

1. 后端先行：`domain/countdown.go` + 单测 → `countdown_store.go` → `app.go` 方法。
2. 前端接入：`tools/countdown/index.js`（cards + detail）→ 注册表登记。
3. 验证：`go test ./...`、`npm run build`、真机走通"新增→卡片出现→编辑→删除"。
4. 纯新增，无既有数据迁移；回滚 = 移除模块与注册。

## Open Questions

- 是否需要"到期的系统通知"：暂不做，留待后续 change。
- dashboard "+" 瓦片：后续单独 change，平台需加 `canCreate` 声明与路由 action 参数。
