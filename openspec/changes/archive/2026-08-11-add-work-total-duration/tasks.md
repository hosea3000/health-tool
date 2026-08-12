## 1. 前端汇总与展示

- [x] 1.1 在 `frontend/src/main.js` 新增恒三位格式化函数，将秒数格式化为 `xxx小时xx分xx秒`。
- [x] 1.2 在时间线标题行（`工作与休息记录` 旁）新增累计工作时长展示，对 `timeline` 中 `kind === 'working'` 的 `durationSeconds` 求和并渲染。

## 2. 行为验证

- [x] 2.1 空时间线显示 `0小时0分0秒`，含多条 `working` 记录（含进行中记录）时求和正确且随刷新更新。
- [x] 2.2 确认 `resting`、`idle-paused` 记录不计入累计工作时长。
- [x] 2.3 运行 `cd frontend && npm run build` 验证前端构建通过；`go test ./...` 保持通过（后端无改动）。
