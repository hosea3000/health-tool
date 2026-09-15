## ADDED Requirements

### Requirement: 更新代理加速
后端 SHALL 支持为检查更新与下载更新配置代理前缀（gh-proxy 系列加速地址）。代理前缀 SHALL 持久化到 `settings.json` 的 `updateProxy` 字段。配置代理后，检查更新请求 SHALL 优先经代理发起；代理不可用（如代理不支持 `api.github.com`）时，检查更新 SHALL 回退直连且不因此报错。代理为空时 SHALL 直连。代理前缀缺少 scheme 时 SHALL 自动补 `https://`，末尾斜杠 SHALL 被忽略。

#### Scenario: 配置代理后优先经代理请求
- **WHEN** 用户配置了代理前缀并触发检查更新
- **THEN** 检查请求先经 `代理前缀 + GitHub API 地址` 发起

#### Scenario: 代理不可用回退直连
- **WHEN** 用户配置了代理前缀但代理请求失败
- **THEN** 检查更新回退直连 `api.github.com`，按直连结果返回三态

#### Scenario: 未配置代理直连
- **WHEN** `updateProxy` 为空且用户触发检查更新
- **THEN** 直接请求 `https://api.github.com`，不拼接代理前缀

#### Scenario: 代理地址补全 scheme
- **WHEN** 用户填写 `gh-proxy.com`（无 scheme）
- **THEN** 实际请求地址为 `https://gh-proxy.com/<原始 GitHub 地址>`

### Requirement: 设置页网络加速入口
设置页「更新」段下方 SHALL 提供「网络加速」小段，包含代理地址输入框与保存按钮，SHALL 明示其用于加速 GitHub 检查与下载。页面加载时 SHALL 回填已保存的代理地址；保存后 SHALL 提示成功且后续检查/下载立即生效。

#### Scenario: 回填已保存代理
- **WHEN** 用户已保存代理地址后进入设置页
- **THEN** 输入框显示已保存的代理地址

#### Scenario: 保存代理生效
- **WHEN** 用户填写代理地址并点击保存
- **THEN** 设置被持久化，界面提示保存成功，后续检查更新与下载经该代理

#### Scenario: 清空代理直连
- **WHEN** 用户清空代理输入并保存
- **THEN** `updateProxy` 被置空，后续检查与下载直连 GitHub

## MODIFIED Requirements

### Requirement: 手动检查更新
后端 SHALL 提供 `CheckForUpdates()` 绑定方法，请求 `https://api.github.com/repos/hosea3000/health-tool/releases/latest`，解析最新 release 的 `tag_name` 与 `html_url`，返回结构化结果（`up-to-date` / `update-available` / `error` 三态），并携带当前版本号、最新版本号、release 链接与提示文案。配置了代理时，请求 SHALL 优先经代理前缀发起，代理失败时回退直连。

#### Scenario: 发现新版本
- **WHEN** 最新 release 的版本号高于当前版本号
- **THEN** 返回 `update-available`，且结果包含最新版本号与 release 页面链接

#### Scenario: 已是最新版本
- **WHEN** 最新 release 的版本号不高于当前版本号
- **THEN** 返回 `up-to-date`，且结果包含当前版本号

#### Scenario: 检查失败
- **WHEN** 网络异常、超时、响应解析失败，或仓库不存在 release
- **THEN** 返回 `error`，且结果包含人性化失败文案

#### Scenario: 经代理检查
- **WHEN** 已配置代理且代理可用
- **THEN** 请求经代理前缀发起，返回结果与直连一致
