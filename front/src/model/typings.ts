import type { VNode } from 'vue';
import type { Column as TableColumn } from 'bkui-vue/lib/table/props';
import { RulesItem, QueryRuleOPEnum } from '@/typings';
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
  | 'region'
  | 'business'
  | 'req-type'
  | 'req-stage';

export type ModelPropertyMeta = {
  display?: PropertyDisplayConfig;
  search?: PropertySearchConfig;
  column?: PropertyColumnConfig;
  form?: PropertyFormConfig;
};

// 模型的基础字段，与业务场景无关
export type ModelProperty = {
  id: string;
  name: string;
  type: ModelPropertyType;
  resource?: ResourceTypeEnum;
  option?: Record<string, any>;
  meta?: ModelPropertyMeta;
  unit?: string;
  index?: number;
};

export type PropertyColumnConfig = {
  sort?: boolean;
  align?: 'left' | 'center' | 'right';
  render?: (args: {
    cell?: any;
    data?: any;
    row?: any;
    column: TableColumn;
    index: number;
    rows?: any[];
  }) => VNode | boolean | number | string;
  width?: number | string;
  minWidth?: number | string;
  defaultHidden?: boolean;
};

export type PropertyFormConfig = {
  rules?: object;
};

export type PropertySearchConfig = {
  op?: QueryRuleOPEnum;
  filterRules?: (value: any) => RulesItem;
  format?: (value: any) => any;
  converter?: (value: any) => Record<string, any>;
};

export type PropertyDisplayConfig = {
  appearance?: string;
  format?: (value: any) => any;
};

// 与列展示场景相关，联合列的配置属性
export type ModelPropertyColumn = ModelProperty & PropertyColumnConfig;

// 与表单场景相关，联合表单的配置属性
export type ModelPropertyForm = ModelProperty & PropertyFormConfig;

// 与展示场景相关，联合展示的配置属性
export type ModelPropertyDisplay = ModelProperty & PropertyDisplayConfig;

// 与搜索场景相关，联合搜索的配置属性
export type ModelPropertySearch = ModelProperty & PropertySearchConfig;

export type ModelPropertyGeneric =
  | ModelProperty
  | ModelPropertyColumn
  | ModelPropertyForm
  | ModelPropertyDisplay
  | ModelPropertySearch;
