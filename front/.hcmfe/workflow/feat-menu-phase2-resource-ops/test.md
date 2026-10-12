# feat-menu-phase2-resource-ops 测试清单

> 工作流: `feat-menu-phase2-resource-ops`
> 关联文档: Coding（`coding.md`）、需求 `docs/reqs/二期资源运营重组.md`（AC 编号对应；标「已移至资源纳管主体」的只在终扫回归）

## 验证范围

- 功能边界: 账单从 `/bill/*` 迁到 `/resource/bill/*` 并兼容旧地址；`useWhereAmI` 把 `/resource/bill/` 判为账单；删除原资源接入路由文件、过渡布局（`main-content.vue` 云账号树两栏）与旧文件；首批全量旧地址终扫。
- 不在范围: 业务回收站（仍用 `recyclebin-manager`，归收尾单 F-010）；业务地址 `/business/:bizId`；服务请求里无引用的旧组件。

## 测试环境

- 前端入口: `<本地开发或测试环境地址，由执行人确认>`
- 测试账号: `<test_account>`（资源运营全部权限，账单功能开启）；`<test_account_no_root>`（无一级账号查看权限、有二级账号查看权限）
- 数据准备: 一级账号、二级账号、账单汇总各有数据；资源纳管下至少一个账号

## 用例清单

### P0 - 主流程（必测）

| ID | 场景 | 前置 | 操作 | 期望 | covered-by | 回滚 |
|----|------|------|------|------|------------|------|
| P0-01 | 账单新地址（AC-003、AC-012） | `<test_account>` | 侧栏依次点一级账号、二级账号、账单管理；账单管理切汇总 / 明细 / 调整页签；一级账号点「录入」、二级账号点「创建」 | 地址为 `/resource/bill/root-account`、`/resource/bill/main-account`、`/resource/bill/manage/summary/manage` 等；顶栏高亮资源运营、侧栏高亮对应项；页面左侧无云账号树；列表走账单接口（Network 无资源纳管接口）；新建页面包屑返回到对应列表 | | — |
| P0-02 | 账单旧地址（AC-004） | — | 带 `?a=1` 打开 `#/bill/root-account`、`#/bill/main-account`、`#/bill/root-account/create`、`#/bill/main-account/create`、`#/bill/account-manage/first-account`、`#/bill/account-manage/second-account`、`#/bill/bill-manage`、`#/bill/bill-manage/summary/operation-record`、`#/bill/bill-manage/detail`、`#/bill/bill-manage/adjust` | 都转到 `/resource/bill/*` 对应页，`a=1` 保留 | | — |
| P0-03 | `/bill/account-manage` 分流（AC-005） | 分别用两个测试账号 | 刷新打开 `#/bill/account-manage?a=1` | 有一级账号权限进 `/resource/bill/root-account`，否则进 `/resource/bill/main-account`，`a=1` 保留 | | — |
| P0-04 | 原资源接入旧地址（AC-010） | — | 打开 `#/resource/resource?type=vpc&accountId=<A>`、`#/resource/resource`、`#/resource/service-apply/cvm`、`#/resource/resource/account`、`#/resource/resource/record`、`#/resource/resource/recycle`、裸 `#/resource` | 依次进资源纳管 VPC 页签（账号 A）、主机页签、新建主机页、账号管理、操作记录、回收管理、资源纳管主机列表；无空白页 | | — |
| P0-05 | 旧资源详情地址（AC-011） | — | 打开 `#/resource/detail/vpc?type=tcloud&id=<ID>` | 进 `/resource/manage/vpc/details/<ID>?vendor=tcloud` | | — |
| P0-06 | 完整菜单树（AC-001） | `<test_account>` | 展开资源运营侧栏 | 自上而下为云资源（资源纳管、账号管理、回收管理）→ 云账号管理（一级账号、二级账号、账单管理）→ 操作记录；除资源纳管外各页左侧无云账号树、无第二条面包屑 | | — |

### P1 - 异常分支（必测）

| ID | 场景 | 前置 | 操作 | 期望 | covered-by |
|----|------|------|------|------|------------|
| P1-01 | 账单查看权限（AC-006、AC-S01） | `<test_account_no_root>` | 直接打开 `#/resource/bill/root-account` 与旧地址 `#/bill/root-account` | 都显示框架内无权限页，只提示联系管理员授权 | |
| P1-02 | 无对应关系的旧地址（AC-013） | — | 打开裸 `#/bill`、`#/bill/unknown`、`#/resource/unknown` | 进 404 页，不出现空白页 | |
| P1-03 | 终扫 | 代码检查 | coding.md §5 终扫表逐项 | 全部满足（`'applyAccount'` 只剩无引用的旧组件） | |

### P2 - UI 细节（按需）

- [ ] 资源纳管、账号管理、回收管理、操作记录、账单各页内容区留白与改前一致（去掉 `main-content` 过渡容器后无多余 / 缺失内边距）

## 验证结论

> 每条用例填写：PASS / FAIL / Skipped + 简要说明（FAIL 必须附复现步骤、错误截图链接或日志片段）

| ID | 结果 | 备注 |
|----|------|------|
| P0-01 ~ P0-06 | | |
| P1-01 ~ P1-03 | | |

- 执行日期:
- 总体结论:
- 后续行动:
