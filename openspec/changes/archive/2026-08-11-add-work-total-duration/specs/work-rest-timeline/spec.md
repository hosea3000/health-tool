# Work-Rest Timeline

## ADDED Requirements

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
