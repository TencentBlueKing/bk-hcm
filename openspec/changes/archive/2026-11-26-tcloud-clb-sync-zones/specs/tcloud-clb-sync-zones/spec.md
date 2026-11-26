# Spec Delta

## Purpose

规定腾讯云负载均衡（CLB）在资源同步新建、差异比对与存量更新过程中，按云上返回的数据提取、校验并回填可用区（`zones`）字段的行为契约，确保随机分配可用区及存量缺失可用区的实例能够准确落库并保持已有数据安全。

## ADDED Requirements

### Requirement: 按网络类型提取云上可用区并过滤空值

系统在同步腾讯云负载均衡时，SHALL 按实例的网络类型从云上返回数据中提取可用区列表：公网实例（`OPEN`）从云上主可用区字段提取单元素可用区列表，内网实例（`INTERNAL`）从云上可用区列表字段按原顺序提取。系统 SHALL 过滤所有空指针与空字符串；当云上未返回任何非空可用区时，提取结果 SHALL 为空（`nil`），不得产出仅包含空字符串的元素。

#### Scenario: 公网实例返回非空主可用区
- **WHEN** 云上公网负载均衡实例的主可用区为 `"ap-shanghai-2"`
- **THEN** 系统提取出的可用区列表 SHALL 为 `["ap-shanghai-2"]`

#### Scenario: 公网实例主可用区为空或空字符串
- **WHEN** 云上公网负载均衡实例的主可用区对象为 `null` 或可用区字符串为 `""`
- **THEN** 系统提取出的可用区列表 SHALL 为空（`nil`），不包含 `""`

#### Scenario: 内网实例返回可用区列表且含空值
- **WHEN** 云上内网负载均衡实例的可用区列表为 `["ap-guangzhou-3", "", "ap-guangzhou-4"]`
- **THEN** 系统提取出的可用区列表 SHALL 过滤空值并保持顺序，结果为 `["ap-guangzhou-3", "ap-guangzhou-4"]`

#### Scenario: 内网实例未返回有效可用区
- **WHEN** 云上内网负载均衡实例的可用区列表为 `null` 或全部为空字符串
- **THEN** 系统提取出的可用区列表 SHALL 为空（`nil`）

### Requirement: 同步新建与更新按云上数据回填可用区

系统在同步新建本地负载均衡记录时 SHALL 写入提取出的可用区列表。在同步存量记录时，若云上提取出的可用区列表非空且与本地记录的可用区列表不一致，系统 SHALL 判定该记录发生变更并将本地记录的可用区更新为云上值；若云上提取出的可用区列表为空，系统 SHALL 不因可用区字段判定变更，且在更新时 SHALL 保持本地已有的可用区值不变。

#### Scenario: 新建记录写入云上可用区
- **WHEN** 本地不存在该负载均衡记录，且云上实例返回有效可用区 `["ap-shanghai-2"]`
- **THEN** 同步新建的本地记录 `zones` 字段 SHALL 为 `["ap-shanghai-2"]`

#### Scenario: 存量记录缺失可用区时回填
- **WHEN** 本地记录的 `zones` 为空（或为 `[""]`），而云上实例返回有效可用区 `["ap-guangzhou-3"]`
- **THEN** 同步 SHALL 判定该记录有变更，并将本地记录的 `zones` 更新为 `["ap-guangzhou-3"]`

#### Scenario: 存量记录可用区与云上不一致时更新
- **WHEN** 本地记录的 `zones` 为 `["ap-guangzhou-1"]`，而云上实例返回有效可用区 `["ap-guangzhou-3"]`
- **THEN** 同步 SHALL 判定该记录有变更，并将本地记录的 `zones` 更新为 `["ap-guangzhou-3"]`

#### Scenario: 云上未返回可用区时不覆盖本地已有值
- **WHEN** 本地记录的 `zones` 为 `["ap-guangzhou-3"]`，而云上实例本次未返回有效可用区（提取结果为空）
- **THEN** 同步 SHALL 不因可用区判定该记录有变更，且更新请求 SHALL 不覆盖或清空本地已有的 `["ap-guangzhou-3"]`

### Requirement: 数据服务批量更新接口支持可选可用区字段

数据服务（`data-service`）的负载均衡批量更新接口 SHALL 支持可选的 `zones` 字段。当请求中传入非空 `zones` 列表时，系统 SHALL 更新对应负载均衡记录的 `zones` 列；当请求中未传入 `zones`（为空）时，系统 SHALL 保持数据库原有的 `zones` 列值不变。

#### Scenario: 批量更新传入非空可用区
- **WHEN** 调用方调用负载均衡批量更新接口并传入 `zones: ["ap-shanghai-2"]`
- **THEN** 数据库中对应负载均衡记录的 `zones` 字段 SHALL 被更新为 `["ap-shanghai-2"]`

#### Scenario: 批量更新未传入可用区时保持原值
- **WHEN** 调用方调用负载均衡批量更新接口（如仅更新业务归属或同步时间）且未传入 `zones`
- **THEN** 数据库中对应负载均衡记录的 `zones` 字段 SHALL 保持原值不变
