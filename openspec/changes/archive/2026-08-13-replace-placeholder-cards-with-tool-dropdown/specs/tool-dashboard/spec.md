# Tool Dashboard

## ADDED Requirements

### Requirement: 顶栏工具下拉导航
dashboard 顶栏 SHALL 提供"工具"下拉列表，固定列出注册表全部工具（不随各工具是否有数据而变化），点击任意项 SHALL 切换到对应工具详情页；下拉 SHALL 键盘可达并标注 `aria-expanded` 状态。

#### Scenario: 下拉列出全部工具
- **WHEN** dashboard 顶栏的"工具"下拉被展开
- **THEN** 列表列出注册表登记的全部工具（含当前无任何卡片数据的工具）

#### Scenario: 点击下拉项进入工具详情页
- **WHEN** 用户在下拉列表中点选某个工具
- **THEN** 界面切换到该工具的详情页，展示该工具完整内容

#### Scenario: 空工具经下拉可达
- **WHEN** 某个工具没有任何卡片、因此不出现在卡片网格中
- **THEN** 用户仍可通过顶栏下拉进入该工具详情页

#### Scenario: 下拉键盘可达
- **WHEN** 用户通过键盘 Tab/方向键/回车操作下拉
- **THEN** 下拉可聚焦、可展开、可选中并进入工具详情页，且 `aria-expanded` 状态正确切换

#### Scenario: 新工具自动出现在下拉
- **WHEN** 一个新工具在注册表登记后
- **THEN** 该工具自动出现在顶栏下拉中，无需修改 dashboard 本体逻辑

## MODIFIED Requirements

### Requirement: 卡片网格扁平排列
dashboard 卡片网格 SHALL 将所有工具的卡片扁平排列为等宽表格，不按工具分组。未保存自定义顺序时，卡片顺序 SHALL 遵循工具注册顺序（久坐提醒在前，倒数日事件随后）；用户保存自定义顺序后，按自定义顺序排列。没有任何卡片的工具 SHALL 不出现在网格中（不渲染占位卡），仅通过顶栏下拉进入。

#### Scenario: 多工具卡片混排
- **WHEN** dashboard 同时存在久坐提醒卡片与多个倒数日事件卡片，且未保存自定义顺序
- **THEN** 所有卡片按注册顺序平铺进同一网格，自动换行，而不是按工具各自竖排

#### Scenario: 顺序不随刷新跳动
- **WHEN** 某个工具按自己的间隔刷新卡片
- **THEN** 该工具卡片只在原位更新，不改变网格中所有卡片的相对顺序

#### Scenario: 空工具不渲染卡片
- **WHEN** 某个工具没有任何卡片
- **THEN** 该工具不出现在卡片网格中，不渲染占位卡；该工具仍可通过顶栏下拉进入详情页

#### Scenario: 占位 key 按失效 key 丢弃
- **WHEN** `card_order.json` 中残留占位卡时代保存的 key（等于工具 id，如 `countdown`）
- **THEN** 加载顺序时该 key 被当作失效 key 忽略，不产生卡片
