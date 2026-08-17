## ADDED Requirements

### Requirement: 开机自启开关
系统 SHALL 提供 Windows 开机自启动能力：开启时在 HKCU `Software\Microsoft\Windows\CurrentVersion\Run` 下写入名为 `health-tool` 的注册表值（内容为当前 exe 绝对路径加 `--hidden` 参数，路径含空格时整体加引号）；关闭时删除该值。自启状态的唯一真相源 SHALL 为该注册表值，`settings.json` 不持久化自启字段。

#### Scenario: 开启自启
- **WHEN** 用户在设置页打开开机自启开关
- **THEN** 系统以用户级权限向 HKCU Run key 写入 `health-tool` 值，内容为带引号的当前 exe 路径加 `--hidden`，无需管理员权限

#### Scenario: 关闭自启
- **WHEN** 用户在设置页关闭开机自启开关
- **THEN** 系统删除 HKCU Run key 下的 `health-tool` 值

#### Scenario: 写入失败反馈
- **WHEN** 注册表写入因权限或策略失败
- **THEN** 绑定方法返回错误，设置页 toggle 保持原状态并向用户展示失败提示，不出现"显示已开启但实际未写入"

#### Scenario: 非 Windows 平台
- **WHEN** 应用运行在非 Windows 平台（开发 stub 环境）
- **THEN** 自启操作返回平台不支持错误，设置页隐藏或禁用该开关

### Requirement: 设置页展示自启实际状态
`GetSettings` 的返回值 SHALL 包含 `autoStart` 字段，其值从注册表实时读取（Run key 下存在非空 `health-tool` 值即视为开启），供设置页开关初始化。`SaveSettings` SHALL 忽略该字段。

#### Scenario: 初始化开关状态
- **WHEN** 用户打开设置页
- **THEN** 开关状态与注册表实际状态一致：Run key 存在 `health-tool` 值时为开，否则为关

#### Scenario: 状态与任务管理器禁用不一致（已接受边界）
- **WHEN** 用户在 Windows 任务管理器「启动应用」中禁用了本应用（`StartupApproved` 标记，Run key 值仍存在）
- **THEN** 设置页仍显示"已开启"；系统不感知 `StartupApproved`，此为已接受的边界行为

### Requirement: 静默启动
应用 SHALL 支持 `--hidden` 启动参数：以该参数启动时不显示主窗口，直接进入托盘驻留（输入监控与提醒正常运行）。不带参数启动时行为不变（正常显示主窗口）。

#### Scenario: 开机静默启动
- **WHEN** 系统开机自启以 `--hidden` 参数拉起应用
- **THEN** 主窗口不显示，应用进入托盘驻留状态，久坐提醒与输入监控正常工作

#### Scenario: 手动启动不受影响
- **WHEN** 用户手动双击 exe 启动（不带 `--hidden`）
- **THEN** 主窗口正常显示，行为与现状完全一致

#### Scenario: 自启实例与手动实例撞车
- **WHEN** 自启隐藏实例已在运行，用户再次手动启动 exe
- **THEN** 既有单实例机制唤起已运行实例的主窗口，不产生第二个进程

#### Scenario: 无参数解析纯函数
- **WHEN** 对启动参数列表执行隐藏标记解析
- **THEN** 仅当参数中存在 `--hidden` 时返回真；该解析为纯函数，可单测
