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

export type QueryFilterType = {
  op: 'and' | 'or';
  rules: Array<RulesItem>;
};

export type RulesItem = {
  op: QueryRuleOPEnum;
  field?: string;
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
  start: number;
  limit: number;
  sort?: string;
  order?: string;
}

export type QueryBuilderType = {
  filter: QueryFilterType;
  page?: IPageQuery;
  fields?: string[];
};

interface IBaseResData {
  code: number;
  message: string;
}

// list 接口响应
export interface IListResData<T> extends IBaseResData {
  data: { details: T; count: number };
}

// todo: 改名为 ICommonResData / APIResponse
// query 接口响应
export interface IQueryResData<T> extends IBaseResData {
  data: T;
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

export interface DropDownPopover {
  trigger: 'manual' | 'click' | 'hover';
  forceClickoutside: boolean;
}
