// 常量映射类型
export type ConstantMapRecord = Record<string, string>;
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

// 旧版的操作符，需要再补充
export enum QueryRuleOPEnumLegacy {
  EQ = 'equal',
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

export type QueryFilterTypeLegacy = {
  condition: 'AND' | 'OR';
  rules: Array<RulesItemLegacy>;
};

export type RulesItem = {
  field?: string;
  op?: QueryRuleOPEnum | QueryRuleOPEnumLegacy;
  value?: string | number | string[] | number[];
  rules?: RulesItem[];
};

export type RulesItemLegacy = {
  field?: string;
  operator?: QueryRuleOPEnumLegacy;
  value?: string | number | string[] | number[];
  rules?: RulesItem[];
};

export interface IOption {
  id: string;
  name: string;
}

// 列表接口分页参数
export interface IPageQuery {
  count?: boolean;
  start?: number;
  limit?: number;
  sort?: string;
  order?: string;
}

export type QueryBuilderType = {
  filter: QueryFilterType | QueryFilterTypeLegacy;
  page?: IPageQuery;
  fields?: string[];
};

export type QueryParamsType = {
  [key: string]: any;
  page?: IPageQuery;
};

interface IBaseResData {
  code: number;
  message: string;
  result: boolean;
}

// list 接口响应
export interface IListResData<T> extends IBaseResData {
  data: { details: T; count: number; info?: T };
}

// todo: 改名为 ICommonResData / APIResponse
// query 接口响应
export interface IQueryResData<T> extends IBaseResData {
  data: T;
}

export interface IOverviewListResData<T, D> extends IBaseResData {
  data: { details: T; count: number; overview: D };
}

export type PaginationType = {
  count: number;
  limit: number;
  current?: number;
  'limit-list'?: number[];
};

export type SortType = {
  column: {
    field: string;
  };
  type: string;
};

export type Awaitable<T> = Promise<T> | T;

export interface IBreadcrumb {
  title: string;
  display: boolean;
}

export type ISearchSelectValue = Array<{ id: string; name: string; values: { id: string; name: string }[] }>;
