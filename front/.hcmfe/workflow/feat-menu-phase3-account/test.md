# feat-menu-phase3-account 测试清单

> 工作流: `feat-menu-phase3-account`
> 关联文档: coding.md（本单无 prd / design / api）

## 验证范围

- 功能边界: 账号管理从「资源运营 → 云资源」迁到「工作台 → 公共事务」、符号改名、路由迁到 `serviceViews`（地址 `/service/account*`）、列表不再展示资源账号、旧地址转址、视图权限与收藏
- 不在范围: 账号登记流程与后端接口、账号详情页功能、目录结构调整（二期已定不按 PR 重组）、资源纳管里的账号状态侧滑

## 测试环境

- 前端入口: `对应业务的测试环境地址，由执行人在测试时确认；不要把具体域名写入仓库`
- 测试账号: `<test_account>`（需具备账号查看权限；另备一个缺账号查看权限的账号）
- 数据准备:
  - 库里同时存在「登记账号」「安全审计账号」「资源账号」三类数据
  - 至少 1 个账号有详情可打开（验证明细与页签）

## 用例清单

### P0 - 主流程（必测）

| ID | 场景 | 前置 | 操作 | 期望 | covered-by | 回滚 |
|----|------|------|------|------|------------|------|
| P0-01 | 菜单位置 | 已登录 | 查看工作台侧栏与资源运营侧栏 | 工作台「个人事务」下有「账号管理」（位于单据管理之后）；「公共事务」下为资源选型、权限策略库；资源运营「云资源」下已无账号管理 | | 无 |
| P0-02 | 进入账号管理列表 | 已登录 | 工作台 → 公共事务 → 账号管理 | 地址为 `/service/account`；顶栏高亮工作台；侧栏高亮账号管理 | | 无 |
| P0-03 | 列表不含资源账号 | 库里有资源账号 | 查看列表并翻页 | 列表无资源账号；请求体 `filter.rules` 含 `{ field: 'type', op: 'neq', value: 'resource' }` | | 无 |
| P0-04 | 账号类型筛选 | 同上 | 打开「账号类型」筛选下拉 | 只有「登记账号」「安全审计账号」，无「资源账号」 | | 无 |
| P0-05 | 叠加搜索条件 | 同上 | 选「登记账号」+ 输入名称搜索 | 仍排除资源账号（默认规则未被覆盖），结果符合条件 | | 清空搜索 |
| P0-06 | 账号详情 | 列表有数据 | 点击账号名进入详情 | 地址 `/service/account/details/<id>/basic`；面包屑「账号详情」；返回回到列表 | | 无 |
| P0-07 | 账号登记 | 有登记权限 | 进入「账号登记」 | 地址 `/service/account/check-in`；顶部高亮工作台；提交后回到列表且新账号可见 | | 删除该账号 |

### P1 - 异常分支（必测）

| ID | 场景 | 前置 | 操作 | 期望 | covered-by |
|----|------|------|------|------|------------|
| P1-01 | 旧地址：列表 | 已登录 | 访问 `/resource/account` | 转到 `/service/account` | |
| P1-02 | 旧地址：详情 | 已知某账号 ID | 访问 `/resource/account/details/<id>`、`/resource/account/details/<id>/basic?_f=1`（旧详情页带页签段，历史收藏里是这个形态） | 均转到 `/service/account/details/<id>/basic`，query（含 `_f`）保留 | |
| P1-03 | 旧地址：登记 | 已登录 | 访问 `/resource/account/check-in`、`/resource/account/add` | 均转到 `/service/account/check-in` | |
| P1-04 | 旧地址：资源状态 | 已登录 | 访问 `/resource/resource/account` | 仍进资源纳管的账号状态侧滑（行为不变） | |
| P1-05 | 无查看权限 | 缺账号查看权限 | 打开工作台侧栏 / 直接访问 `/service/account` | 侧栏隐藏账号管理；直连显示框架内无权限页 | |
| P1-06 | 资源详情跳账号 | 打开 VPC / 子网 / 硬盘 / IP / 网络接口 / 路由表 / 安全组任一详情 | 点账号名跳转 | 进入 `/service/account/details/<id>/basic`，不再拼 `/resource/account` | |
| P1-07 | 收藏与旧键 | 曾收藏过账号管理（旧符号） | 打开收藏区 | 无孤儿项、控制台无报错；重新收藏后可用 | |

### P2 - UI 细节（按需）

- [ ] 列表 `fit` 宽度策略保持（`layout.contentWidth: 'fit'` 未丢）
- [ ] English 下「账号管理」「账号登记」「账号详情」文案正常
- [ ] 账号登记页面包屑隐藏（沿用 `breadcrumb.show=false`）

## 验证结论

> 每条用例填写：PASS / FAIL / Skipped + 简要说明（FAIL 必须附复现步骤、错误截图链接或日志片段）

| ID | 结果 | 备注 |
|----|------|------|
| P0-01 | PASS | 账号管理在个人事务、单据管理之后；公共事务只剩资源选型、权限策略库 |
| P0-02 | PASS | `/service/account`，顶栏与侧栏高亮正确 |
| P0-03 | PASS | 列表无资源账号，请求含 `type neq resource` |
| P0-04 | PASS | 账号类型筛选只有登记账号、安全审计账号 |
| P0-05 | PASS | 叠加搜索后默认过滤仍生效 |
| P0-06 | PASS | 详情地址 `/service/account/details/<id>/basic`，返回正常 |
| P0-07 | PASS | 登记页 `/service/account/check-in`，提交后回列表且可见 |
| P1-01 | PASS | `/resource/account` 转列表 |
| P1-02 | PASS | `/resource/account/details/<id>` 与 `/resource/account/details/<id>/basic?_f=1` 均转详情；后者最初 404（旧地址多一段页签路径未匹配到 legacy 路由），已补 `:accountId/:tab` 规则修复 |
| P1-03 | PASS | `/resource/account/check-in`、`/resource/account/add` 转登记 |
| P1-04 | PASS | `/resource/resource/account` 仍进资源纳管账号状态侧滑 |
| P1-05 | PASS | 无权限时菜单隐藏、直连显示无权限页 |
| P1-06 | PASS | 资源详情 7 处账号名跳 `/service/account/details/<id>/basic`（资源账号，与 PR 一致） |
| P1-07 | PASS | 旧收藏自动清理、无孤儿项 |

- 执行日期: 2026-10-27
- 总体结论: PASS
- 后续行动: 文档收口 → 推进 done；资源账号经工作台详情打开属 PR 现状，权限收紧时的断链约束已记录在权限单描述 4.3/4.4
