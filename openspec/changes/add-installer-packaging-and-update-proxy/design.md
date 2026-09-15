## Context

`health-tool` 是 Wails v2 桌面应用，发布产物目前只有裸 `health-tool.exe`（CI 在 ubuntu 上交叉编译 `windows/amd64`）。应用内置了「检查更新 + 一键更新」链路：检查 GitHub Release，下载裸 exe 覆盖二进制。参考项目 `workbuddy-checkin` 已把发布形态升级为「installer + 裸 exe」，并给更新链路加了 gh-proxy 系列代理加速，本次对齐其做法。

两个约束决定了设计走向：

1. **一键更新是「覆盖 exe」模型**：要让它工作，exe 所在目录必须对当前用户可写。这直接决定了 installer 必须用 user scope 装到 `%LOCALAPPDATA%`，而不是 `Program Files`。
2. **用户数据与安装目录已分离**：数据在 `os.UserConfigDir()/health-tool/`，installer 不碰它即可天然保留，无需迁移逻辑。

此外 `SaveSettings` 当前是三个位置参数，要加代理字段必须改签名——这是一个有意的对外契约演进，会连带影响 wailsjs 生成物与 reminder 工具的调用点。

## Goals / Non-Goals

**Goals:**
- 发布双产物：裸 exe（绿色/更新用）+ NSIS installer（首次安装用）。
- installer 用户级安装到可写目录，保证一键更新不被目录权限阻断。
- installer 中文化，卸载前检测应用是否运行。
- 安装/卸载不破坏用户数据。
- 检查更新与下载更新支持可配置代理前缀，代理失败时检查更新回退直连。
- 设置持久化与绑定契约演进到结构体传参。

**Non-Goals:**
- 不做代码签名（`signtool` 仍留注释）。
- 不做自动静默更新（in-place，无需用户确认）——仍保留「下载 → 确认 → 重启」交互。
- 不做增量/差分更新。
- 不做多平台 installer（仅 Windows/amd64）。
- 不改一键更新的「覆盖 exe」模型为「运行 installer」。

## Decisions

### 决策 1：installer 用 user scope 安装
`wails build -nsis -installscope user` → 装到 `%LOCALAPPDATA%\Programs\health-tool\`。

- **理由**：该目录可写，一键更新的 `dirWritable` 探测通过，更新能直接覆盖 exe。若用默认 `Program Files`，更新会因不可写而降级为「前往 GitHub 手动更新」，整个一键更新功能形同虚设。
- **备选**：默认 admin scope + 一键更新时提权。被否决——提权需要 UAC，体验差且实现复杂。

### 决策 2：一键更新下载裸 exe，不下载 installer
检查更新仍匹配资产名 `health-tool.exe`，替换二进制。

- **理由**：与参考项目一致；更新是「换二进制」而非「重装」，installer 建立的快捷方式、注册表卸载项、WebView2 运行时都已存在，无需重跑。installer 体积大且会触发安装向导。
- **代价**：注册表里「添加/删除程序」显示的版本号会与实际 exe 版本漂移（更新后不刷新）。可接受——功能不受影响，且卸载执行的是目录内当前 exe。

### 决策 3：卸载进程检测按托盘窗口类名
`FindWindow $0 "HealthToolTrayClass" ""`。类名来自 `tray_windows.go:200`。

- **理由**：应用是托盘驻留、关闭主窗口不退出，用户很容易误以为已退出而卸载，导致文件占用删除失败。`FindWindow` 无需额外进程枚举，NSIS 原生可用。
- **注意**：不能照抄参考项目的 `"WorkbuddyCheckinTray"`，类名必须与本项目一致（`HealthToolTrayClass`）。
- **备选**：`tasklist | findstr health-tool.exe`。被否决——托盘类名更精确，且 `StartHidden` 模式下进程名检测与 updater 脚本用的 `tasklist` 重复但无额外收益。

### 决策 4：用户数据目录不纳入卸载清理
卸载只删 WebView2 DataPath（`$AppData\${PRODUCT_EXECUTABLE}`）与安装目录，保留 `%AppData%\health-tool\`。

- **理由**：settings/timeline/countdowns/card_order 是用户资产，重装后应仍在。
- **风险**：卸载不彻底，残留配置。接受——这是「保留用户数据」的明确取舍；需要彻底清理时用户可手动删目录。

### 决策 5：代理前缀实现为纯字符串拼接
`proxiedURL(proxy, rawURL)`：空则原样返回；无 scheme 补 `https://`；去尾斜杠后 `proxy + "/" + rawURL`。

- **理由**：gh-proxy 系列（`gh-proxy.com`、`ghproxy.net` 等）都是「前缀 + 原始完整 URL」的纯代理，无需 URL 解析或参数注入，字符串拼接即最简实现。
- **检查更新回退策略**：`checkForUpdates` 先经代理请求，若结果为 `error` 再直连。部分代理不支持 `api.github.com`（只代理下载），回退保证可用性。下载（`DownloadAndApplyUpdate`）不加回退——代理失败即报错，避免重下。

### 决策 6：SaveSettings 改结构体传参，返回值保持 bool
`func (a *App) SaveSettings(s model.Settings) bool`。

- **理由**：后续再加设置字段不必再动签名；与参考项目风格一致。返回值保持 `bool` 而非 `error`，最小化无关改动（前端现有 `if (await api.saveSettings(...))` 逻辑不用改判断方式）。
- **影响面**：`App.js`/`App.d.ts`（wails build 重生成）、`reminder/index.js:18-23`（含 preview localStorage 分支）、`app_test.go`。
- **与 `package-structure` 的关系**：该 spec 的「绑定方法契约保持不变」是抽取 model/store 重构时的范围约束，非永久冻结。本次是有意的契约演进，不修改该 requirement。

## Risks / Trade-offs

- **注册表版本号漂移** → 已知接受；如需彻底一致可后续让一键更新后回写注册表 DisplayVersion（本次不做）。
- **`FindWindow` 类名依赖托盘实现** → 若未来托盘改为其他库或类名变更，卸载检测会失效。缓解：类名定义集中在 `tray_windows.go`，spec 中已显式记录类名。
- **代理服务不稳定/被墙** → 检查更新有直连回退；下载无回退但用户可清空代理直连。
- **user scope 安装的每用户隔离** → 同一台机器不同用户各自安装、各自数据目录，属预期行为。
- **CI 交叉编译 NSIS** → ubuntu 上需 `apt-get install nsis`；若 runner 缺 makensis 构建会失败而非产出坏 installer（已在 spec 中约定）。

## Migration Plan

1. 后端：`model.Settings` 加 `UpdateProxy` → `updater.go` 加 `proxiedURL` 与代理参数 → `app.go` 改 `SaveSettings` 签名并让检查/下载读代理。
2. 前端：`reminder/index.js` 改结构体调用 → `settings.js` 加「网络加速」段 → 跑 `wails build` 重生成 wailsjs。
3. 打包：改 `project.nsi`（中文 + 卸载检测）→ 改 `.gitignore` → 改 `release.yml`（装 NSIS + `-nsis -installscope user` + 上传 installer）。
4. 验证：`go test ./...`、`cd frontend && npm run build`；本地打 `v*` tag 触发一次 CI 验证 installer 产出。
5. 回滚：installer 相关改动独立于运行时逻辑，去掉 CI 的 `-nsis` 即回退到单产物发布；代理为空时行为与改动前一致。

## Open Questions

无。卸载不提供「清理用户数据」选项、设置页不展示安装方式，均已在决策 4 中明确取舍。
