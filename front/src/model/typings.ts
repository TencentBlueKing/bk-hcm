import { QueryRuleOPEnum } from '@/typings';
import type { ResourceTypeEnum } from '@/common/resource-constant';

export type ModelPropertyType =
  | 'string'
  | 'datetime'
  | 'enum'
  | 'number'
  | 'account'
  | 'user'
  | 'array'
  | 'bool'
  | 'cert'
  | 'ca'
  | 'region';

export type ModelPropertyMeta = {
  display: {
    appearance: string;
  };
  search: {
    op: QueryRuleOPEnum;
  };
};

// 模型的基础字段，与业务场景无关
export type ModelProperty = {
  id: string;
  name: string;
  type: ModelPropertyType;
  resource?: ResourceTypeEnum;
  option?: Record<string, any>;
  meta?: ModelPropertyMeta;
  index?: number;
};

export type ColumnConfig = {
  sort?: boolean;
};

// 与列展示场景相关，联合列的配置属性
export type ModelPropertyColumn = ModelProperty & ColumnConfig;
