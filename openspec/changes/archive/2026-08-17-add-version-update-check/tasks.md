# Tasks: 版本更新检查

## 1. 版本号与检查核心（后端）

- [x] 1.1 在 `package main` 声明运行时版本号变量 `var version = "dev"`（放 `main.go`），供 ldflags 注入覆盖
- [x] 1.2 实现语义化版本比较纯函数（拆分主/次/补丁三段、剥 `v` 前缀、pre-release 按低于正式版、非法版本号容错为"无更新"），放 `updater.go` 并编写单测（覆盖补丁更新、`v` 前缀、pre-release、非法格式）
- [x] 1.3 实现检查器：`CheckForUpdates` 逻辑请求 `https://api.github.com/repos/hosea3000/health-tool/releases/latest`，解析 `tag_name` 与 `html_url`，10 秒超时，映射为三态结果
- [x] 1.4 为检查器编写单测：用 `httptest` 模拟「有更新 / 已是最新 / 404 无 release / 网络错误」四种响应

## 2. 数据契约与绑定

- [x] 2.1 在 `model` 包新增 `UpdateCheckResult` 结构：`Status`（`up-to-date`/`update-available`/`error`）、`CurrentVersion`、`LatestVersion`、`ReleaseURL`、`Message`
- [x] 2.2 在 `App` 上新增 `CheckForUpdates() model.UpdateCheckResult` 绑定方法：`dev` 短路（不请求网络，返回提示文案）、注入可替换的 `http.Client` 便于测试
- [x] 2.3 跑一次 `wails dev` 或 `wails build` 重新生成 `frontend/wailsjs`（`CheckForUpdates` 绑定）

## 3. CI 版本注入

- [x] 3.1 修改 `.github/workflows/release.yml` 的 `Wails build` 步骤，追加 `-ldflags "-X main.version=${GITHUB_REF_NAME#v}"`；保留既有 `productVersion` 写入步骤

## 4. 前端入口与设置页

- [x] 4.1 dashboard 顶栏「健康工具箱」旁新增齿轮图标（内联 SVG），点击切换设置视图
- [x] 4.2 新增 `frontend/src/views/settings.js`：复用 detail shell（返回按钮 + main + footer），渲染「关于」区（应用名 + 当前版本号）与「更新」区
- [x] 4.3 `frontend/src/main.js` 增加 `showSettings()` 分支（dashboard ⇄ settings）
- [x] 4.4 `style.css` 补充设置页布局与齿轮图标样式
- [x] 4.5 更新区交互：「检查更新」按钮加载态防重、三态反馈渲染、`update-available` 时「前往 GitHub 查看」按钮调用 wails runtime `BrowserOpenURL`（浏览器预览降级 `window.open`）；前端走 `hasWailsBridge` 降级模式（浏览器预览时可检查用 mock 结果）

## 5. 验证

- [x] 5.1 `go test ./...` 全部通过
- [x] 5.2 `cd frontend && npm run build` 通过
- [x] 5.3 `wails dev` 手动验证：设置页入口、版本号显示、dev 版本点检查显示"开发版本"提示、返回 dashboard 正常