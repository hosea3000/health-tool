## ADDED Requirements

### Requirement: installer 发布形态
发布流水线 SHALL 为 Windows/amd64 目标同时产出「裸 exe」与「NSIS installer」两个产物，installer 文件名格式为 `<项目名>-amd64-installer.exe`。installer MUST 由 `wails build -nsis` 生成，SHALL NOT 手工运行 makensis 拼装。

#### Scenario: 构建产出双产物
- **WHEN** 流水线执行 `wails build -clean -nsis -installscope user -platform windows/amd64`
- **THEN** `build/bin/` 下同时存在 `health-tool.exe` 与 `health-tool-amd64-installer.exe`

#### Scenario: installer 与 exe 同版本
- **WHEN** 发布 tag 为 `v0.2.0`
- **THEN** installer 与 exe 的文件属性版本号均为 `0.2.0`

### Requirement: installer 用户级安装
installer SHALL 使用 user 安装作用域（`-installscope user`），默认安装目录为 `%LOCALAPPDATA%\Programs\health-tool\`，SHALL NOT 安装到 `Program Files` 等受保护目录。安装目录 MUST 对当前用户可写。

#### Scenario: 默认安装路径
- **WHEN** 用户运行 installer 且未修改安装目录
- **THEN** 应用被安装到 `%LOCALAPPDATA%\Programs\health-tool\`

#### Scenario: 安装目录可写
- **WHEN** 应用安装完成后启动
- **THEN** 应用所在目录对当前用户可写，一键更新的目录可写探测通过

#### Scenario: 安装不要求管理员权限
- **WHEN** 普通用户（非管理员）运行 installer
- **THEN** 安装流程正常完成，不弹出 UAC 提权

### Requirement: installer 中文化
installer 界面语言 SHALL 为简体中文（`MUI_LANGUAGE SimpChinese`）。

#### Scenario: 中文安装向导
- **WHEN** 用户运行 installer
- **THEN** 欢迎页、目录选择页、安装进度页、完成页与卸载页均以简体中文展示

### Requirement: 卸载前进程检测
卸载流程 SHALL 在删除文件前检测应用是否仍在运行，检测方式为按托盘窗口类名 `HealthToolTrayClass` 查找顶层窗口。应用运行中时 SHALL 中止卸载并提示用户先退出应用（右键托盘图标 → 退出），SHALL NOT 强制删除正在使用的文件。

#### Scenario: 应用运行中拒绝卸载
- **WHEN** 应用正在运行（托盘窗口已创建）且用户执行卸载
- **THEN** 卸载中止，弹出提示要求先退出应用，安装目录文件保持完整

#### Scenario: 应用已退出允许卸载
- **WHEN** 应用未运行且用户执行卸载
- **THEN** 卸载正常完成，安装目录被删除

### Requirement: 用户数据与安装解耦
installer 的安装与卸载流程 SHALL NOT 删除用户数据目录 `os.UserConfigDir()/health-tool/`（含 `settings.json`、`timeline.json`、`countdowns.json`、`card_order.json`）；卸载 SHALL 仅允许清理 WebView2 运行时数据目录与应用安装目录。

#### Scenario: 升级安装保留数据
- **WHEN** 用户已使用应用产生数据后，运行新版本 installer 覆盖安装
- **THEN** `settings.json`、`timeline.json`、`countdowns.json`、`card_order.json` 均保留，应用启动后加载到原数据

#### Scenario: 卸载后数据仍在
- **WHEN** 用户执行卸载
- **THEN** 安装目录与快捷方式被移除，`os.UserConfigDir()/health-tool/` 下的用户数据文件保留

### Requirement: 一键更新与 installer 共存
一键更新 SHALL 始终下载并替换裸 `health-tool.exe` 二进制，SHALL NOT 下载或运行 installer。更新替换后的应用 SHALL 继续使用 installer 建立的安装目录、快捷方式与注册表卸载项，不因更新而失效。

#### Scenario: installer 安装后一键更新
- **WHEN** 用户通过 installer 安装应用，随后触发一键更新并重启
- **THEN** 安装目录下的 `health-tool.exe` 被替换为新版本，开始菜单与桌面快捷方式、注册表卸载项仍指向该目录并可用

#### Scenario: 更新不下载 installer
- **WHEN** 一键更新下载资产
- **THEN** 下载的是 `health-tool.exe`，SHALL NOT 下载 `health-tool-amd64-installer.exe`

#### Scenario: 更新后卸载项仍可卸载
- **WHEN** 用户一键更新到新版本后执行卸载
- **THEN** 卸载流程正常完成，删除的是更新后的当前版本

### Requirement: 设置持久化契约
`App.SaveSettings` SHALL 接收单个 `model.Settings` 结构体参数，返回值 SHALL 为 `bool`（成功为 true）。`model.Settings` SHALL 包含 `reminderMinutes`、`restMinutes`、`notificationsEnabled` 与新增的 `updateProxy` 字段。新增字段 SHALL 对旧配置文件向后兼容（缺失时取零值/空串）。

#### Scenario: 结构体传参保存设置
- **WHEN** 前端以 `{reminderMinutes, restMinutes, notificationsEnabled, updateProxy}` 调用 `SaveSettings`
- **THEN** 返回 true 且全部字段被持久化到 `settings.json`

#### Scenario: 旧配置加载取默认代理
- **WHEN** 加载不含 `updateProxy` 字段的旧 `settings.json`
- **THEN** 解析不报错，`UpdateProxy` 取空串（不启用代理）

#### Scenario: wailsjs 绑定同步生成
- **WHEN** 修改 `SaveSettings` 签名后重新执行 `wails build`
- **THEN** `frontend/wailsjs/go/main/App.js` 与 `App.d.ts` 中 `SaveSettings` 反映结构体参数形态
