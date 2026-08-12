## ADDED Requirements

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
