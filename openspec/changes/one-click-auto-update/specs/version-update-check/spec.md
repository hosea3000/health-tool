## MODIFIED Requirements

### Requirement: 手动检查更新
后端 SHALL 提供 `CheckForUpdates()` 绑定方法，请求 `https://api.github.com/repos/hosea3000/health-tool/releases/latest`，解析最新 release 的 `tag_name`、`html_url` 与资产列表，从资产中按文件名 `health-tool.exe` 匹配 `browser_download_url` 作为下载地址，返回结构化结果（`up-to-date` / `update-available` / `error` 三态），并携带当前版本号、最新版本号、release 链接、资产下载地址与提示文案。资产下载地址 SHALL 随结果缓存，供一键更新直接使用。

#### Scenario: 发现新版本
- **WHEN** 最新 release 的版本号高于当前版本号
- **THEN** 返回 `update-available`，且结果包含最新版本号、release 页面链接与 `health-tool.exe` 资产的下载地址

#### Scenario: 新版本缺少资产
- **WHEN** 最新 release 的版本号高于当前版本号，但资产列表中不存在 `health-tool.exe`
- **THEN** 返回 `update-available`，结果中下载地址为空，不提供一键更新入口

#### Scenario: 已是最新版本
- **WHEN** 最新 release 的版本号不高于当前版本号
- **THEN** 返回 `up-to-date`，且结果包含当前版本号

#### Scenario: 检查失败
- **WHEN** 网络异常、超时、响应解析失败，或仓库不存在 release
- **THEN** 返回 `error`，且结果包含人性化失败文案

### Requirement: 检查结果反馈与跳转
设置页更新区 SHALL 提供「检查更新」按钮，按钮触发检查期间 SHALL 进入加载态防止重复点击；检查完成后按三态结果展示反馈：`up-to-date` 显示"已是最新版本"及当前版本号，`update-available` 显示最新版本号并提供「立即更新」按钮与「前往 GitHub 查看」按钮，`error` 显示失败文案。点击「前往 GitHub 查看」SHALL 在系统默认浏览器打开 release 页面；点击「立即更新」SHALL 触发一键更新下载流程。

#### Scenario: 显示已是最新
- **WHEN** 检查结果为 `up-to-date`
- **THEN** 反馈区显示"已是最新版本"与当前版本号，不显示更新与跳转按钮

#### Scenario: 显示新版本并提供更新入口
- **WHEN** 检查结果为 `update-available` 且下载地址非空
- **THEN** 反馈区显示最新版本号，并提供「立即更新」与「前往 GitHub 查看」按钮

#### Scenario: 新版本无下载地址仅跳转
- **WHEN** 检查结果为 `update-available` 但下载地址为空
- **THEN** 反馈区显示最新版本号，仅提供「前往 GitHub 查看」按钮，不提供「立即更新」

#### Scenario: 检查失败提示
- **WHEN** 检查结果为 `error`
- **THEN** 反馈区显示失败文案，用户可再次点击检查
