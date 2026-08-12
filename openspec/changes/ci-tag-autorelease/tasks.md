## 1. 项目配置

- [x] 1.1 在 `wails.json` 中补充 `info` 段：`companyName`、`productName`、`productVersion`（默认 `1.0.0`）、`copyright`，并验证 `wails build` 能正确注入 exe 文件属性

## 2. 发布工作流

- [x] 2.1 修改 `.github/workflows/release.yml`：`on.push` 同时监听 `main` 分支与 `v*` tag，`permissions.contents: write`，单 job 跑在 `windows-latest`
- [x] 2.2 配置步骤：`actions/checkout@v4`（`fetch-depth: 0`）→ `actions/setup-go@v5` → `actions/setup-node@v4`
- [x] 2.3 添加"从 tag 提取版本写入 wails.json"步骤（PowerShell/node 脚本，tag 时 `TrimStart('v')`，分支时用 commit SHA）
- [x] 2.4 添加 `wails build -clean -nsis` 构建步骤（前置：`go install` 安装 wails CLI v2.13.0；NSIS/makensis 由 windows-latest 预装）
- [x] 2.5 tag 时用 `gh release create --generate-notes` 发布；main 分支时用 `actions/upload-artifact` 保存验证产物

## 3. 收尾

- [x] 3.1 删除 `.goreleaser.yaml`（改用纯 GitHub Actions + `gh` CLI）
- [ ] 3.2 推送 tag（如 `v0.1.0`）到远端，确认 GitHub Actions run 全绿
- [ ] 3.3 检查 Release 产物完整：`health-tool.exe`、`*-installer.exe`、changelog
- [ ] 3.4 抽查 exe 文件属性与安装器显示的版本号与 tag 一致
