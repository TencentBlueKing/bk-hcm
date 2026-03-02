# Common Decorator Patterns

This reference documents common patterns and examples for using the decorator system in bk-hcm.

## Search Condition Patterns

### Basic Search Condition

Simple search with standard fields:

```typescript
@Model('resource/search-condition')
export class ResourceSearchCondition {
  @Column('string', { name: '名称', index: 0 })
  name: string;

  @Column('enum', { name: '状态', option: STATUS_MAP, index: 1 })
  status: string;

  @Column('user', { name: '负责人', index: 2 })
  owner: string;

  @Column('business', { name: '所属业务', index: 3 })
  bk_biz_id: number;
}
```

### Search with Custom Filter Rules

When you need complex query logic:

```typescript
@Column('string', {
  name: '资源名称',
  meta: {
    search: {
      filterRules(value: string | string[]) {
        // Handle multiple values with OR condition
        if (Array.isArray(value) && value.length > 1) {
          return {
            op: QueryRuleOPEnum.OR,
            rules: value.map((val) => ({ 
              field: 'res_name', 
              op: QueryRuleOPEnum.CS, 
              value: val 
            })),
          };
        }
        // Handle single value in array
        if (Array.isArray(value) && value.length === 1) {
          return { field: 'res_name', op: QueryRuleOPEnum.CS, value: value[0] };
        }
        // Handle plain string
        return { field: 'res_name', op: QueryRuleOPEnum.CS, value };
      },
    },
  },
  index: 1,
})
res_name: string;
```

### Date Range Search

For datetime fields that need range queries:

```typescript
@Column('datetime', {
  name: '创建时间',
  index: 0,
  meta: {
    search: {
      filterRules(value: [string, string]) {
        if (!value || value.length !== 2) return null;
        return {
          op: QueryRuleOPEnum.AND,
          rules: [
            { field: 'created_at', op: QueryRuleOPEnum.GTE, value: value[0] },
            { field: 'created_at', op: QueryRuleOPEnum.LTE, value: value[1] },
          ],
        };
      },
    },
  },
})
created_at: string;
```

### Enum with Dynamic Options

When options come from API or computed:

```typescript
@Column('enum', {
  name: '云厂商',
  option: async () => {
    const vendors = await fetchVendors();
    return vendors.reduce((acc, v) => ({ ...acc, [v.id]: v.name }), {});
  },
  index: 1,
})
vendor: string;
```

## Table Column Patterns

### Fixed Width Column

```typescript
@Column('string', {
  name: 'ID',
  index: 0,
  meta: {
    column: {
      width: 120,
      fixed: 'left',
    },
  },
})
id: string;
```

### Sortable Column

```typescript
@Column('datetime', {
  name: '更新时间',
  index: 5,
  meta: {
    column: {
      sort: true,
      width: 180,
    },
  },
})
updated_at: string;
```

### Custom Cell Render

```typescript
@Column('string', {
  name: '状态',
  index: 2,
  meta: {
    column: {
      render({ cell }) {
        const statusMap = {
          active: { text: '正常', color: 'green' },
          inactive: { text: '停用', color: 'red' },
        };
        const status = statusMap[cell] || { text: cell, color: 'gray' };
        return <Tag color={status.color}>{status.text}</Tag>;
      },
    },
  },
})
status: string;
```

### Hidden by Default

```typescript
@Column('string', {
  name: '备注',
  index: 10,
  meta: {
    column: {
      defaultHidden: true,
      width: 200,
    },
  },
})
memo: string;
```

## Display Patterns

### Email Masking

```typescript
@Column('string', {
  name: '邮箱',
  index: 3,
  meta: {
    display: {
      format(value: string) {
        if (!value) return '-';
        // Mask middle part: abc***@domain.com
        const [local, domain] = value.split('@');
        if (local.length <= 3) return `${local[0]}***@${domain}`;
        return `${local.slice(0, 3)}***@${domain}`;
      },
    },
  },
})
email: string;
```

### Link Display

```typescript
@Column('string', {
  name: '资源ID',
  index: 0,
  meta: {
    display: {
      render(value: string) {
        return <a href={`/resource/${value}`}>{value}</a>;
      },
    },
  },
})
resource_id: string;
```

### Boolean Display

```typescript
@Column('bool', {
  name: 'MFA已绑定',
  index: 8,
  meta: {
    display: {
      format(value: boolean) {
        return value ? '是' : '否';
      },
    },
    column: {
      render({ cell }) {
        return cell ? <CheckIcon class="success" /> : <CloseIcon class="danger" />;
      },
    },
  },
})
mfa_bindind: boolean;
```

## Form Patterns

### Required Field with Validation

```typescript
@Column('string', {
  name: '账号名称',
  index: 0,
  meta: {
    form: {
      rules: {
        required: true,
        message: '请输入账号名称',
        trigger: 'blur',
      },
    },
  },
})
name: string;
```

### Email Validation

```typescript
@Column('string', {
  name: '邮箱',
  index: 2,
  meta: {
    form: {
      rules: [
        { required: true, message: '请输入邮箱', trigger: 'blur' },
        { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' },
      ],
    },
  },
})
email: string;
```

## Complete Model Example

Here's a complete example combining multiple patterns:

```typescript
import { Model, Column } from '@/decorator';
import { QueryRuleOPEnum } from '@/typings';
import { ACCOUNT_STATUS, SITE_TYPES } from '../constants';

@Model('cloud-account/list')
export class CloudAccountModel {
  @Column('string', {
    name: '账号ID',
    index: 0,
    meta: {
      column: { width: 150, fixed: 'left' },
      search: { op: QueryRuleOPEnum.EQ },
    },
  })
  account_id: string;

  @Column('string', {
    name: '别名',
    index: 1,
    meta: {
      column: { width: 200 },
      search: { op: QueryRuleOPEnum.CS },
    },
  })
  alias: string;

  @Column('enum', {
    name: '站点类型',
    option: SITE_TYPES,
    index: 2,
    meta: {
      column: { width: 100 },
      search: { op: QueryRuleOPEnum.IN },
    },
  })
  site_type: string;

  @Column('user', {
    name: '负责人',
    index: 3,
    meta: {
      column: { width: 120 },
    },
  })
  owner: string;

  @Column('enum', {
    name: '资源纳管状态',
    option: ACCOUNT_STATUS,
    index: 4,
    meta: {
      column: {
        width: 120,
        render({ cell }) {
          const map = { managed: 'success', unmanaged: 'warning' };
          return <Tag theme={map[cell]}>{ACCOUNT_STATUS[cell]}</Tag>;
        },
      },
    },
  })
  manage_status: string;

  @Column('datetime', {
    name: '创建时间',
    index: 5,
    meta: {
      column: { width: 180, sort: true },
      search: {
        filterRules(value: [string, string]) {
          return {
            op: QueryRuleOPEnum.AND,
            rules: [
              { field: 'created_at', op: QueryRuleOPEnum.GTE, value: value[0] },
              { field: 'created_at', op: QueryRuleOPEnum.LTE, value: value[1] },
            ],
          };
        },
      },
    },
  })
  created_at: string;

  @Column('string', {
    name: '备注',
    index: 99,
    meta: {
      column: { defaultHidden: true, width: 200 },
    },
  })
  memo: string;
}
```

**Output Property Array:**

```typescript
[
  { id: 'account_id', name: '账号ID', type: 'string', index: 0, meta: { column: { width: 150, fixed: 'left' }, search: { op: 'eq' } } },
  { id: 'alias', name: '别名', type: 'string', index: 1, meta: { column: { width: 200 }, search: { op: 'cs' } } },
  { id: 'site_type', name: '站点类型', type: 'enum', option: SITE_TYPES, index: 2, meta: { column: { width: 100 }, search: { op: 'in' } } },
  { id: 'owner', name: '负责人', type: 'user', index: 3, meta: { column: { width: 120 } } },
  { id: 'manage_status', name: '资源纳管状态', type: 'enum', option: ACCOUNT_STATUS, index: 4, meta: { column: { width: 120, render: [Function] } } },
  { id: 'created_at', name: '创建时间', type: 'datetime', index: 5, meta: { column: { width: 180, sort: true }, search: { filterRules: [Function] } } },
  { id: 'memo', name: '备注', type: 'string', index: 99, meta: { column: { defaultHidden: true, width: 200 } } }
]
```
