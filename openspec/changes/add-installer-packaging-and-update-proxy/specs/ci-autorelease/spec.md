## MODIFIED Requirements

### Requirement: 发布产物完整性
tag 触发的每个 Release SHALL 包含以下资产：裸 exe 与 NSIS installer；changelog SHALL 由 GitHub 原生生成（`gh release create --generate-notes`）。

#### Scenario: 发布成功
- **WHEN** `gh release create` 完成发布
- **THEN** Release 中包含 `health-tool.exe` 与 `health-tool-amd64-installer.exe`，并显示 GitHub 生成的 changelog

#### Scenario: installer 资产存在
- **WHEN** 用户在 Release 页面查看资产列表
- **THEN** 可下载 installer 用于首次安装，也可下载裸 exe 用于绿色使用或一键更新

### Requirement: 构建产物必须由 wails build 生成
发布所用的二进制 MUST 由 `wails build` 产出（含前端编译与资源注入），SHALL NOT 以原始 `go build` 替代。构建 Windows installer 时，流水线 SHALL 预先在 runner 上安装 NSIS，并 SHALL 向 `wails build` 传入 `-nsis -installscope user`。

#### Scenario: 构建流程
- **WHEN** 流水线执行构建步骤
- **THEN** 实际执行命令为 `wails build -clean -nsis -installscope user -platform windows/amd64`，产物输出到 `build/bin/`

#### Scenario: NSIS 前置安装
- **WHEN** 流水线在 `ubuntu-latest` runner 上准备构建环境
- **THEN** 构建前已通过包管理器（`apt-get install nsis`）安装 NSIS，`makensis` 可用

#### Scenario: 缺少 NSIS 时不发布 installer
- **WHEN** runner 上未安装 NSIS 却传入 `-nsis`
- **THEN** 构建失败，不产出不可用的 installer 资产
