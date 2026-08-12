## 1. 后端规则引擎

- [x] 1.1 `domain/countdown.go`：定义 `Rule`（date / monthly / weekly / biweekly 四形态）与 `CountdownEvent`，字段含校验
- [x] 1.2 实现 `NextOccurrence(rule, now)`：一次性返回固定日期；每周返回最近的周几；每月返回下一月的第 N 天（短月钳位月末）；大小周返回相位匹配的最近周几
- [x] 1.3 实现 `Validate(rule)`：标题非空且 ≤ 20 字、每月 1–31、周几 0–6、biweekly 相位合法
- [x] 1.4 `domain/countdown_test.go` 单测：一次性、每周、每月钳位（2 月 28/29、4 月 30）、大小周相位（本周/隔周/跨月）

## 2. 存储与 App 方法

- [x] 2.1 `countdown_store.go`：照抄 `timeline_store.go` 的 IO 模式，`countdowns.json`（`{events: []}`），含加载/保存与空文件容错
- [x] 2.2 `app.go` 追加 `CountdownEvents() []CountdownView`：返回事件 + 下次日期 + 剩余天数，按 D5 排序
- [x] 2.3 `app.go` 追加 `AddCountdown` / `UpdateCountdown` / `DeleteCountdown`：校验规则、持久化、ID 用纳秒时间戳
- [x] 2.4 `app.go` 在启动时加载 countdowns.json

## 3. 前端工具模块

- [x] 3.1 `tools/countdown/index.js`：登记 `{ id:'countdown', name:'倒数日', refreshInterval:60000 }`
- [x] 3.2 实现 `renderCards()`：每个事件一张卡片（标题 + 天数 + 下次日期），空列表返回空数组；"还剩/今天/已经"三段文案
- [x] 3.3 实现 `renderDetail()`：事件列表（复用三段文案与排序）+ "＋ 新增"按钮
- [x] 3.4 新增/编辑弹窗表单：规则类型切换、具体日期选择器、每月天数、每周/大小周周几、大小周相位开关、标题 20 字限制
- [x] 3.5 删除二次确认；编辑复用同一表单
- [x] 3.6 注册到 `tools/index.js`

## 4. 验证

- [x] 4.1 `go test ./...` 通过
- [x] 4.2 `cd frontend && npm run build` 通过
- [x] 4.3 真机走通：新增事件 → 卡片出现 → 编辑 → 删除 → 重启后事件仍在
