# feat-menu-phase2-resource-detail 测试清单

> 工作流: `feat-menu-phase2-resource-detail`
> 关联文档: Coding（`coding.md`）、需求 `docs/reqs/二期资源详情迁移.md`（AC 编号对应）

## 验证范围

- 功能边界: 资源详情挂到资源纳管下 `/resource/manage/<resourceType>/details/<id>`（左侧云账号树、右栏面包屑与操作按钮）；详情按钮统一 `hcm-auth`（资源侧按账号、业务侧按业务）；旧地址 `/resource/detail/:type` 转址；全站进入资源详情的链接改到新地址；撤掉旧详情路由与右栏过渡面包屑；GCP 防火墙、安全组模板返回安全组页签对应子场景。
- 不在范围: 业务地址 `/business/:bizId`（`?bizs=` 保留）；业务列表页改造；账号详情链接已由账号管理单改写，本单只验收；主机回收后跳转目标（沿用现网，见后续行动）。

## 测试环境

- 前端入口: `<本地开发或测试环境地址，由执行人确认>`
- 测试账号: `<test_account>`（资源纳管全部权限）；`<test_account_partial>`（只对账号 A 有 `iaas_resource_operate` / `iaas_resource_delete`，对账号 B 无）；`<test_biz_account>`（业务资源权限，用于业务详情回归）；`<test_biz_account_readonly>`（业务有查看、无 `biz_iaas_resource_operate` / `biz_iaas_resource_delete`）
- 数据准备: 账号 A 下主机、VPC、子网、安全组（腾讯云）、GCP 防火墙、安全组模板、硬盘、弹性 IP、路由表、镜像、网络接口各至少一个，其中主机 / VPC / 子网 / 弹性 IP 有未分配业务的实例

## 用例清单

### P0 - 主流程（必测）

| ID | 场景 | 前置 | 操作 | 期望 | covered-by | 回滚 |
|----|------|------|------|------|------------|------|
| P0-01 | 进入详情（AC-001） | 资源纳管选中账号 A，主机页签 | 点一台主机 | 地址 `/resource/manage/host/details/<ID>?accountId=<A>&vendor=<厂商>&_f=1`；顶栏高亮资源运营、侧栏高亮资源纳管；左侧云账号树仍选中 A；右栏面包屑显示「主机详情 - ID <ID>」、返回按钮与开机 / 关机等按钮 | | — |
| P0-02 | 返回（AC-002） | P0-01 | 点面包屑返回 | 回到主机页签，账号仍是 A；再从弹性 IP、路由表、安全组页签进入详情后返回，各自回到原页签 | | — |
| P0-03 | 各类详情内容（AC-003） | 账号 A | 依次从 VPC、子网、安全组、GCP 防火墙、模板、硬盘、弹性 IP、路由表、镜像、网络接口页签进详情 | 页签、字段、标题文案与改前一致；地址类型分别为 `vpc`、`subnet`、`security`、`gcp`、`template`、`drive`、`eip`、`routing`、`image`、`network-interface` | | — |
| P0-04 | 详情内互跳 | 主机详情 | 云硬盘页签点硬盘 ID；弹性 IP 详情点所属主机 / 网络接口；路由表详情点所属网络；硬盘详情点挂载主机；关联资源按钮里点 VPC / 子网 / 镜像 | 都进入对应新详情，页面内容随 ID 刷新（不残留上一条数据）；返回逐级回到上一详情 | | — |
| P0-05 | 操作结果（AC-004） | `<test_account>`，未分配业务的实例 | 主机开机 / 关机 / 分配；VPC、子网删除；弹性 IP 绑定 / 解绑；主机云硬盘挂载 / 卸载、弹性 IP 解绑、安全组解绑；安全组新增 / 编辑 / 删除规则、关联实例绑定 / 解绑、资产归属编辑 | 操作结果与改前一致 | | 按操作回滚 |
| P0-06 | 按账号鉴权（AC-004、AC-S01） | `<test_account_partial>` | 进账号 A、账号 B 的同类详情看按钮 | A 下按钮可用；B 下按钮置灰带锁，点击弹权限申请；Network 里鉴权合并为一次请求，`account_id` 返回前不发鉴权请求 | | — |
| P0-07 | 旧地址转址（AC-005） | — | 打开 `#/resource/detail/vpc?type=tcloud&id=<ID>`、`#/resource/detail/eips?id=<ID>`、`#/resource/detail/route?id=<ID>` | 依次进入 `/resource/manage/vpc/details/<ID>?vendor=tcloud`、`/resource/manage/eip/details/<ID>`、`/resource/manage/routing/details/<ID>`，均打开对应详情；再直接打开 `#/resource/manage/ip`、`#/resource/manage/eips/details/<ID>`、`#/resource/manage/route/details/<ID>`，分别进弹性 IP 页签、弹性 IP 详情、路由表详情 | | — |
| P0-08 | 旧地址缺 id / 未知类型（AC-006） | — | 打开 `#/resource/detail/host`、`#/resource/detail/eips`、`#/resource/detail/gcp?accountId=<A>`、`#/resource/detail/unknown?id=1` | 依次进入主机列表、弹性 IP 列表、安全组页签的 GCP 防火墙子场景（账号 A）、「资源类型不存在」提示 | | — |
| P0-09 | 业务目标组 VPC 链接（AC-007） | 业务负载均衡目标组详情 | 点「所属vpc」 | 新窗口打开 `…#/resource/manage/vpc/details/<ID>?vendor=tcloud` | | — |
| P0-10 | 账号链接（AC-008） | 任一资源详情 | 点「账号」字段 | 进入账号管理该账号详情 | | — |
| P0-11 | 业务详情回归（AC-009） | `<test_biz_account>` | 业务下打开主机、VPC、子网、安全组、硬盘、弹性 IP、路由表、镜像、网络接口详情，做 P0-05 中的操作 | 标题、按钮可用状态、操作结果、内部跳转与改前一致 | | 按操作回滚 |
| P0-14 | 业务详情新地址与旧地址（第 10 轮） | `<test_biz_account>` | 业务主机 / VPC / 子网 / 安全组 / 硬盘 / 弹性 IP / 路由表 / 镜像 / 网络接口列表点 ID；安全组页签切 GCP 防火墙、参数模板点 ID；再打开旧地址 `#/business/host/detail?id=<ID>&bizs=<B>`、`#/business/gcp/detail?id=<ID>&bizs=<B>` | 新地址为 `/business/<类型>/detail/<ID>?bizs=<B>`；旧地址转成路径参数且其余参数保留；页面为对应详情（不会串成别的类型）；面包屑返回回到对应列表 | | — |
| P0-15 | 业务侧站内详情链接 | `<test_biz_account>` | 主机详情云硬盘页签点硬盘、硬盘详情点挂载主机、弹性 IP 详情点主机 / 网络接口、路由表详情点所属网络、关联资源按钮点 VPC / 子网 / 镜像、业务回收站「查看详情」、负载均衡 RS 列表点 IP（新窗口）、GCP 网络接口「已关联到」 | 都进入对应业务详情新地址，带 `?bizs=`；返回逐级回到上一页 | | — |
| P0-16 | 详情页切换业务 | 任一业务详情（含 GCP 防火墙、参数模板） | 顶部切换业务 | 进入新业务下该资源的列表，不报 404 | | — |
| P0-13 | 业务详情无操作权限 | `<test_biz_account_readonly>` | 同 P0-11 逐类看按钮 | 按钮置灰带锁，点击弹权限申请（申请项为业务 IaaS 编辑 / 删除）；Network 鉴权合并为一次请求；控制台无报错 | | — |
| P0-12 | 回收管理查看详情 | 资源运营 → 回收管理 | 主机 / 硬盘页签点「查看详情」 | 进入对应新详情；返回回到回收管理 | | — |

### P1 - 异常分支（必测）

| ID | 场景 | 前置 | 操作 | 期望 | covered-by |
|----|------|------|------|------|------------|
| P1-01 | GCP 防火墙 / 模板返回 | GCP 防火墙或模板详情 | 点返回；在新窗口直接打开该详情后点返回 | 回到安全组页签对应子场景（GCP 防火墙 / 参数模板） | |
| P1-02 | 详情页切换账号 | 主机详情 | 在云账号树点另一个账号；再点「全部账号」 | 回到主机列表并选中新账号 / 全部账号；GCP 防火墙详情里切账号回到安全组页签 | |
| P1-03 | 非法类型 / ID | — | 打开 `#/resource/manage/foo/details/1`；`#/resource/manage/host/details/not-exist` | 前者显示「资源类型不存在」与「返回资源纳管」按钮；后者由接口错误提示，页面不白屏 | |
| P1-04 | 旧代码残留 | 代码检查 | `rg "'resourceDetail'" src`；`rg "legacyBreadcrumb\|legacy-breadcrumb" src` | 均为 0 | |

### P2 - UI 细节（按需）

- [ ] 1366px 宽度下各类详情无多余空白，内容不被云账号树遮挡；侧栏展开 / 折叠都正常
- [ ] 右栏面包屑按钮区控制台无 Teleport 警告；其他云厂商主机不显示操作按钮

## 验证结论

> 每条用例填写：PASS / FAIL / Skipped + 简要说明（FAIL 必须附复现步骤、错误截图链接或日志片段）

| ID | 结果 | 备注 |
|----|------|------|
| P0-01 ~ P0-16 | | |
| P1-01 ~ P1-04 | | |

- 执行日期:
- 总体结论:
- 后续行动: 资源侧主机回收后仍跳业务回收站地址（现网与 PR 同），归「二期收尾体验」F-010
