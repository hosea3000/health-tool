# 设计：久坐提醒通知开关（静默记录模式）

## Context

`domain.Monitor` 是四态状态机（Waiting / Working / IdlePaused / Resting），由全局输入钩子的有效活动与每秒 tick 驱动；工作段达到提醒时长即进入提醒休息期并发送通知。设置目前只有两个时长字段（`ReminderMinutes` / `RestMinutes`），经 `SaveSettings` 绑定方法贯通前后端。时间线记录状态转换区间，前端对 `working` 记录求和得到当日工作总时长。

探索阶段已与用户确认三个关键取向：开关只关通知、记录必须继续、不提供托盘入口（仅在久坐提醒详情页设置弹窗配置）。

## Goals / Non-Goals

**Goals:**

- 提供默认开启、持久化的久坐提醒通知开关；
- 静默期间监控与记录完全照常：工作段继续增长、闲置暂停照常触发、工作总时长统计真实；
- 重新开启时不惊吓用户（不立刻补弹提醒）；
- 老配置文件无迁移成本。

**Non-Goals:**

- 不做托盘菜单快捷开关（用户明确不需要）；
- 不做临时勿扰（限时静默后自动恢复）；
- 不做「静默记录但照常进休息期」的变体（会制造虚假 `resting` 记录，见决策 1）；
- 不改输入钩子加载/卸载行为（钩子照跑，事件在状态机内消化）。

## Decisions

### 决策 1：静默期到点不进入提醒休息期（而非"进休息期但不弹通知"）

`Advance` 在工作段达到提醒时长时按开关分流：开 → 现行为（Resting + `Result.Reminder`）；关 → 停留 Working，计时持续增长。

- 为什么：若照常进 Resting，时间线会每轮插入一段用户并未休息的 `resting` 记录，且休息期内输入不算有效活动会扭曲工作段划分；工作总时长（对 `working` 求和）也会少算。「关通知 + 记录真实」只在停留 Working 下成立。
- 备选否决：进入 Resting 但不调 `notifyReminder`——状态机与时间线说谎。
- 前端兼容性：进度条 `derive()` 已用 `Math.min(..., 1)` 钳位 100%，计时大数字直接取 `elapsedSeconds`，超时显示（如 `92:30`）零改动天然正确。

### 决策 2：开关双向生效时机不对称

- **关闭立即生效**：下一秒 tick 即静默，当前段不再触发提醒。否则用户刚关完开关又被弹一次，体验背叛。
- **开启从当前时刻重新计时**：开启时若正处于 Working，将本轮 `activeReminderAfter` 顺延为「已工作时长 + `reminderAfter`」，即从现在起再计满一个提醒时长才提醒。

  为什么不是"下一段生效"：用户可能连续工作数小时不闲置，期间以为通知开着却永远不响。为什么不是"立即按剩余量提醒"：静默中超时 75 分钟后开启，下一秒立刻弹窗同样惊吓。顺延是唯一两全解。

  实现注意：顺延值可能超过 `MaxReminderDuration`（180 分钟校验上限），但该上限只约束**设置值**的校验（`ValidReminderDuration`），`activeReminderAfter` 是本轮生效值，本就允许独立于设置值（见现行代码中两者分离的设计），无需放宽校验。

- **非 Working 状态下开启**：无需处理，下次工作段创建时自然按设置值初始化 `activeReminderAfter`。

### 决策 3：不引入第五个状态，开关是 Monitor 的输入而非状态

静默不是状态机的一个状态（状态机没有停），而是 `Working → Resting` 这条转换边的门闸。开关字段放 `Monitor` 内部（如 `notificationsEnabled bool`），由 `SetNotificationsEnabled` 控制。

- 为什么：符合「`Status()` 是状态唯一真相源、UI 不重建状态机」的项目铁律——前端状态文案不变，工作段照常显示"工作段进行中"，静默仅通过设置值 + 静默标记表达。
- 备选否决：App 层拦截（`recordActivity` / `tick` 入口丢弃事件）——所有入口都要记得检查，漏一处漏提醒；且关闭时的状态残留（Elapsed、timeline）处理更琐碎。

### 决策 4：设置字段用 `bool` 默认 true，依赖现有加载姿势化解零值陷阱

`Settings` 增加 `NotificationsEnabled bool`，`DefaultSettings()` 返回 true。`LoadSettings` 已是「从 `DefaultSettings()` 起始再 unmarshal」的写法，老配置文件缺失该字段时保持 true，**升级用户无感、无迁移**。

- 风险点：测试代码直接构造 `model.Settings{}`（零值 false）的地方需显式置 true；排查 `newAppWithSettings` 相关测试即可。
- 备选否决：`*bool`（nil = 默认开）或反向语义 `Disabled`——本仓库加载路径已正确，引入间接性不值。

### 决策 5：`SaveSettings` 绑定签名扩展，静默标记由设置值推导

`SaveSettings(reminderMinutes, restMinutes int, notificationsEnabled bool)`——前后端数据契约变更，需跑 `wails dev` / `wails build` 重新生成 `frontend/wailsjs`。

静默标记：`AppStatus` 增加 `NotificationsEnabled bool` 字段（`Status()` 随手返回，前端每秒轮询已有），卡片据此渲染 🔇 标记，无需新事件。

- 为什么放 `AppStatus` 而非前端只读 `GetSettings`：卡片每秒已拿 `Status()`，零额外请求；且状态展示与数据契约解耦。

### 决策 6：开关门控久坐提醒工具的全部桌面通知（含自动开始通知）

初版设计曾保留「自动开始通知」（新工作段自动创建时的「新的工作段已开始」通知），理由是开关语义边界清晰。用户验收反馈：关闭通知后首次打开软件、首次敲击键盘即被弹窗，与「关闭通知」的直觉预期相违。修正为：静默即全部桌面通知静默，`recordActivity` 中 `notifyStarted` 与 `notifyReminder` 一样受开关门控（前者在调用点门控，后者经由 `Result.Reminder` 从源头消失）。

- 为什么不拆两个开关：用户心智里只有一个「通知」概念，拆分增加设置页复杂度又无实际收益。

## Risks / Trade-offs

- [静默中用户忘记开关状态，误以为提醒坏了] → 卡片与详情页展示 🔇 静默标记；`Status()` 持续下发开关状态。
- [顺延后的 `activeReminderAfter` 超过设置上限，测试断言混淆] → 单测明确区分设置校验值与本轮生效值；顺延逻辑单独覆盖。
- [`SaveSettings` 签名变更导致旧前端调用报错] → 前后端同仓同版本发布，无兼容窗口；实现时改完 Go 侧立即重新生成 wailsjs 并同步前端调用。
- [静默期超长工作段（如 4 小时）在时间线上表现为一条超长 `working` 记录] → 这是真实数据，视为特性而非缺陷；展示层已能渲染任意时长。

## Migration Plan

1. 合入后老用户 `settings.json` 不含新字段 → 加载时保持默认开启，行为与升级前完全一致。
2. 无回滚复杂度：字段多余时旧版本会忽略（json unmarshal 容忍未知字段）。

## Open Questions

无——探索阶段已收敛（只关通知、记录继续、无托盘入口、A2 语义、开启重新计时均经用户确认或默认采纳）。
