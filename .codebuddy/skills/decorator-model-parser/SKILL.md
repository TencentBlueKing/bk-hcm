---
name: decorator-model-parser
description: >
  This skill provides tools and guidance for parsing TypeScript decorator-based model definitions
  (using @Model and @Column decorators) into structured property arrays. It should be used when
  working with bk-hcm's decorator system to understand model schemas, generate search conditions,
  form configurations, or table column definitions. The skill includes utility functions for
  converting decorated classes to ModelPropertySearch[], ModelPropertyColumn[], and other formats.
---

# Decorator Model Parser

This skill helps parse TypeScript class definitions that use `@Model` and `@Column` decorators
into structured property arrays for various use cases (search, table columns, forms, etc.).

## When to Use

- Analyzing decorated model classes to understand their schema
- Converting `@Column` definitions to property arrays
- Generating search condition configurations
- Creating table column definitions from model classes
- Understanding the bk-hcm decorator system

## Decorator System Overview

The bk-hcm project uses a custom decorator system located at `front/src/decorator/`:

### @Model Decorator

Registers a class as a model with optional name:

```typescript
@Model('model-name')
export class MyModel { ... }
```

### @Column Decorator

Defines property metadata with type and configuration:

```typescript
@Column(type: ModelPropertyType, options?: ColumnOptions)
```

**Supported Types:**
- `string` - Text fields
- `datetime` - Date/time pickers
- `enum` - Select with predefined options
- `number` - Numeric inputs
- `user` - User selector
- `business` - Business selector
- `account` - Account selector
- `bool` - Boolean toggle
- `array` - Array fields
- `list` - List with custom items
- `region` - Region selector
- `json` - JSON editor
- `cloud-area` - Cloud area selector

**Column Options:**
- `name` - Display name (required)
- `index` - Sort order for display
- `option` - Key-value map for enum types
- `list` - Array of items for list types
- `meta` - Additional configuration for search/display/column/form

## Conversion Rules

To convert a decorated class to a property array:

1. Extract each `@Column` decorated property
2. Build the property object using:
   - `id`: Property name (e.g., `account_id`)
   - `name`: From decorator options or property name
   - `type`: First argument of @Column
   - Spread remaining options from second argument

### Example Conversion

**Input Class:**
```typescript
@Model('search-condition')
export class SearchCondition {
  @Column('datetime', { name: '操作时间', index: 0 })
  created_at: string;

  @Column('enum', { name: '状态', option: STATUS_MAP, index: 1 })
  status: string;

  @Column('user', { name: '操作人', index: 2 })
  operator: string;
}
```

**Output Array:**
```typescript
[
  { id: 'created_at', name: '操作时间', type: 'datetime', index: 0 },
  { id: 'status', name: '状态', type: 'enum', option: STATUS_MAP, index: 1 },
  { id: 'operator', name: '操作人', type: 'user', index: 2 }
]
```

## Utility Functions

Refer to `scripts/parse-decorator.ts` for utility functions that automate the conversion process.

### parseColumnDecorator

Parse a single @Column decorator into a property object.

### parseModelClass

Parse an entire decorated class file and return the property array.

### generatePropertyArray

Generate a formatted property array string from parsed decorators.

## Meta Configuration

The `meta` option supports specialized configurations:

### Search Configuration (`meta.search`)
```typescript
{
  op?: QueryRuleOPEnum;           // Query operator
  filterRules?: (value) => Rule;  // Custom filter logic
  format?: (value) => any;        // Value formatter
  enableEmpty?: boolean;          // Allow empty values
}
```

### Column Configuration (`meta.column`)
```typescript
{
  sort?: boolean;                 // Enable sorting
  width?: number | string;        // Column width
  fixed?: 'left' | 'right';       // Fixed position
  render?: (args) => VNode;       // Custom render
}
```

### Display Configuration (`meta.display`)
```typescript
{
  appearance?: string;            // Display style
  format?: (value) => any;        // Value formatter
  render?: (value) => VNode;      // Custom render
}
```

## File Structure Reference

```
front/src/decorator/
├── index.ts              # Exports Model and Column
├── typings.ts            # Type definitions
├── model/
│   └── model.ts          # @Model decorator
├── columns/
│   └── column.ts         # @Column decorator
└── metadata/
    ├── globals.ts        # Global metadata storage
    └── metadata-storage.ts
```

```
front/src/model/
└── typings.ts            # ModelProperty types
```

## Usage Workflow

1. Read the target decorated class file
2. Identify all @Column decorators and their arguments
3. For each decorator, extract:
   - Property name → `id`
   - First argument → `type`
   - Second argument options → spread into result
4. Resolve any referenced constants (like `OPTION_MAP`)
5. Output the structured array
