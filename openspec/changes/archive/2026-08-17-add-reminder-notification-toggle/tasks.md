# 任务：久坐提醒通知开关（静默记录模式）

## 1. 领域层（domain/monitor.go）

- [x] 1.1 `Monitor` 增加 `notificationsEnabled` 字段（默认 true）与 `SetNotificationsEnabled(enabled bool)` 方法：关闭立即生效；开启时若当前为 Working，将 `activeReminderAfter` 顺延为「已工作时长 + `reminderAfter`」，非 Working 状态不做特殊处理
- [x] 1.2 `Advance` 在工作段达到 `activeReminderAfter` 时按开关分流：开 → 进入 Resting 并返回 `Result{Changed: true, Reminder: true}`（现行为）；关 → 停留 Working 返回空 Result，计时持续增长
- [x] 1.3 领域单测覆盖：默认开启、关闭后到点不进 Resting 不置 Reminder、静默期计时持续增长、开启后顺延计时（含已超时场景）、非 Working 状态下开启、静默期闲置暂停照常、静默期新工作段照常创建

## 2. 设置模型与持久化（model/store）

- [x] 2.1 `model.Settings` 增加 `NotificationsEnabled bool` 字段（json: `notificationsEnabled`），`DefaultSettings()` 返回 true
- [x] 2.2 `store/settings_test.go` 补充用例：老配置文件缺失该字段加载后为 true；显式 false 正常加载；保存往返一致
- [x] 2.3 排查测试中直接构造 `model.Settings{}` 的位置，静默模式相关断言显式设置开关值，避免 bool 零值干扰

## 3. 应用层（app.go）

- [x] 3.1 `newAppWithSettings` 构造 `Monitor` 后应用设置中的开关初始值
- [x] 3.2 `SaveSettings` 签名扩展为 `(reminderMinutes, restMinutes int, notificationsEnabled bool)`，校验通过后调用 `monitor.SetNotificationsEnabled`；持久化包含新字段
- [x] 3.3 `model.AppStatus` 增加 `NotificationsEnabled bool` 字段，`Status()` 返回当前开关状态
- [x] 3.4 `app_test.go` 补充：静默模式下 `Status()` 不产生提醒且状态保持 working、elapsed 持续增长；开关切换的生效时机；`SaveSettings` 新签名往返

## 4. 前端（frontend/src/tools/reminder）

- [x] 4.1 设置弹窗增加「久坐提醒通知」开关控件（默认开），保存时随时长一起提交三参数 `SaveSettings`；开关关闭时两个时长输入保持可编辑
- [x] 4.2 卡片与详情页根据 `Status()` 返回的 `notificationsEnabled` 渲染 🔇 静默标记，状态文案不变
- [x] 4.3 验证超时显示：静默期进度条满格后保持满格、计时大数字继续增长（确认现有 `derive()` 钳位逻辑无需改动）
- [x] 4.4 无 Wails 桥的纯前端预览模式同步支持开关（localStorage 持久化）

## 5. 契约再生成与收尾

- [x] 5.1 跑 `wails dev` 或 `wails build` 重新生成 `frontend/wailsjs`（`SaveSettings` 与 `AppStatus` 变更），确认不手改生成产物
- [x] 5.2 `CONTEXT.md` 新增「静默记录」词条：通知开关关闭期间监控与时间线照常、到点不进提醒休息期、工作段持续计时；与「闲置暂停」术语对仗
- [x] 5.3 全量校验：`go test ./...` 与 `cd frontend && npm run build` 通过

## 6. 修正：静默同时门控自动开始通知

- [x] 6.1 `recordActivity` 中 `notifyStarted` 受 `monitor.NotificationsEnabled()` 门控，静默期不弹「新的工作段已开始」通知
- [x] 6.2 `app_test.go` 补充：静默模式下首次有效活动不弹开始通知，重新开启后恢复
- [x] 6.3 同步修正 spec（仅作用于久坐提醒场景反转）、proposal、design 决策 6、CONTEXT.md 词条与前端提示文案
