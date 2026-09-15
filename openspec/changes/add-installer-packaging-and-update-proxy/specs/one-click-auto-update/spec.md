## MODIFIED Requirements

### Requirement: 一键更新下载
「立即更新」被触发后，后端 SHALL 异步下载检查结果中缓存的资产下载地址对应的 `health-tool.exe`，写入 exe 同目录的 `health-tool.exe.part`，下载完成后 SHALL 将 `.part` 重命名为 `health-tool.exe.new`，并在 `.new` 旁写入 `health-tool.exe.new.version`（内容为本次下载的版本号）；下载期间 SHALL 通过事件推送进度（已下载字节、总字节、百分比与阶段）。下载失败 SHALL 推送错误事件并清理 `.part`。配置了代理时，下载地址 SHALL 经代理前缀包装后请求。当前版本号为 `dev` 或运行平台非 Windows 时，SHALL 不发起下载并给出对应提示。

#### Scenario: 下载中推送进度
- **WHEN** 下载进行中
- **THEN** 按已下载/总字节持续推送进度事件，UI 进度条随之更新

#### Scenario: 下载完成落位
- **WHEN** 下载成功完成
- **THEN** `health-tool.exe.part` 重命名为 `health-tool.exe.new`，并写入 `health-tool.exe.new.version` 记录版本号，推送完成事件

#### Scenario: 下载失败可重试
- **WHEN** 下载过程中网络异常或响应非 200
- **THEN** 推送错误事件、删除 `.part`，UI 显示失败文案且可再次点击更新

#### Scenario: 开发版本拒绝下载
- **WHEN** 当前版本号为 `dev` 且触发下载
- **THEN** 不发起网络请求，提示当前为开发版本不支持自动更新

#### Scenario: 非 Windows 平台拒绝下载
- **WHEN** 运行平台非 Windows 且触发下载
- **THEN** 不发起下载，提示当前平台不支持自动更新

#### Scenario: 经代理下载
- **WHEN** 已配置代理且触发一键更新
- **THEN** 下载请求地址为 `代理前缀 + 资产下载地址`
