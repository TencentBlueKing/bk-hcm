# C++ 编码规范摘要

## 1. 基本原则
- **少即是多 (Less Code)**：移除冗余代码（如不必要的 `private:`，默认构造函数）。
- **最少知识原则 (Least Knowledge)**：最小化函数参数类型；只传递需要的数据。
- **确定性 (Determinism)**：使用 `override`，显式 `const` 引用和具体类型以减少歧义。
- **一致性 (Consistency)**：保持逻辑实现和命名约定的一致性。
- **无隐式规则 (No Implicit Rules)**：显式声明依赖；记录隐藏逻辑。
- **直接表达 (Direct Expression)**：使用直接的 API（如 `contains` vs `count`，`ptr == nullptr` vs `!ptr`）。

## 2. 代码风格
- **提交日志 (Commit Logs)**：清晰的摘要，原子提交（每个提交一个功能）。
- **注释 (Comments)**：统一风格（Doxygen）；完整、准确、简洁。解释“为什么”，而不是“是什么”。
- **命名 (Naming)**：简洁、准确、完整。避免令人困惑的名称。
- **依赖 (Dependencies)**：使用 Bazel Tags 或 Git Commit IDs 锁定版本。

## 3. 类与接口
- **封装 (Encapsulation)**：隐藏内部细节；提供最小接口。
- **单一职责 (Single Responsibility)**：将数据持有者与逻辑执行者分离。
- **内聚性 (Cohesion)**：设计插件/模块使其自包含（例如，自注册）。

## 4. 表达式与语句
- **循环 (Loops)**：优先使用 `for (const auto& [k, v] : map)` 进行结构化绑定。
- **回调 (Callbacks)**：优先使用 Lambda 而非 `std::bind`。
- **Auto**：仅在提高可读性时使用（例如迭代器）。

## 5. 性能优化
- **移动语义 (Move Semantics)**：正确使用 `std::move`（不要用于 `const` 或返回值）。
- **原位构造 (In-Place Construction)**：使用 `emplace`, `try_emplace`, `emplace_back`。
- **String View**：对于常量字符串参数使用 `std::string_view` 以避免分配。
- **迭代器 (Iterators)**：重用迭代器；避免重复查找（先 `find` 后 `[]`）。
- **计算 (Computation)**：推迟不必要的计算；预计算不变量。
- **隐藏拷贝 (Hidden Copies)**：注意循环或参数中的隐式拷贝；使用 `const&`。
- **容器 (Containers)**：除非需要排序，否则默认使用 `unordered_map`。

## 6. 并发编程
- **线程 (Threads)**：优先使用 `std::thread`/`std::async`（或 C++20 中的 `std::jthread`）。
- **同步 (Synchronization)**：使用 `std::lock_guard`/`std::unique_lock`。避免直接调用 mutex。
- **原子操作 (Atomics)**：对简单的共享数据使用 `std::atomic`。
- **条件变量 (Condition Variables)**：始终使用 `while` 循环检查条件（防止虚假唤醒）。
- **线程局部存储 (Thread Local)**：使用 `thread_local` 避免全局状态争用。
- **生命周期 (Lifecycle)**：确保进程退出前停止线程；警惕 Lambda 按引用捕获局部变量。

## 7. 资源管理 (RAII)
- **RAII**：通过对象生命周期管理资源。
- **Unique Ptr**：默认选择 (`std::unique_ptr`, `std::make_unique`)。
- **Shared Ptr**：仅用于共享所有权 (`std::shared_ptr`, `std::make_shared`)。
- **Weak Ptr**：打破引用循环。
- **参数传递**：
  - `unique_ptr`：转移所有权。
  - `shared_ptr`：共享所有权。
  - `const shared_ptr&`：可选共享（避免引用计数原子递增）。

## 8. 单元测试
- **覆盖率 (Coverage)**：测试边界条件和异常。
- **隔离 (Isolation)**：对外部依赖使用 Mock。
- **原子性 (Atomicity)**：每个功能/场景一个测试用例。
- **数据 (Data)**：使用测试夹具 (Test Fixtures)；避免硬编码。