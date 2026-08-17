## ADDED Requirements

### Requirement: 一键更新下载
「立即更新」被触发后，后端 SHALL 异步下载检查结果中缓存的资产下载地址对应的 `health-tool.exe`，写入 exe 同目录的 `health-tool.exe.part`，下载完成后 SHALL 将 `.part` 重命名为 `health-tool.exe.new`，并在 `.new` 旁写入 `health-tool.exe.new.version`（内容为本次下载的版本号）；下载期间 SHALL 通过事件推送进度（已下载字节、总字节、百分比与阶段）。下载失败 SHALL 推送错误事件并清理 `.part`。当前版本号为 `dev` 或运行平台非 Windows 时，SHALL 不发起下载并给出对应提示。

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

### Requirement: 目录可写探测与降级
一键更新开始前，后端 SHALL 探测 exe 所在目录的可写性（尝试创建临时文件）；目录不可写时 SHALL 中止更新，提示用户通过「前往 GitHub 查看」手动更新。

#### Scenario: 目录可写正常下载
- **WHEN** exe 目录可写且触发更新
- **THEN** 正常开始下载

#### Scenario: 目录不可写降级
- **WHEN** exe 目录不可写（如 Program Files 受保护目录）且触发更新
- **THEN** 中止更新，提示目录不可写并引导前往 GitHub 手动更新

### Requirement: 待更新状态查询
后端 SHALL 提供 `PendingUpdateInfo()` 绑定方法，返回 exe 同目录是否存在 `health-tool.exe.new` 及其版本号（从 `health-tool.exe.new.version` 读取）。设置页加载时 SHALL 查询该方法，`.new` 存在时 SHALL 直接提供「重启更新」入口，不要求先检查更新。`dev` 版本与非 Windows 平台 SHALL 返回不存在。

#### Scenario: 有待应用的更新
- **WHEN** exe 同目录存在 `health-tool.exe.new` 且用户进入设置页
- **THEN** 返回存在且携带版本号，设置页显示「重启更新」入口

#### Scenario: 无待应用的更新
- **WHEN** exe 同目录不存在 `health-tool.exe.new`
- **THEN** 返回不存在，设置页不显示「重启更新」入口

#### Scenario: 开发版本不暴露
- **WHEN** 当前版本号为 `dev`
- **THEN** 返回不存在

### Requirement: 退出-替换-重启
确认重启后，后端 SHALL 在 exe 同目录生成一次性批处理脚本并启动它，随后走完整退出流程（置位退出请求并调用 `runtime.Quit`，触发 shutdown 保存数据后进程退出）；脚本 SHALL 等待应用进程完全消失，将 `health-tool.exe.new` 覆盖为 `health-tool.exe`（`move /y`），启动新版本，删除脚本自身与 `health-tool.exe.new.version`。覆盖失败 SHALL 重试至多 3 次；仍失败时 SHALL 启动旧版本 exe 并保留 `.new` 与 `.new.version`。

#### Scenario: 正常替换并重启
- **WHEN** 进程退出且覆盖成功
- **THEN** 新版本 exe 启动，脚本与 `.new.version` 被删除

#### Scenario: 覆盖失败重试
- **WHEN** 覆盖因文件被短暂锁定等原因失败
- **THEN** 重试覆盖至多 3 次

#### Scenario: 覆盖仍失败启动旧版
- **WHEN** 重试 3 次后覆盖仍失败
- **THEN** 启动旧版本 exe，保留 `.new` 与 `.new.version` 供下次重试

### Requirement: 更新残留清理
应用启动时 SHALL 清理 exe 同目录的 `health-tool.exe.part` 残渣（下载中断遗留）与孤儿 `health-tool.exe.new.version`（`.new` 不存在时的残留）；`health-tool.exe.new` 与配套的 `.new.version` SHALL 保留，供「重启更新」复用。检查更新发现已是最新版本时，SHALL 清理 `.new` 与 `.new.version`；检查失败时 SHALL 保留两者。

#### Scenario: 启动清理下载残渣
- **WHEN** 启动时发现 `health-tool.exe.part`
- **THEN** 删除该文件

#### Scenario: 启动清理孤儿版本标记
- **WHEN** 启动时存在 `health-tool.exe.new.version` 但不存在 `health-tool.exe.new`
- **THEN** 删除 `.new.version` 文件

#### Scenario: 启动保留待确认缓存
- **WHEN** 启动时发现 `health-tool.exe.new` 与 `health-tool.exe.new.version`
- **THEN** 均不删除，保留供「重启更新」使用

#### Scenario: 已是最新时清理待更新
- **WHEN** 检查更新结果为 `up-to-date` 且存在 `.new` 与 `.new.version`
- **THEN** 删除 `.new` 与 `.new.version`

#### Scenario: 检查失败保留待更新
- **WHEN** 检查更新结果为 `error` 且存在 `.new` 与 `.new.version`
- **THEN** 两者均保留

## MODIFIED Requirements

### Requirement: 下载完成确认与重启入口
下载完成（`.new` 就绪）后，UI SHALL 弹出原生确认对话框，文案 SHALL 包含 `.new` 对应的版本号（从 `health-tool.exe.new.version` 读取）；用户确认后 SHALL 触发退出-替换-重启流程，取消后 SHALL 保持应用运行且保留 `.new`，此时「立即更新」按钮 SHALL 变为「重启更新」；点击「重启更新」SHALL 直接弹出确认对话框，不重新下载。检查发现新版本且 `.new` 版本低于最新版本时，SHALL 保留 `.new` 并仅提供「重启更新」入口，不提示新版本信息。

#### Scenario: 确认后重启
- **WHEN** 下载完成且用户在确认对话框中点击确认
- **THEN** 应用保存数据后退出，替换 exe 并启动新版本

#### Scenario: 取消后稍后重启
- **WHEN** 下载完成且用户取消确认对话框
- **THEN** 应用保持运行，`.new` 保留，按钮变为「重启更新」

#### Scenario: 复用已下载缓存
- **WHEN** `.new` 已存在且用户点击「重启更新」
- **THEN** 不重新下载，直接弹出确认对话框，文案使用 `.new.version` 记录的版本号

#### Scenario: 待更新版本落后于最新版本
- **WHEN** 检查发现新版本且 `.new` 对应的版本低于最新版本
- **THEN** 保留 `.new`，反馈区仅显示「重启更新」入口，不显示最新版本与重新下载入口
