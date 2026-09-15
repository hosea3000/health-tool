## 1. 后端：设置契约与代理加速

- [x] 1.1 `model/model.go`：`Settings` 新增 `UpdateProxy string \`json:"updateProxy"\`` 字段，确认 `DefaultSettings` 与 `LoadSettings` 对缺失字段容错
- [x] 1.2 `updater.go`：新增 `proxiedURL(proxy, rawURL string) string`（空代理原样返回、补 scheme、去尾斜杠拼接）
- [x] 1.3 `updater.go`：`checkForUpdates` 增加 `proxy` 参数，代理非空时先经代理请求，结果为 `error` 时回退直连
- [x] 1.4 `app.go`：`CheckForUpdates` 从 `a.settings.UpdateProxy` 读代理并传入 `checkForUpdates`
- [x] 1.5 `app.go`：`DownloadAndApplyUpdate` 下载地址经 `proxiedURL` 包装后再请求
- [x] 1.6 `app.go`：`SaveSettings` 签名改为 `SaveSettings(s model.Settings) bool`，内部同步 monitor 时长/通知设置并持久化
- [x] 1.7 `updater_test.go`：补充 `proxiedURL` 单测（无 scheme、带尾斜杠、空代理）与 `checkForUpdates` 代理回退用例
- [x] 1.8 `app_test.go`：适配 `SaveSettings` 新签名，验证 `UpdateProxy` 持久化与 `settings.json` 往返

## 2. 前端：调用点与设置页

- [x] 2.1 `frontend/src/tools/reminder/index.js`：`api.saveSettings` 与 `backendSave` 改为传 `model.Settings` 结构体，同步更新 preview localStorage 分支
- [x] 2.2 `frontend/src/views/settings.js`：在「更新」段下方新增「网络加速」小段（代理输入框 + 保存按钮 + 成功/失败提示）
- [x] 2.3 `settings.js`：进入设置页时调 `GetSettings` 回填 `updateProxy`，保存走 `SaveSettings` 结构体
- [x] 2.4 `settings.js`：为无 Wails bridge 的 preview 模式提供 localStorage 回退
- [x] 2.5 运行 `wails build`（或 `wails dev`）重新生成 `frontend/wailsjs/go/main/App.{js,d.ts}`，确认 `SaveSettings` 反映结构体参数
- [x] 2.6 `cd frontend && npm run build` 确认前端构建通过

## 3. 打包：NSIS installer

- [x] 3.1 `build/windows/installer/project.nsi`：`MUI_LANGUAGE English` 改为 `SimpChinese`，加 `!include "LogicLib.nsh"`
- [x] 3.2 `project.nsi`：uninstall section 开头加 `FindWindow $0 "HealthToolTrayClass" ""` 进程检测，运行中弹中文提示并 `Quit`
- [x] 3.3 `project.nsi`：核对卸载段删除项——删除 WebView2 DataPath 与安装目录，SHALL NOT 删除 `%AppData%\health-tool\`
- [x] 3.4 `.gitignore`：忽略 `build/windows/installer/wails_tools.nsh` 与 `build/windows/installer/tmp/`

## 4. CI：双产物发布

- [x] 4.1 `.github/workflows/release.yml`：新增 `Install NSIS` 步骤（`sudo apt-get update && sudo apt-get install -y nsis`）
- [x] 4.2 `release.yml`：`wails build` 命令加 `-nsis -installscope user`
- [x] 4.3 `release.yml`：`gh release create` 追加 `build/bin/health-tool-amd64-installer.exe`
- [ ] 4.4 打一个 `v*` 测试 tag 触发流水线，确认 Release 同时包含 `health-tool.exe` 与 `health-tool-amd64-installer.exe`

## 5. 验证

- [x] 5.1 `go test ./...` 全绿
- [ ] 5.2 Windows 上安装 installer：确认装到 `%LOCALAPPDATA%\Programs\health-tool\`、快捷方式可用、无 UAC 提权
- [ ] 5.3 Windows 上验证代理：配置代理后检查更新/下载经代理，清空代理后直连，代理错误时检查更新回退直连
- [ ] 5.4 Windows 上验证卸载：应用运行中拒绝卸载，退出后可卸载，卸载后 `%AppData%\health-tool\` 数据保留
- [ ] 5.5 验证共存：通过 installer 安装后触发一键更新到新版本，确认 exe 替换、快捷方式与卸载项仍可用
