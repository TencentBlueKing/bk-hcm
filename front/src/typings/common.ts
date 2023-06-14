export interface Verify {
  action: string
  resource_type: string
  bk_biz_id?: number
  resource_id?: number
}
export enum QueryRuleOPEnum {
  EQ = 'eq',
  NEQ = 'neq',
  GT = 'gt',
  GTE = 'gte',
  LT = 'lt',
  LTE = 'lte',
  IN = 'in',
  CS = 'cs',
  CIS = 'cis',
  JSON_EQ = 'json_eq',
  OR = 'or',
  AND = 'and'
}

export type QueryFilterType = {
  op: 'and' | 'or';
  rules: Array<RulesItem>
};

export type RulesItem = {
  field: string;
  op: QueryRuleOPEnum;
  value: string | number | string[];
};

export interface IOption {
  id: string;
  name: string
};
