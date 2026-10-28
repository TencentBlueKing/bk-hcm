## ADDED Requirements

### Requirement: 目录约定与目录检查
系统 SHALL 按 `{插件根目录}/{插件名}/{版本}/{架构}/` 约定扫描插件，插件名、版本、架构 MUST 取自目录路径。插件根目录 MUST 可配置并有默认值（`/data/hcm/plugins`）。同一插件允许多个版本目录，副本 MUST 使用按 semver 比较的最高版本，低版本目录只记日志、不注册。每个副本 MUST 只读取自己架构（`{os}_{arch}`）的目录。每个架构目录下 MUST 恰有一个可执行文件，不要求文件名符合固定模板。插件名不对字符做额外限制。

#### Scenario: 多版本取最高
- **WHEN** `tencentcloud` 下有 `1.83.30` 与 `1.83.34` 两个版本目录
- **THEN** 副本使用 `1.83.34`，`1.83.30` 只记日志

#### Scenario: semver 而非字典序
- **WHEN** 版本目录为 `1.9.0` 与 `1.10.0`
- **THEN** 选用 `1.10.0`

#### Scenario: 混部架构
- **WHEN** 版本目录下有 `linux_amd64` 与 `linux_arm64`，副本运行在 amd64
- **THEN** 只读取 `linux_amd64` 目录

#### Scenario: 版本目录名非法
- **WHEN** 版本目录名不是合法版本号（如 `latest`）
- **THEN** 目录检查失败

#### Scenario: 可执行文件不唯一或缺失
- **WHEN** 架构目录下没有文件或有多个文件
- **THEN** 目录检查失败

#### Scenario: 文件名不做限制
- **WHEN** 可执行文件名为任意名称
- **THEN** 只要目录下仅有这一个文件即通过

### Requirement: 目录检查结果作为副本只读信息
目录检查 SHALL 为每个插件产出一条条目：插件名、版本、架构目录、可执行文件路径、本架构 sha256。条目 MUST 在启动时确定，之后只读；运行中替换目录 MUST NOT 生效，只有重启才生效。文件位置 MUST NOT 写入注册表。

#### Scenario: 运行中替换目录
- **WHEN** 副本运行期间插件目录内容被替换
- **THEN** 副本内存中的条目不变，不重复校验

### Requirement: sha256 校验（只算本架构）
首次注册某版本时，持锁副本 SHALL 只计算**本架构**可执行文件的 sha256 并写入记录 `extension.architectures`（键为 `{os}_{arch}`），MUST NOT 遍历其它架构目录。副本启动时若目录版本与注册表已有记录版本相同，MUST 比对本架构 sha256：一致则通过；不一致则退出；注册表里没有本架构的值则写入。同一版本号下本架构文件内容变化 MUST 被拒绝。

#### Scenario: 首次注册只写本架构 hash
- **WHEN** 首次注册 `1.83.34`，目录下有 amd64 与 arm64 两个架构，副本运行在 amd64
- **THEN** `extension.architectures` 只包含 amd64 的 sha256

#### Scenario: 另一架构首次启动时补齐
- **WHEN** 记录里只有 amd64 的 sha256，某个 arm64 副本启动并处理同一版本
- **THEN** 写入 arm64 的 sha256，已有的 amd64 值不变

#### Scenario: 一致
- **WHEN** 启动时本架构 sha256 与记录一致
- **THEN** 通过，不拉起进程

#### Scenario: 不一致
- **WHEN** 同版本下本架构 sha256 与记录不同
- **THEN** 进程退出，不修改任何记录

#### Scenario: 记录里缺本架构
- **WHEN** 记录里没有本架构的 sha256
- **THEN** 写入本架构 sha256，通过
