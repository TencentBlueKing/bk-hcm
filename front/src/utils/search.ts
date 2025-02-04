import type { ParsedQs } from 'qs';
import merge from 'lodash/merge';
import { ModelProperty, ModelPropertyType } from '@/model/typings';
import { findProperty } from '@/model/utils';
import { QueryFilterType, QueryRuleOPEnum, RulesItem } from '@/typings';

type DateRangeType = Record<'toady' | 'last7d' | 'last15d' | 'last30d', () => [Date[], Date[]]>;
type RuleItemOpVal = Omit<RulesItem, 'field'>;
type GetDefaultRule = (property: ModelProperty, custom?: RuleItemOpVal) => RuleItemOpVal;

export const getDefaultRule: GetDefaultRule = (property, custom) => {
  const { EQ, AND, IN } = QueryRuleOPEnum;
  const searchOp = property?.meta?.search?.op;

  const defaultMap: Record<ModelPropertyType, RuleItemOpVal> = {
    string: { op: searchOp || EQ, value: [] },
    number: { op: searchOp || EQ, value: '' },
    enum: { op: searchOp || IN, value: [] },
    datetime: { op: AND, value: [] },
    user: { op: searchOp || IN, value: [] },
    account: { op: searchOp || EQ, value: '' },
    array: { op: searchOp || IN, value: [] },
    bool: { op: searchOp || EQ, value: '' },
    cert: { op: searchOp || IN, value: [] },
    ca: { op: searchOp || EQ, value: '' },
    region: { op: searchOp || IN, value: [] },
  };

  return {
    ...defaultMap[property.type],
    ...custom,
  };
};

export const convertValue = (
  value: string | number | string[] | number[] | ParsedQs | ParsedQs[],
  property: ModelProperty,
  operator?: QueryRuleOPEnum,
) => {
  const { type } = property || {};
  const { IN, JSON_OVERLAPS } = QueryRuleOPEnum;

  if (type === 'number') {
    return Number(value);
  }

  // 时间范围值为['','']时
  if (type === 'datetime' && Array.isArray(value)) {
    if (!value.filter((val) => val).length) {
      return undefined;
    }
  }

  if ([IN, JSON_OVERLAPS].includes(operator)) {
    if (!Array.isArray(value)) {
      return [value];
    }
  }

  return value;
};

export const transformSimpleCondition = (condition: Record<string, any>, properties: ModelProperty[]) => {
  const queryFilter: QueryFilterType = { op: 'and', rules: [] };
  for (const [id, value] of Object.entries(condition || {})) {
    const property = findProperty(id, properties);
    if (!property) {
      continue;
    }

    // 忽略空值
    if ([null, undefined].includes(value) || !value?.length) {
      continue;
    }

    if (property.type === 'datetime' && Array.isArray(value)) {
      queryFilter.rules.push({
        op: QueryRuleOPEnum.AND,
        rules: [
          {
            op: QueryRuleOPEnum.GTE,
            field: id,
            value: convertValue(value?.[0], property, QueryRuleOPEnum.GTE) as RulesItem['value'],
          },
          {
            op: QueryRuleOPEnum.LTE,
            field: id,
            value: convertValue(value?.[1], property, QueryRuleOPEnum.LTE) as RulesItem['value'],
          },
        ],
      });
      continue;
    }

    const { op } = getDefaultRule(property);
    queryFilter.rules.push({
      op,
      field: id,
      value: convertValue(value, property, op) as RulesItem['value'],
    });
  }

  return queryFilter;
};

export const enableCount = (params = {}, enable = false) => {
  if (enable) {
    return Object.assign({}, params, { page: { count: true } });
  }
  return merge({}, params, { page: { count: false } });
};

export const onePageParams = () => ({ start: 0, limit: 1 });

export const maxPageParams = (max = 500) => ({ start: 0, limit: max });

export const getDateRange = (key: keyof DateRangeType) => {
  const dateRange = {
    toady() {
      const end = new Date();
      const start = new Date(end.getFullYear(), end.getMonth(), end.getDate());
      return [start, end];
    },
    last7d() {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
      return [start, end];
    },
    last15d() {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 15);
      return [start, end];
    },
    last30d() {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 30);
      return [start, end];
    },
  };
  return dateRange[key]();
};

export const getDateShortcutRange = () => {
  const shortcutsRange = [
    {
      text: '今天',
      value: () => getDateRange('toady'),
    },
    {
      text: '近7天',
      value: () => getDateRange('last7d'),
    },
    {
      text: '近15天',
      value: () => getDateRange('last15d'),
    },
    {
      text: '近30天',
      value: () => getDateRange('last30d'),
    },
  ];
  return shortcutsRange;
};
