## ADDED Requirements

### Requirement: Tag 触发自动发布
当仓库推送 `v*` 格式的 git tag 时，GitHub Actions SHALL 自动执行构建与发布流程，无需人工介入。

#### Scenario: 推送新版本 tag
- **WHEN** 推送 `v1.2.3` 格式的 tag
- **THEN** 流水线在 `ubuntu-latest` runner 上交叉编译 Windows 产物，并以该 tag 作为版本号发布到 GitHub Release

#### Scenario: 推送非版本 tag
- **WHEN** 推送的 tag 不以 `v` 开头
- **THEN** SHALL NOT 触发发布流水线

#### Scenario: 分支 push 不触发
- **WHEN** 向任何分支（含 main）推送提交
- **THEN** SHALL NOT 触发发布流水线

### Requirement: 版本号来源于 tag
发布产物的 Windows 文件属性 SHALL 取自当前 tag，去除前导 `v` 后使用。

#### Scenario: 版本注入
- **WHEN** 构建开始时存在 tag `v2.1.0`
- **THEN** `wails.json` 的 `info.productVersion` 被写为 `2.1.0`，且 exe 文件属性显示的版本号为 `2.1.0`

### Requirement: 发布产物完整性
tag 触发的每个 Release SHALL 包含以下资产：exe；changelog SHALL 由 GitHub 原生生成（`gh release create --generate-notes`）。

#### Scenario: 发布成功
- **WHEN** `gh release create` 完成发布
- **THEN** Release 中包含 `health-tool.exe`，并显示 GitHub 生成的 changelog

### Requirement: 构建产物必须由 wails build 生成
发布所用的二进制 MUST 由 `wails build` 产出（含前端编译与资源注入），SHALL NOT 以原始 `go build` 替代。

#### Scenario: 构建流程
- **WHEN** 流水线执行构建步骤
- **THEN** 实际执行命令为 `wails build -clean`，产物输出到 `build/bin/`

### Requirement: 权限要求
工作流 SHALL 声明足够的 `contents: write` 权限以创建 GitHub Release。

#### Scenario: 发布授权
- **WHEN** `gh release create` 尝试创建 Release 与上传资产
- **THEN** 使用的 token 具有 `contents: write` 权限
