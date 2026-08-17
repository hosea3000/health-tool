# CI Autorelease

## Purpose

当仓库推送 `v*` 格式的 git tag 时，自动触发 GitHub Actions 构建 Windows 产物并发布到 GitHub Release，无需人工介入。

## MODIFIED Requirements

### Requirement: 版本号来源于 tag
发布产物的 Windows 文件属性与运行时版本号 SHALL 均取自当前 tag，去除前导 `v` 后使用。

#### Scenario: 文件属性版本注入
- **WHEN** 构建开始时存在 tag `v2.1.0`
- **THEN** `wails.json` 的 `info.productVersion` 被写为 `2.1.0`，且 exe 文件属性显示的版本号为 `2.1.0`

#### Scenario: 运行时版本注入
- **WHEN** 构建开始时存在 tag `v2.1.0`
- **THEN** `wails build` 通过 `-ldflags "-X main.version=2.1.0"` 注入运行时版本号