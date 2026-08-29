# 提案：久坐提醒通知开关（静默记录模式）

## Why

用户在某些场景下（演示、会议、专注冲刺）不希望被久坐提醒打扰，但仍然希望工作时长被如实记录。当前应用没有任何关闭提醒的途径：窗口关闭后托盘驻留继续提醒，闲置暂停和锁屏只是临时中断。需要一个只剪断「打扰」、保留「记录」的开关。

## What Changes

- 新增「久坐提醒通知」开关，默认开启，持久化到 `settings.json`。
- 开关只在久坐提醒工具详情页的偏好设置弹窗中配置；不提供托盘快捷开关。
- 通知关闭（静默记录模式）时：
  - 监控继续运转，有效活动照常创建工作段，闲置暂停照常触发；
  - 工作段达到提醒时长后**不进入提醒休息期**，状态保持工作段、计时持续增长，不产生 `resting` 时间线记录（保证工作总时长统计真实）；
  - 不发送久坐提醒的原生通知与前端 `reminder` 事件。
- 通知重新开启时：从当前时刻起重新计算本轮工作段的提醒时点（顺延「已工作时长 + 提醒时长」），避免下一秒立刻补弹；开关关闭立即生效。
- 开关作用于久坐提醒工具的全部桌面通知：静默期间不发送久坐提醒（原生通知与前端 `reminder` 事件）也不发送「自动开始通知」。
- 久坐提醒卡片与详情页在静默时展示静默标记（如 🔇），状态文案仍如实展示工作段进行中；进度条满格后保持满格，计时数字继续增长。

## Capabilities

### New Capabilities

- `reminder-notification-toggle`: 久坐提醒通知开关——默认开启、持久化、关闭时只剪断提醒与提醒休息期转换而保留全部记录，开启时从当前时刻重新计时的完整行为。

### Modified Capabilities

- `work-rest-timeline`: 「工作段达到提醒时长进入提醒休息期并记录 `resting` 区间」的需求增加条件——仅当通知开关开启时成立；静默模式下工作段达到提醒时长后继续计时，不产生 `resting` 记录，累计工作时长口径因此保持真实。

## Impact

- `domain/monitor.go`：`Monitor` 增加通知开关字段；`Advance` 在到点时按开关分流（开 → Resting + Reminder，关 → 停留 Working）；重新开启时顺延 `activeReminderAfter`。
- `model/model.go`：`Settings` 增加 `NotificationsEnabled bool` 字段（默认 true；`LoadSettings` 从 `DefaultSettings()` 起始再 unmarshal，老配置文件缺失该字段时自然保持开启，无迁移成本）。
- `store/settings.go`：`SaveSettings` 持久化新字段。
- `app.go`：`SaveSettings` 绑定方法签名增加开关参数（**前后端数据契约变更**，`frontend/wailsjs` 需重新生成）；设置变更时同步 `Monitor`。
- `frontend/src/tools/reminder/index.js`：设置弹窗增加开关控件；卡片/详情页展示静默标记与超时计时的显示兼容（进度条已钳位 100%，大数字继续增长）。
- `CONTEXT.md`：新增「静默记录」词条。
- 无新增外部依赖；托盘、输入钩子、锁屏检测均不改。
