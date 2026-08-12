## Why

应用目前是单一工具（久坐提醒），但 wordmark 已定位为"健康工具箱"，topbar 也已预留工具编号。需要落地一个多工具平台骨架，让倒数日等新工具能以卡片形式接入 dashboard，而不是继续把新功能硬塞进单工具页面。

## What Changes

- 新增 **dashboard 首页**：打开应用进入卡片网格展示界面，每张卡片是一个工具实例的预览。
- 引入**工具/卡片分层**：工具是类型（编译进程序），卡片是实例；一个工具可以产出一张（久坐提醒）或多张（倒数日占位）卡片。
- 点击任意卡片进入该**工具详情页**（工具级聚合，非实例级）；详情页提供返回 dashboard 的入口。
- 卡片**自声明刷新频率**：dashboard 按各工具声明的间隔调度刷新（久坐提醒秒级，倒数日分钟级）。
- **工具注册表**：前端以 `{ renderCards, renderDetail, refreshInterval }` 注册工具，新增工具无需改动 dashboard 本体。
- 现有界面迁移：久坐提醒的 hero/insight/timeline 迁入其详情页；topbar 的全局"设置"按钮下沉为久坐提醒详情页内设置，去掉"01"编号。
- **BREAKING**（仅 UI）：topbar 结构变化，"设置"入口从全局变为工具内。

## Capabilities

### New Capabilities
- `tool-dashboard`: dashboard 首页、卡片网格、工具/卡片分层、工具详情页路由、按工具声明的间隔刷新卡片

### Modified Capabilities
- `work-rest-timeline`: 时间线展示位置从主界面首页迁移到久坐提醒工具详情页

## Impact

- `frontend/src/main.js`：拆分为视图路由 + dashboard 视图 + detail 视图 + 工具模块目录。
- 新增 `frontend/src/views/`（dashboard/detail）与 `frontend/src/tools/`（reminder）模块。
- topbar 与设置弹窗的 HTML/样式调整。
- Go 端无结构性改动：保持现有肥 App，后端方法（Status/Timeline/Settings）不变，不引入 Go 工具接口。
- 不新增前端依赖与第三方库。
