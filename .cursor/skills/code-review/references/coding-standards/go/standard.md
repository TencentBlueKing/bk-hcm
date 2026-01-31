# Go 编码规范精要

## 基本原则
- **逻辑清晰**: 准确命名、有用注释、合理代码组织
- **简单**: 最简代码实现目标，优先使用标准库
- **简洁**: 避免重复代码、多余代码、晦涩命名、过度抽象
- **可维护**: 结构化API、完备测试、避免代码耦合
- **一致性**: 包内风格统一，冲突时以规范为准

## 命名规范

### 包名
- 简短、小写、单数，无下划线/驼峰
- 避免 `common`、`util` 等通用名
- 可用缩写: `fmt`、`strconv`、`io`
- 包名与目录名一致

### 接口名
- 单方法接口加 `-er` 后缀: `Reader`、`Writer`
- 避免使用 `Read`、`Write`、`Close` 等规范名称除非语义相同

### 接收器名
- 1-2字母，类型缩写，保持一致
```go
// Good
func(c *Calculator) Add(a,b int) int {}
func(c *Calculator) Sub(a,b int) int {}
```

### 变量名
- 长度与作用域成正比
- 小作用域(1-7行): 单字母可接受
- 省略类型: `users` 优于 `userSlice`
- 常用单字母: `r`(Reader)、`x,y`(坐标)

### 减少重复
- 包名+标识符: `db.New()` 优于 `db.NewDB()`
- 方法名不重复接收器/参数/返回值信息

## 错误处理

### 基本规则
```go
// 错误检查后立即返回
if err != nil {
    return err
}
// 正常代码

// 错误信息小写开头，无标点
err := fmt.Errorf("something bad happened")
```

### 错误结构化
```go
// 使用哨兵值
var ErrNotFound = errors.New("not found")

// 调用方判断
if errors.Is(err, ErrNotFound) { ... }
```

### 错误包装
```go
// %w 放末尾
return fmt.Errorf("user update failed: %w", err)

// 不要暴露实现细节除非有意为之
```

### 错误信息要求
- 说明出了什么问题
- 指出无效输入
- 解释如何修复
- 提供示例

### Panic 使用
- 不用于正常错误处理
- 仅用于不可恢复的内部状态错误
- 初始化失败可用 `log.Fatal`

## 并发

### Context
- 作为函数第一个参数
- 不放入结构体
- 不创建自定义 Context 类型

### Goroutine
- 明确生命周期和退出条件
- 使用 `context.Context` 管理取消
- 同步函数优于异步函数

### sync.Mutex
```go
func (c *Counter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}
```

## 类型使用

### 切片/映射
```go
// 空切片用 nil
var t []string  // Good
t := []string{} // Bad

// 预分配容量
m := make(map[string]bool, size)
s := make([]T, 0, cap)
```

### 接口
- 定义在使用侧，非实现侧
- 实现方返回具体类型
- 编译期验证: `var _ Interface = (*impl)(nil)`
- 不使用指向接口的指针

### 接收器选择
- 需修改接收器: 指针
- 包含不可复制字段(sync.Mutex): 指针
- 大结构体: 指针
- 小值类型/内置类型: 值
- map/func/chan: 值

## 代码格式

### 导入
```go
import (
    "标准库"

    "第三方库"

    "本地包"
)
```
- 避免点导入 `import .`
- 避免空白导入 `import _`（main/测试除外）

### 条件循环
- 不换行 if 条件
- 提取复杂条件为变量
- 变量放等号左侧: `if result == "foo"`

### 字符串拼接
- 简单: `+`
- 格式化: `fmt.Sprintf`
- 逐步构建: `strings.Builder`
- 分隔符: `strings.Join`
- 文件路径: `filepath.Join`

## 测试

### 基本要求
- 失败信息包含: 原因、输入、实际值、期望值
- 格式: `YourFunc(%v) = %v, want %v`
- got 在 want 之前

### 表驱动测试
```go
tests := []struct {
    name  string
    input string
    want  string
}{
    {"case1", "in1", "out1"},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got := Func(tt.input)
        if got != tt.want {
            t.Errorf("Func(%q) = %q, want %q", tt.input, got, tt.want)
        }
    })
}
```

### 错误处理
- `t.Error`: 继续执行，打印所有失败
- `t.Fatal`: 前置条件失败，无法继续
- 不在独立 goroutine 中调用 `t.Fatal`

### 测试辅助
- 调用 `t.Helper()` 定位失败行
- 辅助函数在 context 后、测试逻辑前
- 使用 `t.Cleanup` 注册清理函数

## 常见错误

### 循环变量（go<1.22）
```go
// Bad: 闭包捕获循环变量
for _, v := range items {
    go func() { use(v) }()  // v 被覆盖
}

// Good: 复制或传参
for _, v := range items {
    v := v
    go func() { use(v) }()
}
```

### 中文字符串处理
```go
// Bad: 按字节遍历
for i := 0; i < len(s); i++ { ... }

// Good: 转 rune
for i, r := range []rune(s) { ... }
```

### 随机数
- 初始化种子仅一次（go≥1.22自动初始化）
- 用 `rand.Intn(n)` 非 `rand.Int() % n`
- 安全场景用 `crypto/rand`

### 拷贝
- 不复制 `sync.Mutex`
- 不复制 `bytes.Buffer`
- 需要拷贝时实现 `DeepCopy`

## 性能建议

### 预分配
```go
m := make(map[K]V, expectedSize)
s := make([]T, 0, expectedCap)
```

### 字符串构建
- 大量拼接用 `strings.Builder`
- 输出到 io.Writer 用 `fmt.Fprintf`

### 值 vs 指针传递
- 小值类型直接传值
- 大结构体传指针
- 实际性能以 benchmark 为准