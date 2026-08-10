## ADDED Requirements

### Requirement: Record work and pause intervals
系统 SHALL 在有效活动创建工作段、工作段因闲置或提醒结束、以及提醒休息期开始或结束时，记录对应的工作或暂停区间；待工作状态不得生成记录。

#### Scenario: Effective activity starts a work interval
- **WHEN** 系统处于待工作、闲置暂停或提醒休息期结束后的待工作状态，并收到有效活动
- **THEN** 系统创建一条从该有效活动时间开始的 `working` 记录

#### Scenario: Idle pause closes work
- **WHEN** 工作段连续 5 分钟没有有效活动并进入闲置暂停
- **THEN** 系统结束当前 `working` 记录，并创建从暂停时间开始的 `idle-paused` 记录

#### Scenario: Reminder starts rest
- **WHEN** 工作段达到提醒时长并进入提醒休息期
- **THEN** 系统结束当前 `working` 记录，并创建从提醒时间开始的 `resting` 记录

#### Scenario: Rest ends without a new work segment
- **WHEN** 提醒休息期倒计时结束且用户尚未产生有效活动
- **THEN** 系统结束当前 `resting` 记录，且不创建新的记录

### Requirement: Expose the current-session timeline
系统 SHALL 提供只读时间线查询，返回当前应用运行周期的记录，按开始时间从早到晚排列；进行中的记录 SHALL 标识为未结束，并提供截至查询时间的持续秒数。

#### Scenario: Query an empty timeline
- **WHEN** 应用启动后尚未产生有效活动且调用时间线查询
- **THEN** 系统返回空列表

#### Scenario: Query ordered records
- **WHEN** 当前运行周期已有多个工作或暂停区间并调用时间线查询
- **THEN** 系统返回按 `startedAt` 升序排列的记录，每条记录包含类型、开始时间、结束时间和持续秒数

#### Scenario: Query an ongoing record
- **WHEN** 当前仍处于工作或暂停状态并调用时间线查询
- **THEN** 当前记录的 `endedAt` 为空，持续秒数按查询时刻计算且不得为负数

### Requirement: Display the timeline in the main window
前端 SHALL 在主界面以垂直时间线展示后端返回的记录，区分工作与暂停类型，显示开始时间、结束时间或“进行中”以及持续时长。

#### Scenario: Render records
- **WHEN** 时间线查询返回一条或多条记录
- **THEN** 主界面按返回顺序显示每条记录的类型、时间范围和持续时长

#### Scenario: Render an empty state
- **WHEN** 时间线查询返回空列表
- **THEN** 主界面显示说明尚未产生工作或休息记录的空状态

#### Scenario: Refresh ongoing duration
- **WHEN** 应用仍在运行且当前记录未结束
- **THEN** 前端沿用现有定时刷新，使当前记录的持续时长更新而无需用户手动操作
