## 1. 项目配置

- [x] 1.1 在 `wails.json` 中补充 `info` 段：`companyName`、`productName`、`productVersion`（默认 `1.0.0`）、`copyright`，并验证 `wails build` 能正确注入 exe 文件属性
- [x] 1.2 创建 `.goreleaser.yaml`：`project_name: health-tool`，`builds` 使用 go builder（`goos: windows`、`goarch: amd64`）+ post hook 用 `build/bin/health-tool.exe` 覆盖产物，`archives` 用 zip，`release.extra_files` glob 收集 `build/bin/*setup*.exe`
- [x] 1.3 本地用 `goreleaser check` 校验 `.goreleaser.yaml` 配置合法

## 2. 发布工作流

- [x] 2.1 创建 `.github/workflows/release.yml`：`on.push.tags: ["v*"]`，`permissions.contents: write`，单 job 跑在 `windows-latest`
- [x] 2.2 配置步骤：`actions/checkout@v4`（`fetch-depth: 0`）→ `actions/setup-go@v5` → `actions/setup-node@v4`
- [x] 2.3 添加"从 tag 提取版本写入 wails.json"步骤（PowerShell/node 脚本，`TrimStart('v')` 后写 `info.productVersion`）
- [x] 2.4 添加 `wails build -clean -nsis` 构建步骤（前置：`go install` 安装 wails CLI v2.13.0；NSIS/makensis 由 windows-latest 预装）
- [x] 2.5 添加 `goreleaser/goreleaser-action@v7` 步骤（`version: "~> v2"`、`args: release --clean`、`env.GITHUB_TOKEN`）

## 3. 验证

- [ ] 3.1 推送 tag（如 `v0.1.0`）到远端，确认 GitHub Actions run 全绿
- [ ] 3.2 检查 Release 产物完整：`health-tool.zip`、`*setup*.exe`、`checksums.txt`、changelog
- [ ] 3.3 抽查 exe 文件属性与安装器显示的版本号与 tag 一致
