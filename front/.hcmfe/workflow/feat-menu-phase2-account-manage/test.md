# feat-menu-phase2-account-manage 测试清单

> 工作流: `feat-menu-phase2-account-manage`
> 关联文档: Coding（`coding.md`）、需求 `docs/reqs/二期账号管理菜单.md`（AC 编号对应）

## 验证范围

- 功能边界: 资源运营 → 云资源 → 账号管理（列表、详情 `/basic`、登记）；按实例鉴权的编辑 / 删除；全部账号旧地址转址；资源详情与关联资源按钮里的账号链接改到新详情；菜单加入侧栏并按查看权限显隐；旧账号页面、旧账号信息页签与 `DetailHeader` 删除。
- 不在范围: 账号登记页本身（复用服务请求原组件，只回归入口）；三期迁工作台与「不展示资源账号」。

## 测试环境

- 前端入口: `<本地开发或测试环境地址，由执行人确认>`
- 测试账号: `<test_account>`（有账号查看 / 编辑 / 删除权限）；`<test_account_noperm>`（无账号查看权限）；`<test_account_readonly>`（有查看、无编辑与删除）
- 数据准备: 资源账号、登记账号、安全审计账号各至少一个；一个可删除（无关联资源）的测试账号

## 用例清单

### P0 - 主流程（必测）

| ID | 场景 | 前置 | 操作 | 期望 | covered-by | 回滚 |
|----|------|------|------|------|------------|------|
| P0-01 | 侧栏与列表（AC-001） | 全部权限 | 展开资源运营侧栏，点账号管理 | 云资源组顺序为资源纳管、账号管理、回收管理；地址 `/resource/account`；顶部面包屑「账号管理」；无云账号树；列表含三类账号 | | — |
| P0-02 | 搜索（AC-002） | P0-01 | 按名称、账号类型（三类）、云厂商、负责人搜索；勾「精确」 | 结果与改前「全部账号」列表一致 | | 清空搜索 |
| P0-03 | 进入详情与编辑（AC-003） | P0-01 | 点账号名称；返回；再点「编辑」 | 都进入 `/resource/account/details/<ID>/basic`；只有「基本信息」页签；面包屑「账号详情：ID（<ID>）」可返回列表；名称、账号归属、密钥信息可编辑并保存，保存后刷新仍是新值 | | 改回原值 |
| P0-04 | 编辑账号弹窗 | P0-03 | 点「账号归属」的编辑，改负责人与备注并确定 | 请求为 `PATCH /api/v1/cloud/accounts/<ID>`（不是 `undefined`）；负责人选择器为人员选择组件；保存成功 | | 改回原值 |
| P0-05 | 删除（AC-005） | 可删除的测试账号 | 点「删除」；再对有关联资源的账号点「删除」 | 校验通过后二次确认、删除成功并刷新列表；校验不通过时只出后端原因提示，不弹确认框 | | — |
| P0-06 | 账号登记（AC-006、AC-007、AC-013） | — | 打开 `#/resource/account/check-in`；再打开 `#/service/service-apply/account-add?vendor=tcloud`、`#/resource/account/add` | 都进入 `/resource/account/check-in`，参数保留；顶栏高亮资源运营、侧栏高亮账号管理；提交后仍跳单据管理 | | — |
| P0-07 | 账号信息 / 资源状态旧地址（AC-008、AC-010） | — | 打开 `#/resource/resource/account`；`#/resource/resource/account?accountId=<A>`；`#/resource/resource/account/resource?accountId=<A>` | 第一个进 `/resource/account`；后两个进资源纳管主机列表、选中账号 A，并自动打开账号状态侧滑，地址里不留 `showAccountStatus` | | — |
| P0-08 | 旧详情地址（AC-009、AC-011） | — | 打开 `#/resource/account/detail?accountId=<A>`、`#/resource/account/detail?id=<A>`、`#/resource/resource/account/detail?accountId=<A>`、`#/resource/resource/account/manage?accountId=<A>`；再打开不带账号的 `#/resource/account/detail` | 前四个进 `/resource/account/details/<A>/basic`；不带账号时进账号列表 | | — |
| P0-09 | 资源详情里的账号链接 | 资源纳管选账号 A，进 VPC / 子网 / 硬盘 / 弹性 IP / 路由表 / 网络接口的详情 | 点「账号」字段链接 | 新窗口或当前页进入 `/resource/account/details/<A>/basic` | | — |
| P0-10 | 关联资源按钮里的账号跳转 | 主机详情等使用 `useRouteLinkBtn` 展示账号的地方（资源侧、业务侧各一处） | 点账号名 | 进入账号管理的账号详情，不报路由错误 | | — |

### P1 - 异常分支（必测）

| ID | 场景 | 前置 | 操作 | 期望 | covered-by |
|----|------|------|------|------|------------|
| P1-01 | 无查看权限（AC-012） | `<test_account_noperm>` | 看侧栏；直接打开 `#/resource/account`、`#/resource/account/details/<A>/basic` | 侧栏无账号管理；两个地址都显示框架内无权限页 | |
| P1-02 | 无编辑 / 删除权限（AC-004、AC-S01） | `<test_account_readonly>` | 列表看「编辑」「删除」；详情看三处「编辑」 | 全部置灰；列表按钮点击弹权限申请；整页鉴权请求合并为一次 | |
| P1-03 | 账号不存在 | — | 打开 `#/resource/account/details/not-exist/basic` | 显示「账号不存在」与「返回账号列表」按钮，点击回到列表 | |

### P2 - UI 细节（按需）

- [ ] 列表外层留白、白底容器与其它资源运营页面一致；表格高度撑满、分页在底部
- [ ] 详情页签头与内容区之间无空隙

## 验证结论

> 每条用例填写：PASS / FAIL / Skipped + 简要说明（FAIL 必须附复现步骤、错误截图链接或日志片段）

| ID | 结果 | 备注 |
|----|------|------|
| P0-01 ~ P0-10 | | |
| P1-01 ~ P1-03 | | |

- 执行日期:
- 总体结论:
- 后续行动:
