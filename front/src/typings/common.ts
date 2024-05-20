export interface Verify {
  action: string;
  resource_type: string;
  bk_biz_id?: number;
  resource_id?: number;
}
export enum QueryRuleOPEnum {
  EQ = 'eq',
  NEQ = 'neq',
  GT = 'gt',
  GTE = 'gte',
  LT = 'lt',
  LTE = 'lte',
  IN = 'in',
  NIN = 'nin',
  CS = 'cs',
  CIS = 'cis',
  JSON_EQ = 'json_eq',
  JSON_NEQ = 'json_neq',
  JSON_OVERLAPS = 'json_overlaps',
  OR = 'or',
  AND = 'and',
  JSON_CONTAINS = 'json_contains',
}

export type QueryFilterType = {
  op: 'and' | 'or';
  rules: Array<RulesItem>;
};

export type RulesItem = {
  field: string;
  op: QueryRuleOPEnum;
  value: string | number | string[] | number[];
};

export interface IOption {
  id: string;
  name: string;
}

// 列表接口分页参数
export interface IPageQuery {
  count?: boolean;
  start: number;
  limit: number;
  sort?: string;
  order?: string;
}

interface IBaseResData {
  code: number;
  message: string;
}

// list 接口响应
export interface IListResData<T> extends IBaseResData {
  data: { details: T; count: number };
}

// query 接口响应
export interface IQueryResData<T> extends IBaseResData {
  data: T;
}
