# Tool Dashboard

## Purpose

为多个健康工具提供统一的 dashboard 首页，以卡片网格展示各工具产出的实例预览，并支持点击卡片进入对应工具详情页。dashboard 本体逻辑与具体工具解耦，新工具通过注册表接入。

## Requirements

### Requirement: Dashboard 作为应用首页
应用启动后 SHALL 首先显示 dashboard 视图，该视图以卡片网格陈列各工具产出的实例预览，而不是直接显示某个具体工具的内容。

#### Scenario: 首次启动进入 dashboard
- **WHEN** 应用启动完成
- **THEN** 主窗口显示 dashboard 卡片网格，包含久坐提醒工具产出的一张预览卡片

#### Scenario: 关闭窗口后重新打开进入 dashboard
- **WHEN** 用户关闭主窗口（托盘驻留）后再次打开
- **THEN** 主窗口重新显示 dashboard 卡片网格

### Requirement: 工具产出一张或多张卡片
dashboard 的展示单元 SHALL 是卡片而非工具；一个工具 SHALL 能够产出一张或多张卡片，每张卡片对应一个实例预览。

#### Scenario: 单例工具产出一张卡片
- **WHEN** 久坐提醒工具产出卡片
- **THEN** dashboard 显示一张久坐提醒卡片，展示当前监控状态与计时预览

#### Scenario: 多实例工具产出多张卡片
- **WHEN** 未来某工具（如倒数日）存在多个实例
- **THEN** dashboard 为每个实例各显示一张卡片

### Requirement: 点击卡片进入工具详情页
点击任意卡片 SHALL 从 dashboard 切换到该卡片所属工具详情页；同一工具的多张卡片 SHALL 进入同一个工具详情页。

#### Scenario: 点击卡片进入对应工具详情
- **WHEN** 用户点击一张工具卡片
- **THEN** 界面切换到该工具的详情页，展示该工具完整内容

#### Scenario: 同一工具的多张卡片进入同一详情页
- **WHEN** 用户先后点击同一工具产出的不同卡片
- **THEN** 两次都进入该工具的同一个详情页

### Requirement: 详情页提供返回 dashboard 的入口
工具详情页 SHALL 提供返回 dashboard 的入口，且返回后 dashboard 状态（各卡片内容）保持可继续刷新。

#### Scenario: 从详情页返回 dashboard
- **WHEN** 用户在工具详情页点击返回
- **THEN** 界面回到 dashboard 卡片网格，卡片继续按各自间隔刷新

### Requirement: 顶栏工具下拉导航
dashboard 顶栏 SHALL 提供"工具"下拉列表，固定列出注册表全部工具（不随各工具是否有数据而变化），点击任意项 SHALL 切换到对应工具详情页；下拉 SHALL 键盘可达并标注 `aria-expanded` 状态。

#### Scenario: 下拉列出全部工具
- **WHEN** dashboard 顶栏的"工具"下拉被展开
- **THEN** 列表列出注册表登记的全部工具（含当前无任何卡片数据的工具）

#### Scenario: 点击下拉项进入工具详情页
- **WHEN** 用户在下拉列表中点选某个工具
- **THEN** 界面切换到该工具的详情页，展示该工具完整内容

#### Scenario: 空工具经下拉可达
- **WHEN** 某个工具没有任何卡片、因此不出现在卡片网格中
- **THEN** 用户仍可通过顶栏下拉进入该工具详情页

#### Scenario: 下拉键盘可达
- **WHEN** 用户通过键盘 Tab/方向键/回车操作下拉
- **THEN** 下拉可聚焦、可展开、可选中并进入工具详情页，且 `aria-expanded` 状态正确切换

#### Scenario: 新工具自动出现在下拉
- **WHEN** 一个新工具在注册表登记后
- **THEN** 该工具自动出现在顶栏下拉中，无需修改 dashboard 本体逻辑

### Requirement: 卡片刷新频率由工具声明
每个工具 SHALL 声明其卡片刷新间隔；dashboard SHALL 按各工具声明的间隔刷新对应卡片，而非对所有卡片使用统一频率。

#### Scenario: 秒级刷新的卡片
- **WHEN** dashboard 渲染久坐提醒卡片（声明 1 秒间隔）
- **THEN** 该卡片每秒刷新一次，计时预览持续更新

#### Scenario: 分钟级刷新的卡片
- **WHEN** dashboard 渲染未来倒数日卡片（声明 60 秒间隔）
- **THEN** 该卡片按 60 秒间隔刷新，无需逐秒更新

### Requirement: 新增工具通过注册表接入
新增工具 SHALL 通过前端工具注册表登记其 `renderCards`、`renderDetail` 与 `refreshInterval` 后接入 dashboard，无需修改 dashboard 本体逻辑。

#### Scenario: 登记新工具后 dashboard 自动纳入其卡片
- **WHEN** 一个新工具在注册表登记 renderCards
- **THEN** dashboard 无需改动即展示该工具产出的卡片

#### Scenario: 未登记的模块不产出卡片
- **WHEN** 某个模块未在注册表登记
- **THEN** dashboard 不显示该模块的任何卡片

### Requirement: 全局设置按钮从 topbar 移除
topbar SHALL 不再提供全局"设置"按钮与工具编号；各工具的设置或管理入口 SHALL 由其详情页自行提供。

#### Scenario: 工具设置迁移到详情页
- **WHEN** 用户在久坐提醒详情页内打开设置
- **THEN** 设置弹窗在详情页内展示提醒/休息时长配置，保存后对下一个工作段生效（行为与现状一致）

### Requirement: 卡片网格扁平排列
dashboard 卡片网格 SHALL 将所有工具的卡片扁平排列为等宽表格，不按工具分组。未保存自定义顺序时，卡片顺序 SHALL 遵循工具注册顺序（久坐提醒在前，倒数日事件随后）；用户保存自定义顺序后，按自定义顺序排列。没有任何卡片的工具 SHALL 不出现在网格中（不渲染占位卡），仅通过顶栏下拉进入。

#### Scenario: 多工具卡片混排
- **WHEN** dashboard 同时存在久坐提醒卡片与多个倒数日事件卡片，且未保存自定义顺序
- **THEN** 所有卡片按注册顺序平铺进同一网格，自动换行，而不是按工具各自竖排

#### Scenario: 顺序不随刷新跳动
- **WHEN** 某个工具按自己的间隔刷新卡片
- **THEN** 该工具卡片只在原位更新，不改变网格中所有卡片的相对顺序

#### Scenario: 空工具不渲染卡片
- **WHEN** 某个工具没有任何卡片
- **THEN** 该工具不出现在卡片网格中，不渲染占位卡；该工具仍可通过顶栏下拉进入详情页

#### Scenario: 占位 key 按失效 key 丢弃
- **WHEN** `card_order.json` 中残留占位卡时代保存的 key（等于工具 id，如 `countdown`）
- **THEN** 加载顺序时该 key 被当作失效 key 忽略，不产生卡片

### Requirement: 网格四列等宽响应式
dashboard 卡片网格 SHALL 以 4 列等宽排列卡片；窗口收窄时 SHALL 降列以保证卡片可读性。

#### Scenario: 常规宽度 4 列
- **WHEN** dashboard 窗口处于常规宽度
- **THEN** 卡片以 4 列等宽网格排列

#### Scenario: 窄窗口降列
- **WHEN** 窗口宽度收窄到断点以下
- **THEN** 网格降为 2 列，更窄时降为 1 列，卡片不出现挤压变形

### Requirement: 卡片跨工具拖拽排序
dashboard 卡片网格 SHALL 允许用户通过拖拽自由调整卡片顺序，顺序可以跨工具（如把倒数日事件卡排到久坐提醒卡之前），且不按工具分组限制。

#### Scenario: 拖拽改变卡片位置
- **WHEN** 用户按住一张卡片拖动到网格中的另一个位置
- **THEN** 卡片移动到目标位置，其余卡片相应让位

#### Scenario: 跨工具排序
- **WHEN** 用户把倒数日事件卡拖到久坐提醒卡之前
- **THEN** 网格按新顺序展示，不再受工具分组的注册顺序限制

#### Scenario: 点击不动仍进入详情页
- **WHEN** 用户点击卡片而未发生拖动
- **THEN** 不触发排序，卡片点击行为保持（进入该工具详情页）

### Requirement: 卡片顺序持久化
dashboard 卡片顺序 SHALL 持久化到本地文件，应用重启后恢复用户排好的顺序；持久化数据中不再存在的卡片 key SHALL 被丢弃。

#### Scenario: 重启后顺序保持
- **WHEN** 用户调整卡片顺序后重启应用
- **THEN** dashboard 按保存的顺序渲染卡片

#### Scenario: 失效 key 被丢弃
- **WHEN** 保存的顺序中包含已不存在的事件（如已删除）的 key
- **THEN** 加载时忽略该 key，剩余卡片按保存顺序排列

### Requirement: 新增卡片出现在末尾
新出现的卡片（如新建倒数日事件）SHALL 出现在网格末尾，而不是插入到用户排好的顺序中间。

#### Scenario: 新建事件卡片追加末尾
- **WHEN** 用户在倒数日工具中新建一个事件
- **THEN** 该事件对应卡片出现在 dashboard 网格末尾，已排好的其他卡片顺序不变

### Requirement: 卡片渲染位置稳定
工具按各自间隔刷新卡片时，SHALL 在原位更新对应卡片内容，不改变网格中所有卡片的顺序。

#### Scenario: 刷新不重排
- **WHEN** 某个工具按自己的间隔刷新卡片
- **THEN** 网格中卡片相对顺序保持不变，仅对应卡片内容更新

#### Scenario: 拖拽期间刷新不打断
- **WHEN** 用户正在拖拽一张卡片而该工具恰好到达刷新时刻
- **THEN** 拖拽不被节点替换打断，拖拽结束后卡片正常就位

### Requirement: 卡片网格紧凑密度
dashboard 卡片网格 SHALL 采用紧凑的间距与内边距（页容器、卡片内边距、网格间距与卡片数值字号较调整前明显减小），同时保持 4 列等宽布局与既有响应式降列行为，使一屏可见的卡片数量增多，且不改变卡片结构、拖拽排序与刷新机制。

#### Scenario: 一屏可见更多卡片
- **WHEN** dashboard 的卡片数量超过单屏可视容量
- **THEN** 网格以更紧凑的间距排布，一屏可见的卡片数量较调整前增多，卡片保持 4 列等宽且可读

#### Scenario: 紧凑化不改变交互行为
- **WHEN** 用户拖拽排序、点击卡片进入详情页、或卡片按各自间隔刷新
- **THEN** 行为与调整前一致，仅间距与字号发生变化
