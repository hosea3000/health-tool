# Work-Rest Timeline 修改

## MODIFIED Requirements

### Requirement: Record work and pause intervals
系统 SHALL 在有效活动创建工作段、工作段因闲置或提醒结束、以及提醒休息期开始或结束时，记录对应的工作或暂停区间；待工作状态不得生成记录。当久坐提醒通知开关关闭（静默记录模式）时，工作段达到提醒时长 SHALL NOT 结束当前工作段、SHALL NOT 创建 `resting` 记录，该工作段 SHALL 持续记录直至真实的状态转换（闲置暂停或退出）发生。

#### Scenario: Effective activity starts a work interval
- **WHEN** 系统处于待工作、闲置暂停或提醒休息期结束后的待工作状态，并收到有效活动
- **THEN** 系统创建一条从该有效活动时间开始的 `working` 记录

#### Scenario: Idle pause closes work
- **WHEN** 工作段连续 5 分钟没有有效活动并进入闲置暂停
- **THEN** 系统结束当前 `working` 记录，并创建从暂停时间开始的 `idle-paused` 记录

#### Scenario: Reminder starts rest
- **WHEN** 久坐提醒通知开关开启且工作段达到提醒时长并进入提醒休息期
- **THEN** 系统结束当前 `working` 记录，并创建从提醒时间开始的 `resting` 记录

#### Scenario: Silent mode keeps work interval open
- **WHEN** 久坐提醒通知开关关闭且工作段达到提醒时长
- **THEN** 系统不结束当前 `working` 记录、不创建 `resting` 记录，该记录以进行中状态持续累计

#### Scenario: Rest ends without a new work segment
- **WHEN** 提醒休息期倒计时结束且用户尚未产生有效活动
- **THEN** 系统结束当前 `resting` 记录，且不创建新的记录
