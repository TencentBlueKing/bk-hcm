/**
 * 查询规则项
 */
interface QueryRule {
  field: string;
  operator: string;
  value: string | number | string[] | number[];
}

/**
 * 查询条件
 */
interface QueryConditions {
  condition: 'AND' | 'OR' | '';
  rules: (QueryRule | QueryConditions)[];
}

/**
 * 简单条件项类型
 * 可以是条件字符串 'AND' | 'OR'，或者是规则数组 [field, operator, value]，或者是嵌套条件
 */
type SimpleConditionItem = string | [string, string, unknown] | SimpleCondition;

/**
 * 简单条件数组类型
 */
type SimpleCondition = [string, ...SimpleConditionItem[]];

/**
 * 对 query builder 做校验和清理；简化 query builder 的使用
 * @param simpleConditions 简单版的查询条件
 * @returns 复杂版的查询条件
 */
export function transferSimpleConditions(simpleConditions: SimpleCondition): QueryConditions;
