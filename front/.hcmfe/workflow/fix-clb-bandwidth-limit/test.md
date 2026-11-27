# CLB 规格带宽联动验证

## 范围和环境

- 逻辑检查执行日期：2026-10-19；2026-10-20 经用户确认进入 test 阶段。
- 进入 test 时复核：watch 回调新增 max: number 类型标注，运行逻辑未变；既有检查结果保留为 2026-10-19 的执行记录，浏览器用例仍待验证。
- 验证购买表单带宽范围、刻度、规格切换及配置条目编辑；不涉及后端、计费和独占集群规则。
- 本仓库没有 node_modules，借用 <EXISTING_DEPENDENCIES>/node_modules 的 TypeScript、Vue 和检查工具，不安装或提交依赖。
- 功能验证从本次实际源码截取带宽相关声明，并转译执行；使用真实 Vue reactive / computed / watch / nextTick 和本地规格常量。此检查不等同浏览器或完整表单集成测试。

## 已执行的定向逻辑检查

| ID | 场景和输入 | 期望 | covered-by | 结果 |
| --- | --- | --- | --- | --- |
| P0-01 | 初始 shared，带宽 100 | 上限 10240，值 100 | unit（临时源码执行检查） | PASS |
| P0-02 | 确认 clb.c4.xlarge / 61440 | 上限 61440，值不放大 | unit | PASS |
| P0-03 | 带宽 61440 后切回 shared | 上限和值均为 10240 | unit | PASS |
| P0-04 | 再切大规格 | 上限 61440，值仍为 10240 | unit | PASS |
| P0-05 | 编辑未缓存的 clb.c1.small 条目，带宽 1024 | 使用该规格常量上限 1024，不沿用旧值 | unit | PASS |
| P0-06 | 切 clb.c4.large / 40960，原带宽 1024 | 上限 40960，保留 1024 | unit | PASS |
| P1-01 | 带宽 40000 后将 sla_type 清空 | 上限和值收敛为 10240 | unit | PASS |
| P1-02 | 未知且未缓存的规格 | 上限兜底 10240 | unit | PASS |
| P1-03 | 确认自定义规格上限 81920 | 缓存值优先，上限 81920，保留带宽 | unit | PASS |
| P1-04 | 从自定义规格切回已缓存的 clb.c4.xlarge | 上限恢复 61440 | unit | PASS |

上述十项均额外断言刻度为 1 / max÷4 / max÷2 / max。

## 浏览器待验证清单

2026-10-20 用户选择手动自测，并明确跳过本次 E2E。当前缺少 E2E 配置和安装清单，不安装工具层；下表 Skipped 表示尚未执行，等待用户反馈后逐项回填，不代表验收通过。工作流保持 test。

准备：有购买权限的测试账号、可用地域、共享型与至少两个不同上限的性能容量型规格；业务视角与工作台入口各检查一次。

| ID | 操作 | 期望 | covered-by | 结果 |
| --- | --- | --- | --- | --- |
| UI-P0-01 | 选择大规格、填满带宽、切回共享型 | 滑块和输入框均显示 10240，刻度同步 | manual | Skipped：未执行浏览器 |
| UI-P0-02 | 清单中先编辑大规格 A，再编辑小规格 B | B 的上限和值匹配 B，不残留 A 的范围 | manual | Skipped：未执行浏览器 |
| UI-P1-01 | 切换网络或运营商导致规格重置 | 共享/空规格上限恢复 10240，超限值收敛 | manual | Skipped：未执行浏览器 |
| UI-P1-02 | 输入超过当前上限的值并失焦 | 输入框与滑块范围一致 | manual | Skipped：未执行浏览器 |
| UI-P1-03 | 检查提交请求中的 internet_max_bandwidth_out | 与表单一致且不超过当前规格上限 | manual | Skipped：未发起资源创建 |
| UI-P2-01 | 检查两入口及不同刻度布局 | 无重叠或截断，两入口行为一致 | manual | Skipped：未执行浏览器 |

## 工具检查

- bkdevbuddy lint --fix：未完成，命令找不到 prettier；不采用 CLI 的成功摘要作为通过证据。
- Prettier：PASS，借用已有安装，通过 Prettier API 显式加载本仓库配置执行 format / check。CLI 的 JSX 闭括号格式与现有代码不同，已恢复无关格式变动，最终仅保留功能修复。
- ESLint：PASS，借用已有安装，使用本仓库 .eslintrc.js 和 --no-cache 定向检查（exit 0）。
- git diff --check：PASS。
- req drift：现有 Requirement 均 hasDrift=false；本工作流未绑定 Requirement。
- 尚未运行完整构建、类型检查、E2E 或浏览器验收。

## 2026-10-20 未知规格回归

- 定向 ESLint（两个改动文件）、Prettier API、git diff --check 均通过；bkdevbuddy lint 仍因缺少命令不可用，使用既有相邻仓库工具安装验证。

- PASS（临时源码执行检查）：从用户响应取四个运营商的规格列表，经既有 c1 过滤及弹框 shared 过滤后每组 9 行；c5 两行均保留。
- PASS：执行实际列 render，c5 名称显示 SpecType；四个缺失数值列显示 --，已知规格名称不变，数值 0 正常显示。
- PASS：执行实际确认处理器，c5 返回原始 slaType 和 undefined 带宽，不抛异常；空规格列表输入不抛异常。
- 手动待验证：打开弹框显示 c5 行，选择后预览显示原始规格标识；表单上限按现有策略兜底 10240。未进行浏览器验收，不视作真实云端上限校验。
