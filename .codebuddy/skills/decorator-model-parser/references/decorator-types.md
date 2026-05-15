# Decorator System Type Definitions

This reference documents the complete type system for the bk-hcm decorator framework.

## ModelPropertyType

Supported property types in `@Column` decorator:

| Type | Description | Use Case |
|------|-------------|----------|
| `string` | Text field | General text input |
| `datetime` | Date/time picker | Date ranges, timestamps |
| `enum` | Select dropdown | Fixed option selection |
| `number` | Numeric input | Quantities, IDs |
| `user` | User selector | Personnel assignment |
| `business` | Business selector | Business unit selection |
| `account` | Account selector | Cloud account selection |
| `bool` | Boolean toggle | True/false flags |
| `array` | Array field | Multiple values |
| `list` | List with items | Complex list data |
| `region` | Region selector | Geographic regions |
| `json` | JSON editor | Complex structured data |
| `cloud-area` | Cloud area selector | Cloud region/zone |
| `cert` | Certificate | SSL/TLS certificates |
| `ca` | CA certificate | Certificate authority |

## Column Options Interface

```typescript
interface ColumnOptions {
  // Required
  name: string;              // Display label

  // Optional - Common
  index?: number;            // Sort order (lower = earlier)
  
  // Optional - For enum type
  option?: Record<string | number, any>;  // Key-value option map
  
  // Optional - For list type
  list?: Array<{ [key: string]: any }>;   // List items
  
  // Optional - Resource binding
  resource?: ResourceTypeEnum;  // Associated resource type
  
  // Optional - Display
  unit?: string;             // Unit suffix (e.g., "GB", "个")
  apiOnly?: boolean;         // Only used in API, not displayed
  
  // Optional - Meta configurations
  meta?: {
    search?: PropertySearchConfig;
    column?: PropertyColumnConfig;
    display?: PropertyDisplayConfig;
    form?: PropertyFormConfig;
  };
}
```

## Meta Configuration Details

### PropertySearchConfig

Configuration for search/filter behavior:

```typescript
interface PropertySearchConfig {
  // Query operator
  op?: QueryRuleOPEnum | QueryRuleOPEnumLegacy;
  
  // Custom filter rules generator
  filterRules?: (value: any) => RulesItem;
  
  // Value format transformer
  format?: (value: any) => any;
  
  // Convert value to different field(s)
  converter?: (value: any) => Record<string, any>;
  
  // Allow empty/null values in search
  enableEmpty?: boolean;
  
  // Additional props passed to search component
  props?: Record<string, any>;
}
```

### PropertyColumnConfig

Configuration for table column behavior:

```typescript
interface PropertyColumnConfig {
  // Enable column sorting
  sort?: boolean;
  
  // Column alignment
  align?: 'left' | 'center' | 'right';
  
  // Custom cell renderer (TDesign)
  cell?: PrimaryTableCol['cell'];
  
  // Custom cell renderer (bkui-vue)
  render?: (args: {
    cell?: any;
    data?: any;
    row?: any;
    column?: TableColumn;
    col?: TableColumn;
    index: number;
    rows?: any[];
  }) => VNode | boolean | number | string;
  
  // Column width
  width?: number | string;
  minWidth?: number | string;
  
  // Hidden by default (user can show)
  defaultHidden?: boolean;
  
  // Show tooltip on overflow
  showOverflowTooltip?: boolean;
  
  // Fixed column position
  fixed?: 'left' | 'right';
  
  // Column filter configuration
  filter?: IFilterPropShape;
  
  // Ellipsis configuration (TDesign)
  ellipsis?: PrimaryTableCol['ellipsis'];
}
```

### PropertyDisplayConfig

Configuration for value display:

```typescript
interface PropertyDisplayConfig {
  // Display appearance type
  appearance?: string;
  
  // Props for appearance component
  appearanceProps?: Record<string, any>;
  
  // Value format transformer
  format?: (value: any) => any;
  
  // Custom render function
  render?: (value: any) => VNode | string;
  
  // Show tooltip on overflow
  showOverflowTooltip?: boolean;
}
```

### PropertyFormConfig

Configuration for form behavior:

```typescript
interface PropertyFormConfig {
  // Validation rules
  rules?: object;
}
```

## Query Operators

Common query operators for search:

```typescript
enum QueryRuleOPEnum {
  EQ = 'eq',           // Equal
  NEQ = 'neq',         // Not equal
  GT = 'gt',           // Greater than
  GTE = 'gte',         // Greater than or equal
  LT = 'lt',           // Less than
  LTE = 'lte',         // Less than or equal
  IN = 'in',           // In array
  NIN = 'nin',         // Not in array
  CS = 'cs',           // Contains (string)
  CIS = 'cis',         // Contains ignore case
  JSON_EQ = 'json_eq', // JSON equal
  JSON_IN = 'json_contains',
  OR = 'or',           // OR condition group
  AND = 'and',         // AND condition group
}
```

## Complete Example

```typescript
@Model('cloud-account/search-condition')
export class CloudAccountSearchCondition {
  @Column('string', { 
    name: '账号ID', 
    index: 0,
    meta: {
      search: {
        op: QueryRuleOPEnum.EQ,
      },
      column: {
        width: 200,
        fixed: 'left',
      }
    }
  })
  account_id: string;

  @Column('enum', { 
    name: '站点类型', 
    option: SITE_TYPE_MAP,
    index: 1,
    meta: {
      search: {
        op: QueryRuleOPEnum.IN,
      }
    }
  })
  site_type: string;

  @Column('user', { 
    name: '负责人', 
    index: 2,
    meta: {
      search: {
        enableEmpty: true,
      }
    }
  })
  owner: string;

  @Column('datetime', { 
    name: '创建时间', 
    index: 3,
    meta: {
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
      column: {
        width: 180,
      }
    }
  })
  created_at: string;
}
```

**Parsed Result:**

```typescript
[
  {
    id: 'account_id',
    name: '账号ID',
    type: 'string',
    index: 0,
    meta: {
      search: { op: 'eq' },
      column: { width: 200, fixed: 'left' }
    }
  },
  {
    id: 'site_type',
    name: '站点类型',
    type: 'enum',
    option: { /* SITE_TYPE_MAP contents */ },
    index: 1,
    meta: {
      search: { op: 'in' }
    }
  },
  {
    id: 'owner',
    name: '负责人',
    type: 'user',
    index: 2,
    meta: {
      search: { enableEmpty: true }
    }
  },
  {
    id: 'created_at',
    name: '创建时间',
    type: 'datetime',
    index: 3,
    meta: {
      search: { filterRules: [Function] },
      column: { width: 180 }
    }
  }
]
```
