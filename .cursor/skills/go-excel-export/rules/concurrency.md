# 并发安全与数据预加载

## 并发安全

Stream 模式下，Handler 中的共享状态需要使用锁机制保护。

### 共享状态字段

```go
type StreamHandler struct {
	// ...
	sheetRowMap   map[string]int  // 共享状态：各 Sheet 的当前行号
	sheetRowMutex sync.Mutex      // 保护 sheetRowMap
}
```

### 锁使用模式

```go
func (e *Handler) AddStreamRows(...) error {
	// 读取共享状态前加锁
	e.sheetRowMutex.Lock()
	currentRow, exists := e.sheetRowMap[sheetName]
	e.sheetRowMutex.Unlock()

	// ... 处理数据 ...

	// 更新共享状态前加锁
	e.sheetRowMutex.Lock()
	e.sheetRowMap[sheetName] = currentRow
	e.sheetRowMutex.Unlock()
}
```

## 数据预加载

使用 `sync.Once` 实现关联数据的一次性预加载，避免重复查询。

### Handler 结构

```go
type OrgObsBillStreamExporterHandler struct {
	// ...
	opProductMap map[string]string  // 关联数据缓存
	mapOnce      sync.Once          // 确保只加载一次
}
```

### 预加载方法

```go
func (e *OrgObsBillStreamExporterHandler) loadOpProductMap(kt *kit.Kit) error {
	var loadErr error
	e.mapOnce.Do(func() {
		// 这个函数只会执行一次
		resp, err := e.cs.DataService().GetOpProducts(kt)
		if err != nil {
			loadErr = err
			return
		}
		e.opProductMap = make(map[string]string)
		for _, item := range resp.Items {
			e.opProductMap[item.ID] = item.Name
		}
	})
	return loadErr
}
```

### 在 AddStreamRows 中调用

```go
func (e *Handler) AddStreamRows(...) error {
	// 预加载关联数据（只会执行一次）
	if err := e.loadOpProductMap(kt); err != nil {
		return err
	}

	for _, item := range items {
		// 使用预加载的数据
		opProductName := e.opProductMap[item.OpProductID]
		// ...
	}
}
```

## 分页查询模式

大数据量导出时，使用分页查询避免一次性加载全部数据。

### Handler 字段

```go
type StreamHandler struct {
	currentOffset int  // 当前偏移量
	batchSize     int  // 每批次大小
}
```

### GetItemsStream 实现

```go
func (e *Handler) GetItemsStream(kt *kit.Kit) ([]Item, bool, error) {
	req := &webset.GetDataReq{
		Page: &types.BasePage{
			Start: uint32(e.currentOffset),
			Limit: uint32(e.batchSize),
		},
	}

	resp, err := e.cs.WebServer().GetData(kt, req)
	if err != nil {
		return nil, false, err
	}

	// 更新偏移量
	e.currentOffset += len(resp.Items)

	// 判断是否还有更多数据
	hasMore := len(resp.Items) == e.batchSize

	return resp.Items, hasMore, nil
}
```

## 错误示例

```go
// ❌ 未使用锁保护共享状态
func (e *Handler) AddStreamRows(...) error {
	currentRow := e.sheetRowMap[sheetName]  // 直接访问，线程不安全
	// ...
	e.sheetRowMap[sheetName] = currentRow   // 直接修改，线程不安全
}

// ❌ 每次调用都重新查询关联数据
func (e *Handler) AddStreamRows(...) error {
	// 每次都会查询，效率低下
	opProductMap, _ := e.cs.DataService().GetOpProducts(kt)
	// ...
}

// ❌ 忘记更新偏移量
func (e *Handler) GetItemsStream(kt *kit.Kit) ([]Item, bool, error) {
	resp, _ := e.cs.WebServer().GetData(kt, req)
	// 忘记 e.currentOffset += len(resp.Items)
	return resp.Items, true, nil  // hasMore 判断也会出错
}
```
