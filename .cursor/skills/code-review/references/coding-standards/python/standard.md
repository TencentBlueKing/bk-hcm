# Python 编码规范精华版

## 核心原则

### Pythonic 风格
- 遵循 PEP 8 和 PEP 20（The Zen of Python）
- 代码为人而写，优先可读性和可维护性
- 每件事应有一种直白明显的方法完成

### 命名规范
- 模块/变量/函数/方法/参数：`snake_case`
- 类/异常：`PascalCase`
- 常量：`UPPER_CASE`
- 避免保留字、内置函数名和单字母（除约定俗成场景）

## 代码质量

### 注释
- 简洁明了,解释"为什么"而非"怎么样"
- 避免显而易见的注释
- 使用 docstring 描述函数/类/模块
- 用 TODO 标记待完成任务（含工单链接和负责人）
- 不用注释删除代码，使用版本控制

### 安全
- **严禁硬编码敏感信息**（IP/域名/账号/密码）
- 使用配置文件（JSON/YAML/Pydantic）管理配置
- 避免 `eval`/`exec` 执行字符串代码
- 使用 `json` 模块代替 `eval` 解析字符串

## 语法与特性

### 基础规范
- 缩进：4个空格（禁止空格与制表符混用）
- 行宽：最大120字符
- 字符串：优先双引号
- 用括号连接长代码行，不用反斜杠

### 真值测试
```python
# Good
if my_list:  # 判断列表非空
if my_bool:  # 判断布尔值
if x is not None:  # 内联否定

# Bad
if len(my_list) != 0:
if my_bool != False:
if not x is None:
```

### 导入
- 优先绝对导入
- 避免循环依赖（延迟导入/重构模块）
- 导入顺序：标准库 → 第三方库 → 本地模块

### 字符串
- 使用 f-string 格式化（Python 3.6+）
- 避免循环中字符串拼接，用 `join()` 或列表
```python
# Good
result = "".join([str(i) for i in range(100)])
# Bad
result = ""
for i in range(100):
    result += str(i)
```

### 类型注解
- 使用 typing 模块提供类型提示
- 循环引用用 `TYPE_CHECKING` 或字符串注解

## 惯用法

### 资源管理
```python
# 使用 with 语句
with open('file.txt', 'r') as f:
    content = f.read()
```

### 迭代优化
```python
# 使用 enumerate 代替手动索引
for i, item in enumerate(items):
    print(f"{i}: {item}")

# 使用 zip 并行迭代
for key, value in zip(keys, values):
    print(f"{key}: {value}")
```

### 数据结构
```python
# 使用 defaultdict 避免键检查
from collections import defaultdict
counts = defaultdict(int)
counts[key] += 1

# 列表推导（简单场景）
squares = [x**2 for x in range(10)]

# 生成器表达式（大数据）
squares_gen = (x**2 for x in range(1000000))
```

### 表驱动法
```python
# Good - 用字典代替多个 if-elif
permissions = {
    "admin": ["read", "write", "delete"],
    "user": ["read", "write"],
    "guest": ["read"]
}
perms = permissions.get(role, [])

# Bad - 多个 if-elif
if role == "admin":
    perms = ["read", "write", "delete"]
elif role == "user":
    perms = ["read", "write"]
```

## 性能优化

### 常见优化
- 用集合（set）代替列表进行成员检测：`O(1)` vs `O(n)`
- 用列表推导代替循环添加元素
- 大数据用生成器代替列表
- 减少不必要的属性访问（缓存到局部变量）

### 避免
- 循环中字符串拼接
- 过度使用 `set` 进行数据操作（机器学习场景）
- 浮点数直接比较（用 `math.isclose()`）

## 面向对象

### 属性访问
- 使用 `@property` 装饰器代替 getter/setter
- 私有属性用单下划线 `_var`（约定）或双下划线 `__var`（名称改写）

### 装饰器
- 适用场景：日志/权限检查/缓存/性能监控
- 避免过度使用和多层嵌套
- 保持简单明确，添加清晰文档

### 数据类
```python
from dataclasses import dataclass

@dataclass
class User:
    name: str
    age: int
    email: str = ""
```

## 异常处理

### 最佳实践
```python
# Good - 捕获具体异常
try:
    result = risky_operation()
except ValueError as e:
    logger.error(f"Invalid value: {e}")
except KeyError as e:
    logger.error(f"Missing key: {e}")

# Bad - 捕获所有异常
try:
    result = risky_operation()
except:
    pass
```

### 日志
- 使用 logging 模块，不用 print
- 设置合适的日志级别（DEBUG/INFO/WARNING/ERROR/CRITICAL）

## 并发编程

### 选择模型
- **IO密集型**：协程（asyncio）或多线程
- **CPU密集型**：多进程
- 注意 GIL（全局解释器锁）限制

### 协程
```python
import asyncio

async def fetch_data(url):
    # 异步操作
    return data

# 并发调用
results = await asyncio.gather(
    fetch_data(url1),
    fetch_data(url2)
)
```

## 测试

### 测试策略
- 编写单元测试和集成测试
- 使用表格驱动测试（参数化）
- Mock 外部依赖
- 追求合理的覆盖率（不盲目追求100%）

### 工具
- 测试框架：pytest/unittest
- 代码检查：Pylint/Flake8
- 格式化：Black
- 类型检查：mypy

## 避免的反模式

### 黑魔法
- 慎用动态修改类和对象
- 避免猴子补丁（Monkey Patching）
- 谨慎使用反射和元编程
- 不过度使用装饰器

### 其他
- 避免全局变量（用配置模块集中管理）
- 不嵌套列表推导（影响可读性）
- 不嵌套三元表达式
- 避免深层嵌套（用提前返回或逻辑合并）

## 工具链

### 环境管理
- 使用虚拟环境（venv/conda/Miniforge）
- 软件源：优先公司内部镜像

### 代码质量
- Linter：Pylint（推荐）
- Formatter：Black（推荐）
- CI集成：CodeCC 腾讯代码分析

### 版本选择
- 优先 Python 3.9+（3.9之前已EOL）
- 避免 Python 2（已停止维护）

## 快速检查清单

- [ ] 遵循命名规范（snake_case/PascalCase）
- [ ] 无硬编码敏感信息
- [ ] 使用 with 管理资源
- [ ] 优先标准库方法
- [ ] 捕获具体异常类型
- [ ] 使用 f-string 格式化
- [ ] 添加类型注解
- [ ] 编写 docstring
- [ ] 使用 Linter 和 Formatter
- [ ] 编写测试用例