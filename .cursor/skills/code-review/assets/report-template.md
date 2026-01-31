# Code Review 报告

## 1. 基本信息

| 项目 | 内容 |
| :--- | :--- |
| **审查时间** | {{review_time}} |
| **需求关联** | #{{requirement_id}} {{requirement_title}} |
| **审查范围** | {{review_scope}} |
| **代码变动** | {{code_lines}}行 (+{{added_lines}}/-{{deleted_lines}}), {{file_count}}文件 |
| **审查结论** | **{{conclusion}}** (得分: {{overall_score}}/10) |

---

## 2. 需求与设计符合性

{{#if has_requirement}}
### 2.1 功能实现情况
{{#each requirement_features}}
- [{{#if implemented}}x{{else}} {{/if}}] **{{title}}**: {{#if implemented}}已实现{{else}}未实现 - {{description}}{{/if}}
{{/each}}

### 2.2 API 规范检查
| 接口 | 状态 | 说明 |
| :--- | :--- | :--- |
{{#each api_checks}}
| {{api_name}} | {{#if compliant}}✅{{else}}❌{{/if}} | {{#if compliant}}符合规格{{else}}期望: {{expected}}<br>实际: {{actual}}{{/if}} |
{{/each}}

### 2.3 边界条件
{{#each boundary_checks}}
- {{#if handled}}✅{{else}}⚠️{{/if}} **{{scenario}}**: {{#if handled}}已处理{{else}}未处理 (建议: {{suggestion}}){{/if}}
{{/each}}
{{else}}
*本次审查未关联具体需求文档，跳过需求符合性分析。*
{{/if}}

---

## 3. 代码质量与复杂度分析

### 3.1 复杂度概览
| 指标 | 数值 | 阈值 | 状态 |
| :--- | :--- | :--- | :--- |
| 平均圈复杂度 | {{avg_complexity}} | {{complexity_threshold}} | {{complexity_status}} |
| 最大圈复杂度 | {{max_complexity}} | {{complexity_threshold}} | {{max_complexity_status}} |
| 最大嵌套深度 | {{max_nesting_depth}} | {{nesting_threshold}} | {{nesting_status}} |

{{#if complex_functions}}
**⚠️ 需要重构的高复杂度函数**:
{{#each complex_functions}}
- `{{location}}` (CC: {{complexity}}, Lines: {{lines}})
{{/each}}
{{/if}}

### 3.2 可读性评分
- **命名规范**: {{naming_score}}/10 {{#if naming_comment}}({{naming_comment}}){{/if}}
- **注释完整性**: {{comment_score}}/10 {{#if comment_comment}}({{comment_comment}}){{/if}}
- **代码结构**: {{structure_score}}/10 {{#if structure_comment}}({{structure_comment}}){{/if}}

> **详细评价**: {{readability_details}}

---

## 4. 深度代码审查

### 4.1 严重问题 (Critical) 🛑
{{#if severe_issues}}
> 必须立即修复的问题，涉及逻辑错误、崩溃风险或严重规范违背。

{{#each severe_issues}}
#### {{index}}. {{title}}
- 📍 **位置**: `{{location}}`
- 📝 **描述**: {{description}}
- 📏 **规范**: {{rule_reference}}
{{#if code_snippet}}
- 💻 **代码**:
  ```{{language}}
  {{code_snippet}}
  ```
{{/if}}
- 💡 **建议**: {{suggestion}}
{{#if fixed_code}}
- 🔧 **修复示例**:
  ```{{language}}
  {{fixed_code}}
  ```
{{/if}}
{{/each}}
{{else}}
✅ 未发现严重问题。
{{/if}}

### 4.2 重要问题 (Major) ⚠️
{{#if important_issues}}
> 强烈建议修复的问题，涉及代码质量、潜在Bug或可维护性。

{{#each important_issues}}
#### {{index}}. {{title}}
- 📍 **位置**: `{{location}}`
- 📝 **描述**: {{description}}
- 💡 **建议**: {{suggestion}}
{{/each}}
{{else}}
✅ 未发现重要问题。
{{/if}}

### 4.3 优化建议 (Minor) 💡
{{#if suggestion_issues}}
> 改进代码风格、性能或可读性的建议。

{{#each suggestion_issues}}
- [ ] **{{title}}** (`{{location}}`): {{suggestion}}
{{/each}}
{{else}}
无优化建议。
{{/if}}

---

## 5. 专项评估

### 5.1 安全风险 (Security) 🛡️
{{#if security_issues}}
共发现 **{{security_issue_count}}** 处安全风险：
{{#each security_issues}}
- [ ] **{{risk_level}}**: {{title}} (`{{location}}`) - {{description}}
{{/each}}
{{else}}
✅ 未发现明显安全风险。
{{/if}}

### 5.2 性能考量 (Performance) 🚀
{{#if performance_issues}}
共发现 **{{performance_issue_count}}** 处性能问题：
{{#each performance_issues}}
- [ ] **{{title}}** (`{{location}}`): {{description}} (影响: {{impact}})
{{/each}}
{{else}}
✅ 未发现明显性能瓶颈。
{{/if}}

---

## 6. 审查总结与评分

### 6.1 维度评分
| 维度 | 得分 | 权重 |
| :--- | :--- | :--- |
| 需求符合性 | {{requirement_score}} | 30% |
| 编程规范 | {{standard_score}} | 25% |
| 可读性 | {{readability_score}} | 20% |
| 复杂度 | {{complexity_score}} | 15% |
| 安全性 | {{security_score}} | 10% |

**综合得分**: **{{overall_score}}** / 10

### 6.2 最终结论
{{conclusion}}

---

## 附录：审查元数据
- **工具版本**: CodeReview Skill v{{version}}
- **规则集**: {{enabled_rules}}
- **语言统计**:
{{#each language_stats}}
  - {{language}}: {{file_count}} files, {{lines}} lines
{{/each}}
- **生成时间**: {{generation_time}}

---

### A. 相关文档
{{RELATED_DOCS}}

### B. 参考资料
{{REFERENCES}}