## Context

「检查更新」能力已上线:CI 打 `v*` tag 自动构建 `health-tool.exe` 并发布到 GitHub release(资产名为 `health-tool.exe`);应用内 `CheckForUpdates()` 请求 `releases/latest`,三态反馈,`update-available` 时跳转 GitHub 手动下载。缺口是「下载 → 替换 → 重启」闭环:用户目前需手动下载、退出应用、覆盖 exe、重新启动。

本项目的三个特性直接影响设计:

- **关窗 ≠ 退出**:`beforeClose` 隐藏窗口,输入监控/托盘/提醒继续运行。更新链路必须走 `requestQuit` + `runtime.Quit` 真正退出。
- **单实例锁**:`SingleInstanceLock`(UniqueId `health-tool`)。旧进程未退出时启动新进程会被转发给旧进程并自行退出,因此新版本必须在旧进程完全退出后启动。
- **按平台双文件模式**:`input_*` / `tray_*` / `notification_*` / `lock_*` 均为 Windows 实现 + 非 Windows stub;`version` 默认 `dev`,dev 短路已有先例。

## Goals / Non-Goals

**Goals:**

- 界面一键完成「下载 → 确认 → 退出 → 替换 → 重启」,替换过程零用户操作。
- 下载进度实时反馈;下载完成后弹确认窗,取消后可稍后重启。
- 任何失败路径下应用都可用:替换失败启动旧版、目录不可写降级跳转 GitHub。
- 零新增第三方依赖,沿用 Windows/stub 双文件模式与现有 Quit 链路。

**Non-Goals:**

- 不处理 UAC 提权(Program Files 等不可写目录直接降级,不弹提权)。
- 不保留旧版本备份、不做自动回滚。
- 不做后台静默预下载、不自动检查(仍为手动点击触发)。
- 不做 exe 代码签名、sha256 校验(HTTPS + GitHub 为可信源,校验列为可选项)。
- 非 Windows 平台不提供一键更新。

## Decisions

### 1. 替换机制:进程退出后由批处理脚本 `move /y` 一步覆盖,不保留 .old

替代方案对比:

| 方案 | 结论 |
|---|---|
| 进程内 rename 两步走(exe→.old,.new→exe),保留 .old | 被否决:需要 .old 管理与恢复策略,与「不保留 .old」决策冲突 |
| 独立 updater.exe 子进程 | 被否决:需构建第二个二进制,CI 复杂化 |
| 第三方库(minio/selfupdate 等) | 被否决:Windows 替换运行中文件绕不开同一问题,引入依赖不值 |
| **bat 等进程退出后 `move /y .new → exe`** | **采纳** |

关键依据:批处理脚本等待旧进程完全退出后,exe 文件锁与单实例锁均已释放,`move /y` 是单步原子覆盖,任意时刻目录中都存在完整可用的 exe(旧或新),无需 .old 与恢复逻辑。

### 2. 确认时机:下载完成后、退出前,应用内原生对话框

替换必须等进程退出,进程退出后应用无法再弹窗;若由 bat 弹系统窗,应用界面全无、用户不点则永不启动,体验差。因此确认点放在下载完成(即 `.new` 就绪)之后、退出之前:前端调用 Wails 前端 runtime 的 `MessageDialog`(原生对话框,文案含新版本号),确认则调用后端退出-替换-重启方法,取消则保持现状。

### 3. `.part` / `.new` 两段式文件命名

- `health-tool.exe.part`:下载中;存在即中断残渣,启动时清理。
- `health-tool.exe.new`:下载完成后由 `.part` rename 而来;合法的「已下载待确认」缓存,取消后保留。

由此获得「重启更新」语义:再次点击时检测到 `.new` 已存在,跳过下载直接弹确认窗。

### 4. 资产 URL 在 `CheckForUpdates` 时解析并缓存

`githubRelease` 结构扩展 `assets` 字段,按文件名 `health-tool.exe` 匹配 `browser_download_url`,随检查结果一并返回并缓存;点「立即更新」不再二次请求 API,规避 GitHub 未认证 60 次/小时限流,也避免「检查有更新但资产缺失」的不一致。

### 5. 下载进度:Wails events 推送

Go 端 `runtime.EventsEmit(ctx, "update:progress", …)` 推送下载进度(已下载字节/总字节/百分比/阶段),前端 `EventsOn` 订阅更新进度条。下载在 goroutine 中异步执行,绑定方法立即返回。

### 6. 目录可写探测与降级

点击「立即更新」时先在 exe 所在目录尝试写入临时文件,失败(Program Files / UAC 场景)则提示并降级为跳转 GitHub 手动更新。下载文件直接写入 exe 同目录,保证后续 `move` 同卷原子。

### 7. 失败兜底

- bat 内 `move` 失败(杀软短暂锁定等)重试 3 次;仍失败则启动旧 exe(仍在原路径)并保留 `.new`,应用可用,下次「重启更新」可再试。
- 下载中断仅产生 `.part`,下次下载前删除重来。
- dev 版本与非 Windows 平台不提供一键更新(沿用 dev 短路先例与 stub 模式)。

## Risks / Trade-offs

- [bat 控制台窗口闪现] → 接受(一闪而过,与项目极简风格一致);如体验不佳可换 PowerShell `-WindowStyle Hidden`,列为后续优化。
- [bat 路径转义地狱:用户目录含空格/`&`/`!` 等] → bat 中所有路径加引号、特殊字符转义,单测覆盖生成内容;bat 落在 exe 同目录,用 `%~dp0` 引用自身路径,减少注入面。
- [杀软对自我替换告警] → 无签名软件常见现象,接受;`move` 失败重试 + 启动旧 exe 兜底,保证应用不失效。
- [替换期间断电,`.new` 残留] → 下次启动清理 `.part`,`.new` 被「重启更新」复用或下次覆盖;exe 本体始终完整,应用可启动。
- [GitHub API 限流] → 检查结果缓存资产 URL,一键更新不重复请求;下载走 `browser_download_url`(codeload CDN),不占 API 配额。

## Migration Plan

- 无数据迁移:用户数据(timeline/settings/countdowns)位于 `os.UserConfigDir()/health-tool/`,与 exe 替换无关。
- 发布:打 `v*` tag 由现有 CI 构建发布;旧版本用户点击「检查更新」→「立即更新」即可升级,无需重新分发安装方式。
- 回滚:手动从 GitHub 下载旧版本 exe 覆盖(不提供自动回滚,符合 Non-Goals)。

## Open Questions

- sha256 校验:是否让 CI 顺带生成 checksum 资产并在下载后校验(防损坏,低成本加分项)。当前不阻塞,可后续追加。
- bat 隐藏方式:默认接受闪窗;若用户反馈不佳,切换 PowerShell 隐藏窗口方案。

## 设计修订(评审后)

评审发现「下载完成 → 取消 → 下次进入应用」场景下待更新状态全部丢失,追加以下决策:

- **待更新状态由后端暴露**:新增绑定方法 `PendingUpdateInfo()`(返回 `.new` 是否存在与版本),设置页加载时查询;存在则渲染「重启更新」入口。前端不再用内存变量记忆下载状态。
- **`.new` 版本元数据**:下载完成时在 `.new` 旁写 `health-tool.exe.new.version`(内容为版本号)。确认框文案版本一律从该文件读取,消除对内存缓存与弹框时机的依赖。
- **检查同步事实**:`CheckForUpdates` 返回 `up-to-date` 时清理 `.new` 与 `.new.version`(网络错误时保守保留,不清理)。
- **版本落后保留旧 `.new`**:检查发现 `.new` 版本低于最新 release 时保留 `.new`,反馈区只显示「重启更新」入口,不提示新版本;用户重启后再检查更新完成升级。
- **短路逻辑移除**:`DownloadAndApplyUpdate` 只负责无 `.new` 时的下载;「重启更新」统一走 `ApplyUpdateAndRestart`(校验 `.new` → 读版本 → 弹确认框 → 退出重启)。
- **元数据清理**:bat 在 move 成功后删除 `.new.version`;启动时发现 `.new` 不存在而 `.version` 存在时删除孤儿 `.version`。
