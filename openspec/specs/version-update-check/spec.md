# Version Update Check

## Purpose

为用户提供手动版本更新检查能力：应用运行时持有当前版本号，用户可在设置页触发检查，从 GitHub Release 获取最新版本并与当前版本做语义化比较，三态反馈检查结果，并在有更新时引导打开 GitHub Release 页面。

## Requirements

### Requirement: 运行时版本号
主包 SHALL 提供运行时版本号变量，默认值为 `dev`，并 SHALL 支持在构建时通过 ldflags（`-X main.version`）注入正式版本号。注入的版本号去除前导 `v`（如 tag `v0.1.2` 注入为 `0.1.2`）。

#### Scenario: 构建注入版本
- **WHEN** 使用 `-ldflags "-X main.version=0.1.2"` 构建
- **THEN** 运行时当前版本号为 `0.1.2`

#### Scenario: 未注入保持默认
- **WHEN** 构建未传 ldflags
- **THEN** 运行时当前版本号为 `dev`

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

### Requirement: 语义化版本比较
版本比较 SHALL 按主版本、次版本、补丁版本三段数字逐位比较；比较前双方 SHALL 剥离前导 `v`。pre-release 后缀（非纯数字段）SHALL 视为低于对应的正式版本。无法解析的版本号 SHALL 被容错处理，不作为更新。

#### Scenario: 补丁版本更新
- **WHEN** 当前为 `0.1.1`，最新 release 为 `0.1.2`
- **THEN** 判定为有更新

#### Scenario: 带 v 前缀的比较
- **WHEN** 当前为 `0.1.1`，最新 release 的 `tag_name` 为 `v0.1.2`
- **THEN** 剥离前缀后判定为有更新

#### Scenario: 预发布版本不高于对应正式版本
- **WHEN** 当前为 `0.2.0`，最新 release 的 `tag_name` 为 `0.2.0-beta`
- **THEN** 判定为无更新（pre-release 视为低于对应的正式版本）

#### Scenario: 非法版本号容错
- **WHEN** 最新 release 的 `tag_name` 无法解析为版本号
- **THEN** 判定为无更新，不产生错误

### Requirement: dev 版本跳过检查
当前版本号为 `dev` 时，检查 SHALL 不发起网络请求，直接返回 `up-to-date` 并提示当前为开发版本。

#### Scenario: 开发版本点击检查
- **WHEN** 当前版本号为 `dev` 且用户触发检查
- **THEN** 不发起网络请求，返回 `up-to-date`，文案说明当前为开发版本

### Requirement: 设置页入口
dashboard 顶栏「健康工具箱」旁 SHALL 提供设置入口（齿轮图标），点击 SHALL 打开设置页；设置页 SHALL 展示应用名称与当前版本号。

#### Scenario: 进入设置页
- **WHEN** 用户在 dashboard 点击设置图标
- **THEN** 视图切换到设置页，显示应用名称与当前版本号

#### Scenario: 返回 dashboard
- **WHEN** 用户在设置页点击返回
- **THEN** 视图回到 dashboard

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