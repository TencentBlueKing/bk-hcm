# Java 编码规范精要

> 基于腾讯 Java 语言编程指南浓缩整理

## 一、异常处理

### 核心原则
- **API设计者**:无必要不抛异常
- **API使用者**:永远不要忽略受检异常
- 可恢复用受检异常,默认用非受检异常
- 异常不是控制流工具

### 关键规则
1. **精确捕获**:逐个列出异常类型,避免泛泛捕获`Exception`
2. **必要清理**:使用`finally`块或`try-with-resources`释放资源
3. **优先标准异常**:`IllegalArgumentException`(参数错误)、`IllegalStateException`(状态错误)、`NullPointerException`(空值)、`IndexOutOfBoundsException`(越界)、`UnsupportedOperationException`(未实现)
4. **文档化**:用`@throws`标签记录抛出条件
5. **异常链**:保留原始异常,添加业务上下文
6. **避免循环抛异常**:预先验证,避免高频异常影响性能

### 禁止事项
- ❌ 不要catch后忽略(必须注释说明且变量名为`ignored`)
- ❌ 不要直接重用`Exception`/`RuntimeException`/`Throwable`/`Error`
- ❌ 不要在finally块中抛出异常

## 二、并发编程

### 核心原则
- 健壮性优先于高性能
- 读写都要同步
- 优先使用不可变性
- 优先使用虚线程(JDK21+/Kona JDK8+)处理I/O密集型
- 优先使用`java.util.concurrent`工具库

### 关键规则
1. **同步共享数据**:读写都需同步,避免使用`volatile`(非原子)
2. **使用原子类**:`AtomicLong`/`AtomicInteger`等替代同步方法
3. **优先Executors**:使用`ExecutorService`而非直接创建`Thread`
4. **虚线程场景**:I/O密集型、高并发网络服务;不适合CPU密集型
5. **并发集合**:`ConcurrentHashMap`、`CopyOnWriteArrayList`、`Collections.synchronizedList()`
6. **复合操作加锁**:迭代器、条件检查等复合操作需自行加锁
7. **线程安全文档化**:明确说明类的线程安全级别(不可变/无条件安全/有条件安全/非线程安全/线程对立)

## 三、集合使用

### 核心原则
- 优先返回空集合而非null
- 性能敏感场景考虑初始容量
- 考虑并发场景选择合适实现

### 关键规则
1. **空安全**:返回`Collections.emptyList()`或`List.of()`而非null
2. **初始容量**:预知大小时指定初始容量避免扩容
3. **不可变集合**:对外暴露用`Collections.unmodifiableList()`或`List.of()`
4. **并发安全**:
   - 使用`CopyOnWriteArrayList`(读多写少)
   - 使用`Collections.synchronizedList()`
   - 复合操作自行加锁
5. **防御性拷贝**:构造和返回时拷贝可变集合

## 四、Null处理

### 核心原则
- 函数返回空值用`Optional<T>`
- 类成员/入参用`@Nullable`标明可空
- 默认所有域非空

### 关键规则
1. **Optional使用**:
   - 返回值用`Optional<T>`表示可能为空
   - Optional自身永远非null
   - 不要用`@Nullable Optional<T>`
2. **容器非空**:所有容器(List/Set/Map)应非null,区分空容器和null容器无意义
3. **注解标记**:可空用`@Nullable`,项目内统一注解库
4. **静态检查**:推荐使用Checker Framework或NullAway

## 五、Stream流

### 核心原则
- 避免`parallelStream`(小数据集反而慢,并发风险高)
- 不返回`Stream<T>`作为API结果
- Stream操作无副作用

### 关键规则
1. **顺序流优先**:除非大量CPU密集型纯计算且经过性能测试
2. **返回集合**:API返回`List`/`Set`而非`Stream`
3. **无副作用**:不修改外部状态,不依赖可变状态,保持幂等性
4. **正确终止**:用`collect()`等终止操作获取结果

## 六、枚举

### 核心原则
- 优先使用enum替代int常量
- 永远不用`ordinal()`方法
- 谨慎使用`default`分支
- 用`EnumSet`替代位字段
- 用`EnumMap`替代序数索引

### 关键规则
1. **实例字段**:用字段存储关联数据,不依赖`ordinal()`
2. **类型安全**:编译时检查,避免整型常量的类型混淆
3. **switch处理**:可控范围省略`default`获得编译检查
4. **EnumSet**:替代位字段,类型安全且高效
5. **EnumMap**:专为枚举键设计,性能优于HashMap
6. **添加行为**:可添加字段、方法、抽象方法实现

## 七、不可变性

### 核心原则
- 不可变性是优先推荐的编程方式
- 不可变对象线程安全、易理解、易测试
- 合理使用防御性拷贝

### 关键规则
1. **设计原则**:
   - 不提供变异子方法(setter)
   - 类不可继承(final或私有构造)
   - 所有字段`private final`
   - 可变字段防御性拷贝
2. **防御性拷贝**:
   - 构造时拷贝可变参数
   - 返回时拷贝可变字段
   - 或返回不可修改视图
3. **Builder模式**:字段多时用Builder创建不可变对象
4. **性能考虑**:现代JVM的GC足够高效,不必过度担心对象创建开销

## 八、类结构

### 标准顺序
1. 类文档注释
2. 类声明
3. 静态变量(public→protected→package→private)
4. 实例变量(同上顺序)
5. 构造函数
6. 方法(public→protected→package→private)

### 方法组织
- **就近原则**(推荐):私有方法紧邻调用它的公共方法
- **公开-私有原则**:所有公共方法前置,私有方法后置
- **重载方法**:必须相邻放置,变参方法放最后
- **内部类/枚举**:放在类末尾

### 访问控制
- 最小化访问权限
- 优先private,必要时才扩大
- 包级访问避免污染全局

## 九、命名规范

### 类命名
- **接口**:不加`I`前缀或`Interface`后缀
- **抽象类**:加`Abstract`或`Base`前缀
- **实现类**:接口名+`Impl`(单一实现)或具体特征名(多实现)
- **异常**:`Exception`后缀(受检)、`Error`后缀(错误)
- **测试**:`Test`后缀

### 方法命名
- **创建**:`of`/`from`/`create`/`valueOf`/`getInstance`
- **查找**:`find`/`search`/`query`(统一使用)
- **添加**:`add`/`append`/`insert`/`put`
- **删除**:`remove`/`delete`/`clear`
- **转换**:`to`/`convert`/`transform`
- **判断**:`is`/`has`/`can`

### 变量命名
- **常量**:全大写下划线分隔
- **集合**:复数形式或`xxxList`/`xxxSet`
- **布尔**:`is`/`has`/`can`开头
- **避免无意义词**:不用`Data`/`Info`/`Object`等后缀(除非必要)

## 十、其他最佳实践

### 资源管理
- 优先`try-with-resources`
- 手动管理需在finally中关闭
- 关闭异常不应覆盖主异常

### 性能
- 避免过早优化
- 字符串拼接用`StringBuilder`(循环中)
- 集合指定初始容量
- 缓存常用不可变对象

### 代码质量
- 使用静态分析工具
- 编写单元测试
- 代码审查关注异常处理、资源释放、线程安全
- 保持团队编码风格一致

### 设计模式
- **Builder**:复杂对象构建
- **Factory**:对象创建逻辑复杂时
- **Singleton**:用枚举实现(最佳方式)
- **Strategy**:用枚举+抽象方法实现策略

## 十一、禁止事项清单

❌ 不要使用`ordinal()`方法
❌ 不要忽略受检异常
❌ 不要在循环中抛异常
❌ 不要返回null集合
❌ 不要使用`parallelStream`(除非必要)
❌ 不要返回`Stream`作为API结果
❌ 不要在Stream中修改外部状态
❌ 不要使用位字段(用EnumSet)
❌ 不要使用序数索引(用EnumMap)
❌ 不要在枚举中使用可变字段
❌ 不要暴露可变内部状态
❌ 不要过度使用受检异常

## 参考资源

- 《Effective Java》第三版 - Joshua Bloch
- 《Java并发编程实战》- Brian Goetz
- 腾讯Java语言编程指南
- Oracle Java官方文档