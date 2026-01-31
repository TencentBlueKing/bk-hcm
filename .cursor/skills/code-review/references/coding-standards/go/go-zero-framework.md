# go-zero 框架专项审查指南

## 概述

本文档提供针对 go-zero 框架的专项代码审查指南。仅当检测到项目使用 go-zero 框架时才应用这些审查规则。

## 框架检测

在开始审查前，检测项目是否使用 go-zero 框架：

1. 检查项目根目录或子目录是否存在 `.api` 文件（go-zero 的 API 定义文件）
2. 检查 `go.mod` 文件中是否包含 `github.com/zeromicro/go-zero` 依赖
3. 如果检测到 go-zero 框架，在审查报告中说明并启用本专项检查

## 审查类别

### 1. API 定义审查（*.api 文件）

**规则标识**: `[category:go-zero-api]`

#### 1.1 API 设计规范
**规则**: `[rule:go-zero-api-design]`

**检查要点**:
- 路由定义是否清晰、符合 RESTful 规范
- 路由分组是否合理（按业务模块或功能分组）
- HTTP 方法使用是否正确（GET/POST/PUT/DELETE/PATCH）
- 路径参数命名是否规范

**示例问题**:
```api
// ❌ 不好的设计
post /getUserInfo (GetUserRequest) returns (GetUserResponse)

// ✅ 好的设计
get /api/v1/users/:id (GetUserRequest) returns (GetUserResponse)
```

#### 1.2 类型定义
**规则**: `[rule:go-zero-type-definition]`

**检查要点**:
- Request/Response 结构体定义是否完整
- 字段标签（json、form、path、validate）是否正确
- 是否有必要的验证标签（required、min、max等）
- 字段命名是否符合规范（驼峰命名）

**示例问题**:
```api
// ❌ 缺少验证标签
type CreateUserRequest {
    Username string `json:"username"`
    Password string `json:"password"`
}

// ✅ 包含验证标签
type CreateUserRequest {
    Username string `json:"username" validate:"required,min=3,max=20"`
    Password string `json:"password" validate:"required,min=8"`
}
```

#### 1.3 JWT 和认证配置
**规则**: `[rule:go-zero-auth]`

**检查要点**:
- JWT 配置是否正确（Auth 字段）
- 需要认证的接口是否都配置了 jwt
- 认证配置是否一致

**示例问题**:
```api
// ❌ 需要认证但未配置
@server(
    group: user
)
service user-api {
    @handler getUserProfile
    get /api/v1/users/profile (GetProfileRequest) returns (GetProfileResponse)
}

// ✅ 正确配置认证
@server(
    group: user
    jwt: Auth
)
service user-api {
    @handler getUserProfile
    get /api/v1/users/profile (GetProfileRequest) returns (GetProfileResponse)
}
```

#### 1.4 中间件配置
**规则**: `[rule:go-zero-middleware]`

**检查要点**:
- 中间件声明是否正确
- 中间件应用范围是否合理
- 中间件顺序是否正确

### 2. 业务逻辑层审查（internal/logic/）

**规则标识**: `[category:go-zero-logic]`

#### 2.1 Logic 结构设计
**规则**: `[rule:go-zero-logic-structure]`

**检查要点**:
- Logic 结构体是否正确接收 `svcCtx`
- 是否正确使用依赖注入
- 构造函数命名是否符合规范（New*Logic）

**示例问题**:
```go
// ❌ 不好的设计
type GetUserLogic struct {
    ctx context.Context
    db  *sql.DB  // 直接依赖具体实现
}

// ✅ 好的设计
type GetUserLogic struct {
    logx.Logger
    ctx    context.Context
    svcCtx *svc.ServiceContext  // 通过 svcCtx 获取依赖
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
    return &GetUserLogic{
        Logger: logx.WithContext(ctx),
        ctx:    ctx,
        svcCtx: svcCtx,
    }
}
```

#### 2.2 Context 传递
**规则**: `[rule:go-zero-context]`

**检查要点**:
- 是否正确传递和使用 `context.Context`
- 链路追踪信息是否正确传递
- 是否在所有外部调用中传递 context

**示例问题**:
```go
// ❌ 未传递 context
func (l *GetUserLogic) GetUser(req *types.GetUserRequest) (*types.GetUserResponse, error) {
    user, err := l.svcCtx.UserModel.FindOne(req.Id)  // 缺少 context
    // ...
}

// ✅ 正确传递 context
func (l *GetUserLogic) GetUser(req *types.GetUserRequest) (*types.GetUserResponse, error) {
    user, err := l.svcCtx.UserModel.FindOne(l.ctx, req.Id)
    // ...
}
```

#### 2.3 错误处理
**规则**: `[rule:go-zero-error]`

**检查要点**:
- 是否使用 go-zero 的错误处理机制
- 错误码定义是否规范
- 是否正确使用 `errorx` 包
- 错误信息是否友好且不泄露敏感信息

**示例问题**:
```go
// ❌ 直接返回底层错误
func (l *GetUserLogic) GetUser(req *types.GetUserRequest) (*types.GetUserResponse, error) {
    user, err := l.svcCtx.UserModel.FindOne(l.ctx, req.Id)
    if err != nil {
        return nil, err  // 直接暴露数据库错误
    }
    // ...
}

// ✅ 使用 errorx 包装错误
func (l *GetUserLogic) GetUser(req *types.GetUserRequest) (*types.GetUserResponse, error) {
    user, err := l.svcCtx.UserModel.FindOne(l.ctx, req.Id)
    if err != nil {
        if err == model.ErrNotFound {
            return nil, errorx.NewCodeError(404, "用户不存在")
        }
        return nil, errorx.NewDefaultError("获取用户信息失败")
    }
    // ...
}
```

#### 2.4 数据库操作
**规则**: `[rule:go-zero-db]`

**检查要点**:
- 是否正确使用 sqlx 或 gorm
- 是否有 SQL 注入风险
- 事务处理是否正确
- 是否正确处理数据库连接和关闭
- 是否使用了缓存（如果配置了缓存）

**示例问题**:
```go
// ❌ SQL 注入风险
func (l *GetUserLogic) SearchUsers(keyword string) ([]*User, error) {
    query := fmt.Sprintf("SELECT * FROM users WHERE name LIKE '%%%s%%'", keyword)
    // ...
}

// ✅ 使用参数化查询
func (l *GetUserLogic) SearchUsers(keyword string) ([]*User, error) {
    query := "SELECT * FROM users WHERE name LIKE ?"
    err := l.svcCtx.DB.SelectContext(l.ctx, &users, query, "%"+keyword+"%")
    // ...
}

// ❌ 事务处理不当
func (l *TransferLogic) Transfer(from, to int64, amount float64) error {
    tx, _ := l.svcCtx.DB.Begin()
    tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", amount, from)
    tx.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", amount, to)
    tx.Commit()  // 未检查错误
}

// ✅ 正确的事务处理
func (l *TransferLogic) Transfer(from, to int64, amount float64) error {
    return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
        if err := l.deductBalance(ctx, session, from, amount); err != nil {
            return err
        }
        if err := l.addBalance(ctx, session, to, amount); err != nil {
            return err
        }
        return nil
    })
}
```

#### 2.5 缓存使用
**规则**: `[rule:go-zero-cache]`

**检查要点**:
- 是否正确使用 go-zero 的缓存机制
- 缓存键设计是否合理（避免冲突）
- 是否有缓存穿透、击穿、雪崩风险
- 缓存过期时间设置是否合理
- 是否正确处理缓存更新和删除

**示例问题**:
```go
// ❌ 缓存键设计不当
func (l *GetUserLogic) GetUser(id int64) (*User, error) {
    key := fmt.Sprintf("user:%d", id)  // 可能与其他业务冲突
    // ...
}

// ✅ 使用命名空间
func (l *GetUserLogic) GetUser(id int64) (*User, error) {
    key := fmt.Sprintf("user:profile:%d", id)
    // ...
}

// ❌ 缓存穿透风险
func (l *GetUserLogic) GetUser(id int64) (*User, error) {
    user, err := l.svcCtx.UserModel.FindOne(l.ctx, id)
    if err != nil {
        return nil, err  // 不存在的数据不缓存，导致穿透
    }
    return user, nil
}

// ✅ 缓存空值防止穿透
func (l *GetUserLogic) GetUser(id int64) (*User, error) {
    user, err := l.svcCtx.UserModel.FindOne(l.ctx, id)
    if err == model.ErrNotFound {
        // 缓存空值，设置较短过期时间
        l.svcCtx.Cache.SetWithExpire(key, nil, 60)
        return nil, errorx.NewCodeError(404, "用户不存在")
    }
    // ...
}
```

#### 2.6 RPC 调用
**规则**: `[rule:go-zero-rpc]`

**检查要点**:
- RPC 调用是否正确传递 context
- 是否有超时控制
- 是否有熔断和降级处理
- 错误处理是否正确

**示例问题**:
```go
// ❌ 未传递 context，无超时控制
func (l *GetOrderLogic) GetOrder(req *types.GetOrderRequest) (*types.GetOrderResponse, error) {
    order, err := l.svcCtx.OrderRpc.GetOrder(context.Background(), &order.GetOrderReq{
        Id: req.Id,
    })
    // ...
}

// ✅ 正确传递 context，有超时控制
func (l *GetOrderLogic) GetOrder(req *types.GetOrderRequest) (*types.GetOrderResponse, error) {
    ctx, cancel := context.WithTimeout(l.ctx, 3*time.Second)
    defer cancel()
    
    order, err := l.svcCtx.OrderRpc.GetOrder(ctx, &order.GetOrderReq{
        Id: req.Id,
    })
    if err != nil {
        // 处理超时、熔断等错误
        return nil, errorx.NewDefaultError("获取订单信息失败")
    }
    // ...
}
```

### 3. 配置文件审查（etc/*.yaml）

**规则标识**: `[category:go-zero-config]`

#### 3.1 安全配置
**规则**: `[rule:go-zero-config-security]`

**检查要点**:
- 敏感信息（密码、密钥）是否明文存储
- JWT 密钥强度是否足够（建议至少32字符）
- 是否使用环境变量或密钥管理服务
- Redis 密码是否明文存储

**示例问题**:
```yaml
# ❌ 明文存储敏感信息
Auth:
  AccessSecret: "simple123"  # 密钥太弱
  AccessExpire: 7200

Mysql:
  DataSource: "root:password123@tcp(localhost:3306)/db"  # 明文密码

# ✅ 使用环境变量
Auth:
  AccessSecret: ${JWT_SECRET}  # 从环境变量读取
  AccessExpire: 7200

Mysql:
  DataSource: ${MYSQL_DSN}
```

#### 3.2 性能配置
**规则**: `[rule:go-zero-config-performance]`

**检查要点**:
- 超时时间设置是否合理
- 连接池大小是否合理
- 缓存配置是否合理
- 日志级别是否适合生产环境

**示例问题**:
```yaml
# ❌ 配置不合理
Timeout: 100000  # 超时时间过长（100秒）
MaxConns: 10000  # 连接池过大
LogLevel: debug  # 生产环境使用 debug 级别

# ✅ 合理配置
Timeout: 3000    # 3秒超时
MaxConns: 100    # 合理的连接池大小
LogLevel: info   # 生产环境使用 info 级别
```

### 4. 生成代码处理

**规则**: `[rule:go-zero-generated]`

go-zero 会自动生成部分代码，这些代码通常不需要审查。建议在配置文件中忽略：

**推荐的忽略配置**（在 `.codereview` 中）:
```yaml
exclude_paths:
  - "internal/types/types.go"           # API 类型定义（自动生成）
  - "internal/handler/**/*handler.go"   # Handler 层（自动生成）
  - "**/*.pb.go"                        # Protobuf 生成文件
```

**需要审查的代码**:
- `internal/logic/` - 业务逻辑层（手写）
- `internal/svc/servicecontext.go` - 服务上下文（部分手写）
- `internal/config/config.go` - 配置定义（部分手写）
- `internal/middleware/` - 自定义中间件（手写）
- `internal/model/` - 数据模型（部分手写）

### 5. 项目结构审查

**规则**: `[rule:go-zero-structure]`

**检查要点**:
- 是否遵循 go-zero 推荐的项目结构
- 目录命名是否规范
- 文件组织是否合理

**标准项目结构**:
```
.
├── api/                    # API 定义文件
│   └── *.api
├── etc/                    # 配置文件
│   └── *.yaml
├── internal/
│   ├── config/            # 配置定义
│   ├── handler/           # HTTP handler（自动生成）
│   ├── logic/             # 业务逻辑
│   ├── middleware/        # 中间件
│   ├── model/             # 数据模型
│   ├── svc/               # 服务上下文
│   └── types/             # 类型定义（自动生成）
└── *.go                   # 主程序入口
```

## 最佳实践

### 1. 依赖注入
通过 `ServiceContext` 管理所有依赖，便于测试和维护。

### 2. 错误处理
使用 `errorx` 包统一错误处理，定义清晰的错误码。

### 3. 日志记录
使用 `logx` 包记录日志，包含必要的上下文信息。

### 4. 性能优化
- 合理使用缓存
- 避免 N+1 查询
- 使用连接池
- 设置合理的超时时间

### 5. 安全性
- 输入验证
- SQL 参数化查询
- 敏感信息加密存储
- 使用 JWT 进行认证

## 参考资源

- [go-zero 官方文档](https://go-zero.dev/)
- [go-zero GitHub](https://github.com/zeromicro/go-zero)
- [go-zero 最佳实践](https://go-zero.dev/docs/tutorials)
