# Design: 版本更新检查

## Context

- Wails v2 桌面应用：Go 后端（`package main`，含 `app.go` 绑定层）+ 无框架 Vite 前端（`views/dashboard.js` / `views/detail.js` 视图切换）。
- 用户数据走 `model` / `store` / `domain` 分层；绑定方法即前后端数据契约，`wailsjs/go/main/App.js` 由 Wails 根据 `App` 导出方法自动生成。
- 版本号目前只存在于 `wails.json`（构建时写入 exe 文件属性），**运行时不可读**。
- CI 打 `v*` tag 触发发布（`.github/workflows/release.yml`），产生 GitHub Release。
- 前端有 `hasWailsBridge` 检测模式：在纯浏览器预览时降级到本地 mock，不依赖后端绑定。

## Goals / Non-Goals

**Goals:**
- 运行时获得当前版本号，并支持手动检查 GitHub Release 上的最新版本。
- 语义化版本比较，正确区分"有更新 / 已是最新 / 检查失败"三态。
- 通过 dashboard 顶栏设置入口进入新的设置页，页面内完成检查与跳转。
- `dev`（本地开发/未注入）构建不发起网络检查。

**Non-Goals:**
- 自动检查、定时检查、启动时静默检查。
- 「忽略此版本」记忆功能。
- 自动下载/安装更新。
- 将现有的提醒/休息设置收纳进设置页（设置中心化是后续独立 change）。

## Decisions

### 1. 版本号来源：编译时 ldflags 注入

在 `package main` 声明 `var version = "dev"`。CI 的 `wails build` 追加 `-ldflags "-X main.version=<tag去v>"`（v2.13.0 已确认支持 `-ldflags` flag）。本地开发与未打 tag 构建保持默认 `dev`。

- **为什么不是读 exe 资源**：`-X` 只对 main 包变量生效且需要 win32 API + 破坏跨平台，复杂且不必要。
- **为什么不是硬编码常量**：每次发布忘记改就产生错误版本号，ldflags 由 tag 单一来源派生，不可能失同步。
- `productVersion` 写入 `wails.json` 的既有逻辑**保留**——两者各司其职：前者进 exe 文件属性（资源管理器可见），后者供运行时比较。

### 2. 检查源：GitHub Releases API（未认证）

`GET https://api.github.com/repos/hosea3000/health-tool/releases/latest`，解析 `tag_name` 与 `html_url`。未认证限速 60 次/小时/IP——手动检查频率极低，完全够用。

- 超时设定 10 秒（`http.Client`），避免 UI 长时间悬挂。
- 404（仓库尚无 release）归类为明确的错误态提示，而非"已是最新"。
- 解析失败、网络错误统一归入 `error` 态，前端展示人性化文案。

### 3. 版本比较：手写语义化比较，不引第三方依赖

比较逻辑拆三段数字（主/次/补丁），两侧先剥 `v` 前缀。tag 含 pre-release 后缀（如 `v0.2.0-beta`）时按"低于正式版"处理（简单规则：非纯数字段视为 0 且小于正规版本号；精确规则采用三段数字逐位比较，忽略 pre-release 段）。

- **为什么不用第三方库**：幂等自查，本项目 `go.mod` 依赖极克制（一个 ~30 行纯函数 + 单测）。
- 不可解析的版本号（如 tag 乱填非 semver）在比较时容错为"不作为更新"。

### 4. 数据契约：结构化三态结果

`CheckForUpdates()` 返回 `model.UpdateCheckResult`：

```
Status         string  // "up-to-date" | "update-available" | "error"
CurrentVersion string  // 当前运行时版本（dev 时也返回）
LatestVersion  string  // 最新版本（仅 update-available 时非空）
ReleaseURL     string  // release 页面地址（仅 update-available 时非空）
Message        string  // 人性化错误/提示文案
```

- **为什么不用错误字符串或布尔**：三态是前端分支渲染的基础，"有更新 + 版本号 + 链接"需要一并返回，避免前端二次拼装。
- `dev` 时直接返回 `up-to-date` + Message「当前为开发版本」。

### 5. UI：顶栏设置图标 + 独立设置页，页面内反馈取代模态弹窗

- 图标：内联 SVG 齿轮（`⚙` 字符在部分 Windows 字体下渲染成方框，SVG 跨平台一致），置于 dashboard 顶栏 wordmark 右侧；交互沿用现有 `details` 下拉的点击外部收起模式——本次图标直接作为视图跳转按钮，不展开下拉。
- 设置页 `views/settings.js`：复用 `detail.js` 的 shell（`topbar` 返回按钮 `wordmark-link` + `main` 内容 + `footer`），`main.js` 增加 `showSettings()` 分支（dashboard ⇄ settings）。
- 页面内「更新」区：检查按钮 → loading → 结果反馈（三态文案）+ 有更新时「前往 GitHub 查看」按钮。
- 跳转：前端调用 wails runtime 绑定的 `BrowserOpenURL`（浏览器预览时降级 `window.open`），无需后端再包一层方法。

### 6. 测试策略

- 语义化比较为纯函数，覆盖：补丁级新旧、`v` 前缀剥离、pre-release 后缀、非法版本号容错。
- 检查器注入 `http.Client` 与 `now`，用 `httptest` 模拟 API 响应（有更新 / 无更新 / 404 / 超时）。
- 前端以 `hasWailsBridge` 降级保证浏览器里可预览。

## Risks / Trade-offs

- [未认证 GitHub API 限速，理论上手动狂点可触发 403] → 403 归入 `error` 态提示"稍后再试"；手动检查天然低频，风险低。
- [离线/代理环境下检查失败] → 归入 `error` 态，页面内提示网络原因，不影响其他功能。
- [`-ldflags` 注入遗漏导致发布版版本号为 `dev`] → 用户点检查只见"开发版本"文案；由 CI tag 单一来源 + 发布后自检降低概率。
- [GitHub API 返回的 `tag_name` 格式异常] → 比较容错为"不作为更新"，不回导致崩溃。
- [设置页与 detail 页 shell 重复] → 当前按复制 detail shell 处理（~15 行），后续若出现第三个页面再抽公共布局组件。

## Migration Plan

- 无数据迁移：不新增持久化，`CheckForUpdates` 是纯新增绑定方法，与旧 exe 完全兼容。
- 发布顺序：前端构建 → 代码合并后打 tag 即产出带 `main.version` 的 exe；旧版本用户无更新入口，待新版本发布后自然获得。
- 回滚：功能无状态，回退代码即可；CI 的 ldflags 追加与 `productVersion` 写入互不影响。

## Open Questions

- 无（方案已收敛）。