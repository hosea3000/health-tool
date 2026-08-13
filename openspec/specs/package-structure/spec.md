# Package Structure

## Purpose

约束代码库的分层结构：将数据契约类型与持久化逻辑从 `package main` 抽取为独立的 `model` 与 `store` 包，建立 `main → store → model → domain` 的单向无环依赖，同时保证重构前后对外行为与持久化格式不变。

## Requirements

### Requirement: 绑定方法契约保持不变
重构后 `App` 的导出绑定方法（`Status`、`Timeline`、`GetSettings`、`SaveSettings`、`CountdownEvents`、`AddCountdown`、`UpdateCountdown`、`DeleteCountdown`、`GetCardOrder`、`SaveCardOrder`）的签名语义与返回结构的 JSON 字段名 MUST 与重构前完全一致。

#### Scenario: wailsjs 生成物接口不变
- **WHEN** 重构完成后重新生成 `frontend/wailsjs`
- **THEN** 前端调用这些方法的参数名与返回字段名与重构前一致

### Requirement: 持久化格式与位置不变
用户数据文件（`settings.json`、`timeline.json`、`countdowns.json`、`card_order.json`）的 JSON schema 与存储路径（`os.UserConfigDir()/health-tool/`）MUST 保持与重构前完全一致。

#### Scenario: 旧数据文件可正常加载
- **WHEN** 应用加载重构前已写入的任一用户数据文件
- **THEN** 数据被正确解析，无需迁移

### Requirement: 包依赖无环
`main`、`model`、`store`、`domain` 包之间的依赖方向 MUST 为单向无环：`main → store → model → domain`，且 `main → domain`。

#### Scenario: 编译无 import cycle
- **WHEN** 在重构后的代码上运行 `go build ./...`
- **THEN** 编译通过且无 import cycle 错误
