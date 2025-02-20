import type { ParsedQs } from 'qs';
import merge from 'lodash/merge';
import { ModelPropertyGeneric, ModelPropertySearch, ModelPropertyType } from '@/model/typings';
import { findProperty } from '@/model/utils';
import { ISearchSelectValue, QueryFilterType, QueryRuleOPEnum, RulesItem } from '@/typings';
import dayjs from 'dayjs';
import isoWeek from 'dayjs/plugin/isoWeek';

dayjs.extend(isoWeek);

type DateRangeType = Record<
  'today' | 'last7d' | 'last15d' | 'last30d' | 'last90d' | 'last120d' | 'last180d' | 'naturalMonth' | 'naturalIsoWeek',
  () => [Date[], Date[]]
>;
type RuleItemOpVal = Omit<RulesItem, 'field'>;
type GetDefaultRule = (property: ModelPropertySearch, custom?: RuleItemOpVal) => RuleItemOpVal;

export const getDefaultRule: GetDefaultRule = (property, custom) => {
  const { EQ, AND, IN } = QueryRuleOPEnum;
  const searchOp = property.op || property?.meta?.search?.op;

  const defaultMap: Record<ModelPropertyType, RuleItemOpVal> = {
    string: { op: searchOp || EQ, value: [] },
    number: { op: searchOp || EQ, value: '' },
    enum: { op: searchOp || IN, value: [] },
    datetime: { op: AND, value: [] },
    user: { op: searchOp || IN, value: [] },
    account: { op: searchOp || IN, value: [] },
    array: { op: searchOp || IN, value: [] },
    bool: { op: searchOp || EQ, value: '' },
    cert: { op: searchOp || IN, value: [] },
    ca: { op: searchOp || EQ, value: '' },
    region: { op: searchOp || IN, value: [] },
    business: { op: searchOp || IN, value: [] },
    'req-type': { op: searchOp || IN, value: [] },
    'req-stage': { op: searchOp || IN, value: [] },
  };

  return {
    ...defaultMap[property.type],
    ...custom,
  };
};

export const convertValue = (
  value: string | number | string[] | number[] | ParsedQs | ParsedQs[],
  property: ModelPropertySearch,
  operator?: QueryRuleOPEnum,
) => {
  const { type, format, meta } = property || {};
  const { IN, JSON_OVERLAPS } = QueryRuleOPEnum;

  const formatter = format || meta?.search?.format;
  if (formatter) {
    return formatter(value);
  }

  if (['number', 'business'].includes(type)) {
    if (Array.isArray(value)) {
      return value.map((val) => Number(val));
    }
    if ([IN, JSON_OVERLAPS].includes(operator) && !Array.isArray(value)) {
      return [Number(value)];
    }
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

export const transformSimpleCondition = (condition: Record<string, any>, properties: ModelPropertyGeneric[]) => {
  const queryFilter: QueryFilterType = { op: 'and', rules: [] };

  for (const [id, value] of Object.entries(condition || {})) {
    const property = findProperty(id, properties);
    if (!property || isValueEmpty(value)) {
      continue;
    }

    if (property.meta?.search?.filterRules) {
      queryFilter.rules.push(property.meta.search.filterRules(value));
      continue;
    }

    const rule = createQueryRule(id, value, property);
    if (rule) {
      queryFilter.rules.push(rule);
    }
  }

  return queryFilter;
};

const isValueEmpty = (value: any): boolean =>
  [null, undefined, ''].includes(value) || (Array.isArray(value) && !value.length);

const createQueryRule = (id: string, value: any, property: ModelPropertyGeneric): RulesItem => {
  if (property.type === 'datetime' && Array.isArray(value)) {
    return {
      op: QueryRuleOPEnum.AND,
      rules: [
        createRuleItem(id, value[0], property, QueryRuleOPEnum.GTE),
        createRuleItem(id, value[1], property, QueryRuleOPEnum.LTE),
      ],
    };
  }

  const { op } = getDefaultRule(property);
  if (Array.isArray(value)) {
    return value.length === 1
      ? createRuleItem(id, value[0], property, op)
      : {
          op: QueryRuleOPEnum.OR,
          rules: value.map((val) => createRuleItem(id, val, property, op)),
        };
  }

  return createRuleItem(id, value, property, op);
};

const createRuleItem = (field: string, value: any, property: ModelPropertyGeneric, op: QueryRuleOPEnum): RulesItem => ({
  op,
  field,
  value: convertValue(value, property, op) as RulesItem['value'],
});

export const transformFlatCondition = (condition: Record<string, any>, properties: ModelPropertyGeneric[]) => {
  const params: Record<string, any> = {};
  for (const [id, value] of Object.entries(condition || {})) {
    const property = findProperty(id, properties) as ModelPropertySearch;

    if (!property || isValueEmpty(value)) {
      continue;
    }

    const converter = property.converter || property.meta?.search?.converter;
    if (converter) {
      Object.assign(params, converter(value));
      continue;
    }

    params[id] = convertValue(value, property);
  }

  return params;
};

// 获取简单搜索条件 - search-select
export const getSimpleConditionBySearchSelect = (
  searchValue: ISearchSelectValue,
  options: Array<{ field: string; formatter: Function }> = [],
) => {
  // 非数组，直接返回空函数，不过滤
  if (!Array.isArray(searchValue)) return null;

  const applyFormatters = (value: string, field: string) =>
    options.find((opt) => opt.field === field)?.formatter(value) || value;

  // 将搜索值转换为 rules，rule之间为AND关系，rule.values之间为OR关系
  return Object.fromEntries(
    searchValue.reduce((conditionMap, { id, values }) => {
      const formattedValues = values.map((v) => applyFormatters(v.id, id));
      conditionMap.set(id, [...(conditionMap.get(id) || []), ...formattedValues]);
      return conditionMap;
    }, new Map<string, Array<string>>()),
  );
};

// 处理本地搜索，返回一个filterFn - search-select
export const getLocalFilterFnBySearchSelect = (
  searchValue: ISearchSelectValue,
  options: Array<{ field: string; formatter: Function }> = [],
) => {
  const condition = getSimpleConditionBySearchSelect(searchValue, options) ?? {};
  const rules = Object.entries(condition).map(([key, values]) => ({ key, values }));

  // 构建过滤函数
  return (item: any) =>
    rules.every(
      ({ key, values }) =>
        // 将itemValues转为字符串，这样既可以比较数字，又可以比较字符串和字符串数组
        item[key] && values.some((v) => String(item[key]).includes(v)),
    );
};

export const enableCount = (params = {}, enable = false) => {
  if (enable) {
    return Object.assign({}, params, { page: { count: true } });
  }
  return merge({}, params, { page: { count: false } });
};

export const onePageParams = () => ({ start: 0, limit: 1 });

export const maxPageParams = (max = 500) => ({ start: 0, limit: max });

export const getDateRange = (key: keyof DateRangeType, include?: boolean) => {
  const calculateRange = (days: number) => {
    const end = new Date();
    const start = new Date();
    start.setTime(end.getTime() - 3600 * 1000 * 24 * (include ? days : days - 1));
    return [start, end];
  };

  const dateRange = {
    today: () => {
      const end = new Date();
      const start = new Date(end.getFullYear(), end.getMonth(), end.getDate());
      return [start, end];
    },
    last7d: () => calculateRange(7),
    last15d: () => calculateRange(15),
    last30d: () => calculateRange(30),
    last90d: () => calculateRange(90),
    last120d: () => calculateRange(120),
    last180d: () => calculateRange(180),
    naturalMonth: () => {
      const now = dayjs();
      const start = now.startOf('month').toDate();
      const end = now.endOf('month').toDate();
      return [start, end];
    },
    naturalIsoWeek: () => {
      const now = dayjs();
      const start = now.startOf('isoWeek').toDate();
      const end = now.endOf('isoWeek').toDate();
      return [start, end];
    },
  };
  return dateRange[key]();
};

export const getDateShortcutRange = (include?: boolean) => {
  const shortcutsRange = [
    {
      text: '今天',
      value: () => getDateRange('today', include),
    },
    {
      text: '近7天',
      value: () => getDateRange('last7d', include),
    },
    {
      text: '近15天',
      value: () => getDateRange('last15d', include),
    },
    {
      text: '近30天',
      value: () => getDateRange('last30d', include),
    },
    {
      text: '近90天',
      value: () => getDateRange('last90d', include),
    },
    {
      text: '近180天',
      value: () => getDateRange('last180d', include),
    },
  ];
  return shortcutsRange;
};

export const convertDateRangeToObject = (dateRange: Date[]) => {
  const start = new Date(dateRange[0]);
  const end = new Date(dateRange[1]);

  return {
    start: { year: start.getFullYear(), month: start.getMonth() + 1, day: start.getDate() },
    end: { year: end.getFullYear(), month: end.getMonth() + 1, day: end.getDate() },
  };
};
