## 1. 项目配置

- [x] 1.1 在 `wails.json` 中补充 `info` 段：`companyName`、`productName`、`productVersion`（默认 `1.0.0`）、`copyright`，并验证 `wails build` 能正确注入 exe 文件属性

## 2. 发布工作流

- [x] 2.1 修改 `.github/workflows/release.yml`：`on.push.tags: ["v*"]` 仅 tag 触发，`permissions.contents: write`，单 job 跑在 `ubuntu-latest`（`wails build -platform windows/amd64` 交叉编译）
- [x] 2.2 配置步骤：`actions/checkout@v4`（`fetch-depth: 0`）→ `actions/setup-go@v5` → `actions/setup-node@v4`
- [x] 2.3 添加"从 tag 提取版本写入 wails.json"步骤（PowerShell/node 脚本，`TrimStart('v')` 后写 `info.productVersion`）
- [x] 2.4 添加 `wails build -clean` 构建步骤（前置：`go install` 安装 wails CLI v2.13.0）
- [x] 2.5 用 `gh release create --generate-notes` 发布单 exe

## 3. 收尾

- [x] 3.1 删除 `.goreleaser.yaml`（改用纯 GitHub Actions + `gh` CLI）
- [x] 3.2 推送 tag（如 `v0.1.1`）到远端，确认 GitHub Actions run 全绿
- [x] 3.3 检查 Release 产物完整：`health-tool.exe`、changelog
- [x] 3.4 抽查 exe 文件属性显示的版本号与 tag 一致
