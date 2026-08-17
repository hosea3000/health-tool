## Context

应用是 Wails v2 Windows 托盘应用，`beforeClose` 隐藏窗口而非退出，输入监控、托盘、通知持续运行。已有 `SingleInstanceLock`，二次启动唤起主窗口。设置页（`frontend/src/views/settings.js`）目前只有「关于」与「更新」两个 section。平台能力沿用双文件模式：`tray_windows.go`（`//go:build windows`）与 stub（`//go:build !windows`）。

自动更新流程（`updater_apply_windows.go`）用 bat 脚本 `move /y` 原地覆盖 exe，路径不变；因此 Run key 里的绝对路径在自动更新后依然有效，无需配合逻辑。

开发环境是 Linux：`wails dev` 走非 Windows stub。CI 在 ubuntu 交叉编译，注册表逻辑无法集成测试。

## Goals / Non-Goals

**Goals:**
- 用户可在设置页一键开关开机自启动（HKCU Run key，用户级权限）。
- 开机自启后应用静默进入托盘驻留，不弹主窗口。
- `GetSettings` 能反映自启实际状态，供 UI 初始化。
- 非 Windows 平台 stub 化，不破坏开发流。

**Non-Goals:**
- 不感知任务管理器「启动应用」的 `StartupApproved` 禁用标记（UI 可能与实际不符，接受）。
- 不做 exe 移动后的路径自愈。
- 不在托盘菜单放自启开关。
- 不引入启动延迟、条件触发等任务计划程序特性。
- 不重构设置存储结构（`settings.json` 不存自启字段，真相源在注册表）。

## Decisions

### D1：机制选 HKCU Run key（放弃启动文件夹 / 任务计划程序）
- Run key：用户级权限、`golang.org/x/sys/windows/registry` 直接读写、任务管理器可见可管理。
- 启动文件夹：需 COM 创建 `.lnk`，引入 ole 依赖，成本高无收益。
- 任务计划程序：支持延迟/条件触发，对托盘应用过度设计。
- Run key 值：`"C:\...\health-tool.exe" --hidden`（路径含空格时整体加引号——构造逻辑抽为纯函数并单测）。

### D2：`--hidden` 标记 + Wails 启动选项
- `main.go` 解析 `--hidden`（纯函数 `hiddenFromArgs`），映射到 `options.App.StartHidden: true`（Wails v2.13 实际字段名，已验证存在）。
- 手动双击 exe（无 flag）行为不变：正常显示窗口。开机实例与手动实例撞车时由既有 `SingleInstanceLock` 唤起窗口，语义天然正确。

### D3：注册表为唯一真相源（放弃 settings.json 双写）
- UI 状态 = Run key 是否存在该值，不引入第二个需要同步的状态。
- 代价：用户在任务管理器禁用自启后（Windows 写 `StartupApproved\Run` 二进制标记而非删 key），UI 仍显示"已开启"。作为已接受边界记录在 spec。
- 自愈（路径过期重写）留作后续增量，不进本期。

### D4：独立绑定方法 `SetAutoStart(bool) error`，不改 `SaveSettings` 签名
- 注册表写入可能失败（权限、策略），独立方法把 error 干净地返回给 UI 的 toggle。
- `Settings` 结构体新增 `AutoStart bool` 字段，仅由 `GetSettings` 从注册表实时读取填充，`SaveSettings` 忽略该字段、`settings.json` 不持久化它——避免"文件里存的值"与"注册表实际值"打架。
- 前端契约变更：需 `wails dev` / `wails build` 重新生成 `frontend/wailsjs`。

### D5：状态查询读 Run key 值本身
- 「是否开启」= `HKCU\...\Run` 下名为 `health-tool` 的值是否存在且非空。
- 查询与开启使用同一 exe 路径（`os.Executable()`），保证读到的一定是自己写的。

### D6：平台拆分与文件布局
- `autostart_windows.go` / `autostart_stub.go`，导出 `setAutoStart(enabled bool) error` 与 `autoStartEnabled() (bool, error)`（包内小写，App 方法做薄封装）。
- stub 返回"平台不支持"错误；设置页在无 Wails bridge（纯前端预览）或状态读取失败时隐藏/禁用 toggle。

## Risks / Trade-offs

- [Wails `Hidden` 选项在 Windows 行为不符预期（闪窗/不隐藏）] → 已确认实际字段为 `options.App.StartHidden`；真机表现靠人工验证清单（tasks 4.3）确认。
- [任务管理器禁用后 UI 失真] → spec 记录为已接受边界；后续增量可读 `StartupApproved` 修复。
- [用户挪动 exe 导致 Run key 路径过期] → 接受；自动更新场景已由 `move /y` 原地覆盖保证不受影响。
- [CI 无 Windows 环境测不到注册表] → 可测部分抽纯函数（命令行值构造、flag 解析）单测；注册表读写靠人工在 Windows 验证。
- [`Settings` 加字段后旧 `settings.json` 反序列化] → 新字段零值即 false，且不持久化，无兼容问题。

## Migration Plan

纯新增能力，无数据迁移。回滚 = 设置页关闭自启（删 Run key）后回退版本；旧版本不认识 `--hidden` 也无碍（Run key 值只在自启时由新版本消费）。

## Open Questions

（无。D2 的 Wails 隐藏启动字段已确认为 `StartHidden`，无需 fallback。）
