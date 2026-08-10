# Single-Instance Window Activation

## Purpose

确保应用在同一用户会话中只运行一个实例，并在重复启动时唤醒已有主窗口，同时保持监控、提醒和当前工作状态不变。

## Requirements

### Requirement: Allow only one running application instance

应用 SHALL 使用稳定的应用标识保证同一用户会话中只有一个应用实例继续运行；当已有实例存在时，后续启动进程 MUST 在不启动独立监控和提醒循环的情况下退出。

#### Scenario: First launch acquires the instance

- **WHEN** 当前用户会话中没有正在运行的应用实例并启动应用
- **THEN** 应用获取单实例锁并正常启动主窗口、输入监控和提醒循环

#### Scenario: Second launch does not create a second monitor

- **WHEN** 已有应用实例正在运行并再次启动应用
- **THEN** 后续进程退出，且系统中仍只有已有实例执行输入监控和提醒循环

### Requirement: Activate the existing main window

当后续启动被单实例锁拦截时，已有实例 SHALL 显示主窗口；如果窗口处于最小化状态，已有实例 SHALL 先恢复窗口再显示。

#### Scenario: Existing window is hidden

- **WHEN** 已有实例因关闭主窗口而驻留托盘，用户再次启动应用
- **THEN** 已有实例显示前一个主界面，且原有监控状态保持不变

#### Scenario: Existing window is minimised

- **WHEN** 已有实例主窗口处于最小化状态，用户再次启动应用
- **THEN** 已有实例取消最小化并显示前一个主界面

#### Scenario: Existing window is already visible

- **WHEN** 已有实例主窗口已经可见，用户再次启动应用
- **THEN** 已有实例保持单实例运行并将主窗口带到用户可见的位置，不创建新窗口

### Requirement: Preserve existing application state

唤醒已有实例 SHALL 不重建 `App`、`domain.Monitor`、输入监控、提醒计时器或当前时间线记录。

#### Scenario: Wake preserves active work segment

- **WHEN** 已有实例正在记录工作段，用户再次启动应用并唤醒主窗口
- **THEN** 主界面显示同一工作段状态和计时进度，且不会重置工作段
