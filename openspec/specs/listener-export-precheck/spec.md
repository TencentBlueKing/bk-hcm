# listener-export-precheck Specification

## Purpose

定义导出监听器及其下属资源（URL 规则、RS）的预检能力：包含预检接口契约与 IAM 鉴权要求、始终执行的监听器归属正确性校验、按勾选负载均衡数量分流的数量类限制规则（参数层与业务层共用同一阈值口径），以及导出链路按监听器 ID 查询的分批要求。目标是在防止「批量勾选大量负载均衡」造成超大导出的同时，不误伤「少量负载均衡精确导出」的场景。

## Requirements

### Requirement: 导出监听器预检接口契约

系统 SHALL 提供 `POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/listeners/export/pre_check` 接口，接收待导出的负载均衡与监听器范围，返回本次导出是否可以执行。

请求体中 `listeners` 为数组，每个元素包含 `lb_id`（必填）与 `lbl_ids`（可选）；`lbl_ids` 为空表示导出该负载均衡下的全部监听器。`only_export_listener` 为可选布尔值，为 `true` 表示仅导出监听器信息，不导出 URL 规则与 RS。

支持的云厂商为腾讯云（`tcloud`）与腾讯自研云（`tcloud-ziyan`），其余厂商不支持导出。

#### Scenario: 预检通过

- **GIVEN** 请求参数合法且所有校验项均通过
- **WHEN** 调用预检接口
- **THEN** 系统 SHALL 返回 `{"pass": true, "reason": ""}`，HTTP 状态为成功

#### Scenario: 预检不通过

- **GIVEN** 存在任一校验项未通过
- **WHEN** 调用预检接口
- **THEN** 系统 SHALL 返回 `{"pass": false, "reason": "<不通过原因>"}`，且 `reason` 为可直接展示给用户的中文描述

#### Scenario: 不支持的云厂商

- **WHEN** 路径参数 `vendor` 不是 `tcloud` 或 `tcloud-ziyan`
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

#### Scenario: 导出接口复用同一预检逻辑

- **GIVEN** 调用方跳过预检接口直接调用 `POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/listeners/export`
- **WHEN** 该请求的范围未通过预检规则
- **THEN** 导出接口 SHALL 以相同规则拦截并返回错误，不产生导出文件

### Requirement: 预检需通过 IAM 鉴权

系统 SHALL 在执行预检前，基于请求中所有 `lb_id` 查询负载均衡基础信息，并通过蓝鲸 IAM 校验调用方在该业务下对负载均衡的 Update 权限。

#### Scenario: 调用方无权限

- **WHEN** 调用方在目标业务下不具备负载均衡的 Update 权限
- **THEN** 系统 SHALL 返回无权限错误，不执行后续任何校验

### Requirement: 勾选负载均衡数量决定是否执行数量类限制

系统 SHALL 以请求中所有 `listeners` 元素的 `lb_id` 去重后的数量作为「勾选负载均衡数量」，并据此决定是否执行数量类限制：

- 勾选负载均衡数量小于或等于阈值（默认 5）时，系统 SHALL 跳过全部数量类限制，包括请求参数层的数量校验与业务层的资源数量校验
- 勾选负载均衡数量大于阈值时，系统 SHALL 执行全部数量类限制

参数层与业务层 MUST 使用同一个阈值口径与同一个常量。阈值 MUST 以命名常量定义，MUST NOT 使用魔数。将阈值配置为 `0` 时判断条件恒不成立，系统行为 SHALL 等价于全量执行数量限制。

该分流规则 SHALL 对两类请求形态统一生效，不区分 `listeners` 元素是否携带 `lbl_ids`。

#### Scenario: 勾选 5 个负载均衡且监听器数量超过原限制

- **GIVEN** 请求中去重后的负载均衡数量为 5
- **AND** 这些负载均衡下的四层监听器数量为 8000，超过原有的 5000 限制
- **WHEN** 执行预检
- **THEN** 系统 SHALL 返回 `pass: true`

#### Scenario: 勾选 1 个负载均衡且 RS 数量极大

- **GIVEN** 请求中去重后的负载均衡数量为 1
- **AND** 该负载均衡下的七层 RS 数量为 50000
- **WHEN** 执行预检
- **THEN** 系统 SHALL 返回 `pass: true`

#### Scenario: 勾选 6 个负载均衡且各项数量均未超限

- **GIVEN** 请求中去重后的负载均衡数量为 6
- **AND** 四层监听器、七层监听器、七层 URL 规则、四层 RS、七层 RS 数量均未超过各自限制
- **WHEN** 执行预检
- **THEN** 系统 SHALL 返回 `pass: true`

#### Scenario: 勾选 6 个负载均衡且监听器数量超限

- **GIVEN** 请求中去重后的负载均衡数量为 6
- **AND** 这些负载均衡下的四层监听器数量为 6000
- **WHEN** 执行预检
- **THEN** 系统 SHALL 返回 `pass: false`，`reason` 说明四层监听器数量及其限制值

#### Scenario: 同一负载均衡被拆分为多个请求元素

- **GIVEN** 请求的 `listeners` 包含 6 个元素，但其中 `lb_id` 去重后仅为 3 个
- **WHEN** 执行预检
- **THEN** 系统 SHALL 按 3 个负载均衡参与阈值判断，跳过数量类限制

#### Scenario: 记录勾选数量便于观察

- **WHEN** 因勾选负载均衡数量不超过阈值而跳过数量限制
- **THEN** 系统 SHALL 输出包含勾选负载均衡数量、监听器 ID 数量与 `rid` 的 Info 级别日志

### Requirement: 请求参数层校验按性质区分

系统 SHALL 把请求参数层的校验分为正确性校验与数量类校验两类，前者始终执行，后者参与勾选负载均衡数量分流。

正确性校验（两种情况下都执行）：

- `listeners` 数组 MUST 非空
- 每个 `listeners` 元素的 `lb_id` MUST 非空

数量类校验（仅在勾选负载均衡数量大于阈值时执行）：

- `listeners` 数组长度 MUST 不超过 5000
- 所有元素的 `lbl_ids` 去重合并后的总数 MUST 不超过 5000
- 单个 `listeners` 元素的 `lbl_ids` 长度 MUST 不超过 100

#### Scenario: listeners 为空

- **WHEN** 请求体的 `listeners` 为空数组
- **THEN** 系统 SHALL 返回 InvalidParameter 错误，与勾选负载均衡数量无关

#### Scenario: lb_id 为空

- **WHEN** 任一 `listeners` 元素的 `lb_id` 为空字符串
- **THEN** 系统 SHALL 返回 InvalidParameter 错误，与勾选负载均衡数量无关

#### Scenario: 单个负载均衡下勾选超过 100 个监听器

- **GIVEN** 请求的 `listeners` 只含 1 个元素，去重后负载均衡数量为 1
- **AND** 该元素的 `lbl_ids` 包含 3000 个监听器 ID
- **WHEN** 调用预检接口
- **THEN** 系统 SHALL 不因数量校验拒绝该请求，并返回 `pass: true`

#### Scenario: 勾选负载均衡数量超过阈值时单元素 lbl_ids 超过 100

- **GIVEN** 请求中去重后的负载均衡数量为 6
- **AND** 某个 `listeners` 元素的 `lbl_ids` 长度为 200
- **WHEN** 调用预检接口
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

#### Scenario: 勾选负载均衡数量超过阈值时 lbl_ids 总数超过 5000

- **GIVEN** 请求中去重后的负载均衡数量为 60
- **AND** 所有元素的 `lbl_ids` 去重后总数为 6000
- **WHEN** 调用预检接口
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

### Requirement: 监听器归属正确性校验始终执行

对于携带了 `lbl_ids` 的每个 `listeners` 元素，系统 SHALL 校验这些监听器确实属于该元素的 `lb_id`。该校验 MUST 在任何数量限制判断之前执行，且 MUST NOT 因勾选的负载均衡数量少而被跳过。

#### Scenario: 监听器不属于指定的负载均衡

- **GIVEN** 请求中某个 `lbl_id` 在数据库中不存在，或其所属负载均衡与同元素的 `lb_id` 不一致
- **WHEN** 执行预检
- **THEN** 系统 SHALL 返回 InvalidParameter 错误，提示该监听器不属于该负载均衡

#### Scenario: 勾选负载均衡数量少时仍执行归属校验

- **GIVEN** 请求中去重后的负载均衡数量不超过跳过限制阈值
- **AND** 请求中存在不属于对应负载均衡的 `lbl_id`
- **WHEN** 执行预检
- **THEN** 系统 SHALL 返回归属校验失败的错误，而非放行

### Requirement: 按监听器 ID 查询必须分批以避免触碰下游上限

data-service 侧监听器、URL 规则、目标组监听器关联、RS 四张表的 DAO 均限制单个 `IN` 条件最多 10000 个元素。参数层数量校验被跳过后，请求携带的监听器 ID 数量可能超过该上限，因此 cloud-server 在按监听器 ID 或负载均衡 ID 查询时 MUST 分批，单批元素数量 MUST 不超过 500。

拆分为多次查询后，系统 SHALL 保证合并结果与单次查询在语义上一致，MUST NOT 产生重复记录。

#### Scenario: 监听器 ID 数量超过下游 IN 上限

- **GIVEN** 请求中去重后的负载均衡数量为 1
- **AND** 该元素的 `lbl_ids` 包含 12000 个监听器 ID
- **WHEN** 执行导出
- **THEN** 系统 SHALL 分批查询并正常产出导出文件，MUST NOT 返回 `at most have 10000 elements` 一类的底层错误

#### Scenario: 混合传参时目标组监听器关联不重复

- **GIVEN** 请求同时包含一个只带 `lb_id` 的元素与另一个带 `lbl_ids` 的元素
- **AND** 后者的监听器属于前者的负载均衡范围之外
- **WHEN** 执行导出
- **THEN** 导出文件中的 RS 记录 SHALL 与合并前一致，同一条目标组监听器关联 MUST NOT 被重复计入

#### Scenario: 同一关联同时命中两个查询维度

- **GIVEN** 请求中某个负载均衡以 `lb_id` 形式出现，同时其下某个监听器以 `lbl_ids` 形式出现在另一个元素中
- **WHEN** 执行导出
- **THEN** 系统 SHALL 按关联记录 ID 去重，该监听器绑定的 RS 在导出文件中只出现一次

### Requirement: 勾选负载均衡数量超过阈值时的数量限制项

当勾选负载均衡数量大于阈值时，系统 SHALL 依次执行以下数量限制，任一项超限即终止并返回该项的错误。各限制项之间以及同一限制项内的四层与七层之间均为独立判断，MUST NOT 将四层与七层的数量合并统计。

| 限制项 | 统计范围 | 默认上限 |
|---|---|---|
| 四层监听器数量 | 协议为 TCP / UDP / TCP_SSL / QUIC 的监听器 | 5000 |
| 七层监听器数量 | 协议为 HTTP / HTTPS 的监听器 | 5000 |
| 七层 URL 规则数量 | 规则类型为七层的 URL 规则，按负载均衡维度与监听器维度分别统计后累加 | 5000 |
| 四层 RS 数量 | 绑定状态为成功且监听器规则类型为四层的 RS | 5000 |
| 七层 RS 数量 | 绑定状态为成功且监听器规则类型为七层的 RS | 5000 |

#### Scenario: 四层与七层数量各自未超限但合计超过单项上限

- **GIVEN** 勾选负载均衡数量大于阈值
- **AND** 四层监听器数量为 4000、七层监听器数量为 4000
- **WHEN** 执行预检
- **THEN** 系统 SHALL 返回 `pass: true`，因为两者分别与各自的 5000 上限比较

#### Scenario: 仅导出监听器时跳过规则与 RS 限制

- **GIVEN** 勾选负载均衡数量大于阈值
- **AND** 请求的 `only_export_listener` 为 `true`
- **AND** 四层与七层监听器数量均未超限，但七层 URL 规则数量为 8000
- **WHEN** 执行预检
- **THEN** 系统 SHALL 返回 `pass: true`，不执行 URL 规则与 RS 的数量限制

#### Scenario: 七层 URL 规则数量超限

- **GIVEN** 勾选负载均衡数量大于阈值
- **AND** `only_export_listener` 为 `false` 或未传
- **AND** 七层 URL 规则数量为 6000
- **WHEN** 执行预检
- **THEN** 系统 SHALL 返回 `pass: false`，`reason` 说明规则数量及其限制值
