# Work-Rest Timeline

## ADDED Requirements

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

## MODIFIED Requirements

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
