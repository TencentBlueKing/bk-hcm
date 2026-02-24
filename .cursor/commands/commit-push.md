---
description: 智能提交代码并推送到个人远程仓库
---

智能提交并推送代码。参数: `$ARGUMENTS`

## 执行步骤

### 1. 检查工作区状态

首先检查 git 状态，区分已跟踪和未跟踪的文件：

- 已跟踪的已修改文件：默认会被提交
- 未跟踪的文件：列出并询问用户是否添加

### 2. 处理未跟踪文件

如果有未跟踪的文件，列出它们并询问用户：

- 选择要添加的文件（可多选）
- 或跳过继续提交

### 3. 生成 Commit Message

分析修改的文件，自动生成符合规范的 commit message：

- 格式: `<type>: <description>\n\n--story=<tapid>` 或 `--bug=<bugid>`
- type 可选: feat / fix / docs / refactor / chore / style
- 展示给用户确认或修改

**TAPD ID 解析规则**：如果对话中提供了 TAPD 链接，从链接中提取正确的需求/缺陷 ID：
- TAPD 链接格式: `.../tapd_fe/{project_id}/story/detail/{full_id}` 或 `.../bug/detail/{full_id}`
- 完整 ID 结构（19位）: `1` + `0` + `{project_id}` (8位) + `{real_id}` (9位)
- 提取方法: 从完整 ID 中去掉前10位（`10` + 项目ID），剩余9位即为真正的 TAPD ID
- 示例: 项目ID `69995598`，完整ID `1069995598130013520` → 去掉前10位 `1069995598` → 真正ID `130013520`

### 4. 执行提交

使用 `git commit -a -m "<message>"` 提交已跟踪的修改文件。

### 5. 推送到远程

**自动识别个人远程仓库**：通过名称模式匹配（me / local / personal / mine）自动找到个人远程。

- 如果找到匹配的远程，自动使用
- 如果找到多个或未找到，列出所有远程让用户选择

使用 `git push -u <remote> <branch>` 推送并关联远程分支。

### 6. 打开 PR 链接

推送成功后，从 git push 输出中解析 PR 创建链接，自动在浏览器中打开。
