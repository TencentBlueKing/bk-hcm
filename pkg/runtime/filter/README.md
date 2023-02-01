# filter

filter 是一种查询表达式，它可以根据自身规则内容生成对应的Sql语句。

## 特性
1. 支持多种查询操作符。
2. 支持JSON字段操作符：=、in。
2. 支持多种 value 类型。
3. 支持嵌套。


## 函数功能说明
- Validate(opt *ExprOption) (hitErr error) - 用于校验Filter的合法性，ExprOption传入特定的限制参数。
- SqlWhereExpr(opt *SQLWhereOption) (where string, err error) - 生成SQL查询语句，SQLWhereOption可以设置参数优先级和扩展查询条件。
- UnmarshalJSON(raw []byte) error - 自定义JSON序列化函数
- LogMarshal() string - 自定义日志打印Filter函数
- WithType() RuleType - 实现RuleFactory，嵌套表达式需要，返回表达式类型。
- RuleField() string - 实现RuleFactory，嵌套表达式需要，返回表达式字段，没用用到。
- SQLExprAndValue(opt *SQLWhereOption) (string, map[string]interface{}, error) - 实现RuleFactory，嵌套表达式需要，生成子查询SQL语句和返回对应的映射值。

注：
1. 需注意当使用JSON字段操作符时，字段名仅需要将嵌套字段通过 '.' 关联即可。e.g: "extension.vpc.id"
2. JSON字段操作符返回的SQL语句和值映射中，映射Key为避免和其他字段名冲突，采用将 "extension.vpc.id" 转为 "extensionvpcid" 当作映射Key。
   生成Sql语句如下 'select * from security_group where extension->>"$.vpc.id" = :extensionvpcid'

## 示例
1. 名称为Jim，且年龄大于18岁。
```go
expr := &Expression{
    Op: And,
    Rules: []RuleFactory{
        &AtomRule{
            Field: "name",
            Op:    Equal.Factory(),
            Value: "Jim",
        },
        &AtomRule{
            Field: "age",
            Op:    GreaterThan.Factory(),
            Value: 18,
        },
    },
}
```

2. 名称为Jim，且年龄大于18或者身高小于1.8。
```go
expr := &Expression{
    Op: And,
    Rules: []RuleFactory{
        &AtomRule{
            Field: "name",
            Op:    Equal.Factory(),
            Value: "Jim",
        },
        &Expression{
            Op: Or,
            Rules: []RuleFactory{
                &AtomRule{
                    Field: "age",
                    Op:    GreaterThan.Factory(),
                    Value: 18,
                },
                &AtomRule{
                    Field: "height",
                    Op:    LessThan.Factory(),
                    Value: 1.8,
                },
            },
        },
    },
}
```

3. 名称为Jim，且 Extension Json字段中vpc的id为3。
```go
expr := &Expression{
   Op: And,
   Rules: []RuleFactory{
      &AtomRule{
         Field: "name",
         Op:    Equal.Factory(),
         Value: "Jim",
      },
      &AtomRule{
         Field: "extension.vpc.id",
         Op:    Equal.Factory(),
         Value: 3,
      },
   },
}
```
