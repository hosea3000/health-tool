# Work-Rest Timeline

## Purpose

记录当前应用运行周期内的工作与暂停区间，提供只读的有序查询结果，并在主界面以时间线形式展示这些记录及其持续时间。
## Requirements
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
原需求范围由"当前应用运行周期"扩展为"当天"，系统 SHALL 仍提供只读时间线查询，按开始时间从早到晚返回当天记录；进行中的记录 SHALL 标识为未结束，并提供截至查询时间的持续秒数。

#### Scenario: Query an empty timeline
- **WHEN** 应用启动后（包括新的一天）尚未产生有效活动且调用时间线查询
- **THEN** 系统返回空列表

#### Scenario: Query ordered records
- **WHEN** 当前当天已有多个工作或暂停区间并调用时间线查询
- **THEN** 系统返回按 `startedAt` 升序排列的记录，每条记录包含类型、开始时间、结束时间和持续秒数

#### Scenario: Query an ongoing record
- **WHEN** 当前仍处于工作或暂停状态并调用时间线查询
- **THEN** 当前记录的 `endedAt` 为空，持续秒数按查询时刻计算且不得为负数

#### Scenario: Query records after restart
- **WHEN** 应用在同一天内重启且当天已有持久化记录
- **THEN** 系统返回当天全部记录，按 `startedAt` 升序排列，进行中记录以查询时刻计算持续秒数且不得为负数

### Requirement: Display the timeline in the main window
前端 SHALL 在久坐提醒工具详情页以垂直时间线展示后端返回的记录，区分工作与暂停类型，显示开始时间、结束时间或“进行中”以及持续时长。

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

### Requirement: Aggregate and display the total work duration
前端 SHALL 对时间线中 `working` 类型记录的持续秒数求和，并在主界面时间线标题行以 `xxx小时xx分xx秒` 恒三位格式展示累计工作时长；口径不得包含 `resting` 或 `idle-paused` 记录。

#### Scenario: Show a running total while working
- **WHEN** 时间线包含一条或多条 `working` 记录（含进行中记录）
- **THEN** 主界面时间线标题行显示各 `working` 记录持续秒数之和，格式为 `xxx小时xx分xx秒`，且随现有每秒刷新更新

#### Scenario: Show zero before any record
- **WHEN** 时间线查询返回空列表
- **THEN** 主界面时间线标题行显示 `0小时0分0秒`

#### Scenario: Exclude non-work records
- **WHEN** 时间线同时包含 `working`、`resting` 与 `idle-paused` 记录
- **THEN** 累计工作时长只统计 `working` 记录，`resting` 与 `idle-paused` 记录不计入

### Requirement: Persist the current day's timeline
系统 SHALL 将当天的时间线记录持久化为本地 JSON 文件，文件包含记录所属日期与最后写入时间；记录仅在状态转换时写入，不得在每次轮询时写入。

#### Scenario: Persist on state transition
- **WHEN** 状态发生转换（工作段开始、闲置暂停、提醒休息期开始或结束）
- **THEN** 系统将该时刻的时间线记录写入本地 JSON 文件，文件包含当天日期与记录的最后写入时间

#### Scenario: Preserve records across restart on the same day
- **WHEN** 应用重启且持久化文件的日期与当天一致
- **THEN** 系统恢复该文件中的记录，`Timeline()` 继续返回这些记录

#### Scenario: Close an open record on graceful exit
- **WHEN** 应用退出时仍存在进行中的记录
- **THEN** 系统以退出时刻闭合该记录并写入文件

#### Scenario: Close an open record recovered from a crash
- **WHEN** 应用启动时持久化文件中存在未闭合（`endedAt` 为空）的记录
- **THEN** 系统以文件中的最后写入时间闭合该记录

### Requirement: Start a fresh day when the persisted date differs
当持久化文件的日期与当天不一致时，系统 SHALL 将该文件视为过期数据，从空记录开始；应用运行期间跨过午夜时 SHALL 清空当前记录并从新的一天重新计数。

#### Scenario: Restart on a new day
- **WHEN** 应用重启且持久化文件的日期早于当天
- **THEN** 系统忽略旧文件中的记录，`Timeline()` 返回空列表

#### Scenario: Roll over at midnight while running
- **WHEN** 应用持续运行并跨过午夜
- **THEN** 系统清空当前时间线并从新的一天开始记录，持久化文件同步更新

