## MODIFIED Requirements

### Requirement: Display the timeline in the main window
前端 SHALL 在久坐提醒工具详情页以垂直时间线展示后端返回的记录，区分工作与暂停类型，显示开始时间、结束时间或"进行中"以及持续时长。

#### Scenario: Render records
- **WHEN** 时间线查询返回一条或多条记录
- **THEN** 久坐提醒工具详情页按返回顺序显示每条记录的类型、时间范围和持续时长

#### Scenario: Render an empty state
- **WHEN** 时间线查询返回空列表
- **THEN** 久坐提醒工具详情页显示说明尚未产生工作或休息记录的空状态

#### Scenario: Refresh ongoing duration
- **WHEN** 应用仍在运行且当前记录未结束
- **THEN** 前端沿用现有定时刷新，使当前记录的持续时长更新而无需用户手动操作

#### Scenario: Timeline not shown on the dashboard
- **WHEN** 用户停留在 dashboard 首页
- **THEN** 时间线不显示在 dashboard 上，仅在进入久坐提醒工具详情页后展示
