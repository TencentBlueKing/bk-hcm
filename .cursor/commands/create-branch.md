---
description: 基于指定分支创建新的开发分支
---

## 使用说明

创建新的开发分支，支持两种调用方式：

### 方式一：带参数调用

```
/bk-hcm/create-branch <新分支名> <目标分支名>
```

示例：
```
/bk-hcm/create-branch feat-my-feature origin/v1.8.x
```

### 方式二：交互式调用

```
/bk-hcm/create-branch
```

不提供参数时，会通过对话交互收集信息。

---

**重要**：根据此命令文件所在的 `.cursor/commands/` 目录，自动识别项目根目录并在该目录下执行 git 命令。

## 执行流程

### 1. 识别项目目录

从命令文件路径中提取项目根目录（去掉 `.cursor/commands/create-branch.md` 部分）。

### 2. 解析参数或收集信息

检查 `$ARGUMENTS` 是否提供了参数：

- **有参数**：解析参数，格式为 `<新分支名> <目标分支名>`
- **无参数**：使用交互式对话收集以下信息：
  - **新分支名称**：要创建的新分支名
  - **基础分支**（选择）：要基于的远程分支名，提供常用选项：
    - `origin/v1.8.x`（1.8.x 版本）
    - `origin/dev`（开发分支）
    - `origin/main`（主分支）
    - 其他（让用户输入自定义分支名）
  - **是否启动开发服务器**（选择）：创建完成后是否启动 `npm run dev`

### 3. 执行步骤

在识别的项目目录下执行：

1. Fetch 更新远程分支信息：

   ```bash
   git fetch origin
   ```

2. 基于目标分支创建新分支：

   ```bash
   git checkout <基础分支> && git checkout -b <新分支名>
   ```

3. 确认分支创建成功：

   ```bash
   git branch --show-current
   ```

4. 可选 - 启动开发服务器（在 front 子目录）：
   ```bash
   cd front && npm run dev
   ```

## 示例

### 带参数调用

```
/bk-hcm/create-branch feat-ticket-new-feature origin/v1.8.x
```

### 交互式调用

```
/bk-hcm/create-branch

Q: 请输入新分支名称
A: feat-my-new-feature

Q: 请选择基础分支
A: origin/v1.8.x

Q: 是否启动开发服务器？
A: 是
```
