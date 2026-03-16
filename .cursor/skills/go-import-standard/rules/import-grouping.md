# 导入分组规范

## 分组顺序

导入语句必须按以下顺序分为三组，**组与组之间用空行分隔**：

1. **标准库** - Go 标准库包
2. **内部包** - 项目内部包（以 `hcm/` 开头）
3. **第三方包** - 外部依赖包（`github.com/`、`go.opentelemetry.io/`、`gopkg.in/` 等）

## 标准格式示例

```go
import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
	"hcm/pkg/logs"

	"github.com/emicklei/go-restful/v3"
	"github.com/prometheus/client_golang/prometheus"
)
```

> **文件参考**:

## 符合度

- **符合度**: ~98%
- 项目中几乎所有文件都严格遵循此分组顺序
- `goimports` 工具会自动格式化为此顺序

## 常见错误

### ❌ 错误：未分组

```go
import (
	"context"
	"hcm/pkg/logs"
	"github.com/jmoiron/sqlx"
	"fmt"
)
```

### ✅ 正确：分组并用空行分隔

```go
import (
	"context"
	"fmt"

	"hcm/pkg/logs"

	"github.com/jmoiron/sqlx"
)
```

### ❌ 错误：内部包和第三方包顺序颠倒

```go
import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"hcm/pkg/logs"
)
```

## 工具支持

使用 `goimports` 自动格式化导入：

```bash
goimports -w .
```

在 GoLand/VS Code 中配置保存时自动运行 `goimports`，可确保导入始终符合规范。
