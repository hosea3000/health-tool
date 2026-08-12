## 1. 后端顺序存储

- [x] 1.1 `card_order_store.go`：`card_order.json`（`{order: []}`）读写，照抄 `timeline_store.go` 模式，空文件容错
- [x] 1.2 `app.go` 追加 `GetCardOrder() []string` / `SaveCardOrder(order []string) bool`；启动时加载进 App 内存
- [x] 1.3 `wails generate module` 重新生成绑定，确认 `GetCardOrder` / `SaveCardOrder` 出现在 `App.js`
- [x] 1.4 `app_test.go` 补测试：保存/读取顺序、空文件返回 nil、持久化往返

## 2. dashboard 扁平化与对账

- [x] 2.1 `views/dashboard.js`：弃用 `card-group` 分组，卡片直接 append 进 `#card-grid`
- [x] 2.2 挂载对账：`Promise.all` 收集各工具卡片 → 读存储顺序过滤失效 key → 按序渲染，新 key 末尾
- [x] 2.3 刷新对账：按 `data-card` key 原位替换内容、新 key 追加末尾、消失 key 移除节点
- [x] 2.4 无存储文件时默认注册顺序渲染（现状行为）

## 3. 工具模块补卡片 key

- [x] 3.1 `tools/reminder/index.js`：卡片 `data-card="reminder"`
- [x] 3.2 `tools/countdown/index.js`：事件卡 `data-card="countdown:<id>"`

## 4. 拖拽交互

- [x] 4.1 卡片 `draggable="true"`；dragstart/dragover/drop/dragend 事件，整卡可拖
- [x] 4.2 drop 后序列化 DOM 卡片顺序 → `SaveCardOrder` 持久化
- [x] 4.3 拖拽进行中跳过该工具刷新（`dragging` 标志），dragend 恢复
- [x] 4.4 `app.css`：拖拽态样式（cursor、dragging 半透明、插入指示）

## 5. 验证

- [x] 5.1 `go test ./...` 通过
- [x] 5.2 `cd frontend && npm run build` 通过
- [x] 5.3 真机验证：跨工具拖拽排序、重启后顺序保持、新建事件卡在末尾、点击仍进详情、刷新不重排
