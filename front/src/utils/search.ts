import type { ParsedQs } from 'qs';
import merge from 'lodash/merge';
import { ModelPropertyGeneric, ModelPropertySearch, ModelPropertyType } from '@/model/typings';
import { findProperty } from '@/model/utils';
import {
  ISearchCondition,
  ISearchSelectValue,
  QueryFilterType,
  QueryFilterTypeLegacy,
  QueryRuleOPEnum,
  QueryRuleOPEnumLegacy,
  RulesItem,
} from '@/typings';
import dayjs from 'dayjs';
import isoWeek from 'dayjs/plugin/isoWeek';
import { parseIP } from './common';
import { isString } from 'lodash';

dayjs.extend(isoWeek);

type DateRangeType = Record<
  'today' | 'last7d' | 'last15d' | 'last30d' | 'last90d' | 'last120d' | 'last180d' | 'naturalMonth' | 'naturalIsoWeek',
  () => [Date[], Date[]]
>;
type RuleItemOpVal = Omit<RulesItem, 'field'>;
type GetDefaultRule = (property: ModelPropertySearch, custom?: RuleItemOpVal) => RuleItemOpVal;

export const getDefaultRule: GetDefaultRule = (property, custom) => {
  const { EQ, AND, IN, JSON_EQ } = QueryRuleOPEnum;
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
    json: { op: searchOp || JSON_EQ, value: '' },
    'cloud-area': { op: searchOp || IN, value: [] },
  };

  return {
    ...defaultMap[property.type],
    ...custom,
  };
};

export const convertValue = (
  value: string | ParsedQs | (string | ParsedQs)[],
  property: ModelPropertySearch,
  operator?: QueryRuleOPEnum | QueryRuleOPEnumLegacy,
) => {
  const { type, format, meta } = property || {};
  const isArrayOperator = [
    QueryRuleOPEnum.IN,
    QueryRuleOPEnum.JSON_OVERLAPS,
    QueryRuleOPEnumLegacy.IN,
    QueryRuleOPEnumLegacy.JSON_OVERLAPS,
  ].includes(operator);

  const formatter = format || meta?.search?.format;
  if (formatter) {
    return formatter(value);
  }

  if (['number', 'business'].includes(type)) {
    if (Array.isArray(value)) {
      return value.map((val) => Number(val));
    }
    if (isArrayOperator && !Array.isArray(value)) {
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

  if (isArrayOperator && !Array.isArray(value)) {
    return [value];
  }

  return value;
};

const isValueEmpty = (value: any): boolean =>
  [null, undefined, ''].includes(value) || (Array.isArray(value) && !value.length);

const createRuleItem = (
  field: string,
  value: any,
  property: ModelPropertyGeneric,
  op: QueryRuleOPEnum | QueryRuleOPEnumLegacy,
  legacy?: boolean,
): RulesItem => {
  return {
    [!legacy ? 'op' : 'operator']: op,
    field,
    value: convertValue(value, property, op) as RulesItem['value'],
  };
};

export const transformSimpleCondition = (
  condition: Record<string, any>,
  properties: ModelPropertyGeneric[],
  legacy?: boolean,
) => {
  const queryFilter: QueryFilterType | QueryFilterTypeLegacy = !legacy
    ? { op: 'and', rules: [] }
    : { condition: 'AND', rules: [] };

  for (const [id, value] of Object.entries(condition || {})) {
    const property = findProperty(id, properties);

    if (!property || (!property.meta?.search?.enableEmpty && isValueEmpty(value))) {
      continue;
    }

    if (property.meta?.search?.filterRules) {
      //* 如果是search-select，可能需要配合validateValues使用，避免无效请求
      const filterRules = property.meta.search.filterRules(value);
      // 判断构建的条件是否有效，filterRules的结构为 { field, op, value } | { op, rules }
      if (filterRules && (filterRules.value || filterRules.rules.length > 0)) {
        queryFilter.rules.push(filterRules);
      }
      continue;
    }

    if (property.type === 'datetime' && Array.isArray(value)) {
      queryFilter.rules.push({
        op: QueryRuleOPEnum.AND,
        rules: [
          createRuleItem(id, value[0], property, QueryRuleOPEnum.GTE),
          createRuleItem(id, value[1], property, QueryRuleOPEnum.LTE),
        ],
      });
      continue;
    }

    const { op } = getDefaultRule(property);
    queryFilter.rules.push(createRuleItem(id, value, property, op, legacy));
  }

  return queryFilter;
};

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
  options: Array<{ field: string; formatter?: Function }> = [], // TODO: options可以设计成 Record
) => {
  // 非数组，直接返回空
  if (!Array.isArray(searchValue)) return null;

  const applyFormatters = (value: string, field: string) =>
    options.find((opt) => opt.field === field)?.formatter?.(value) ?? value;

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
  options: Array<{
    field: string;
    formatter?: Function;
    checker?: (key: string, values: string[], item: any) => boolean;
  }> = [],
) => {
  // 非数组，直接返回空函数，不过滤
  if (!Array.isArray(searchValue)) return () => true;

  const condition = getSimpleConditionBySearchSelect(searchValue, options);
  const rules = Object.entries(condition).map(([key, values]) => ({ key, values }));

  // 构建过滤函数
  return (item: any) =>
    rules.every(({ key, values }) => {
      const checker = options.find((opt) => opt.field === key)?.checker;
      if (checker) return checker(key, values, item);

      // 将itemValues转为字符串，这样既可以比较数字，又可以比较字符串和字符串数组
      // TODO: 这里不能简单处理，应该还是要往convertValue去靠
      return item[key] && values.some((v) => String(item[key]).includes(v));
    });
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

/**
 * @param condition search-qs构建的查询条件
 * @returns search-select组件的model值
 */
export const buildSearchSelectValueBySearchQsCondition = (
  condition: ISearchCondition = {},
  properties: ModelPropertySearch[] = [],
) => {
  const searchValue: ISearchSelectValue = [];
  Object.entries(condition).forEach(([id, value]) => {
    const property = properties.find((property) => property.id === id);
    const getDisplayName = (value: any, property: ModelPropertySearch) => {
      const { option } = property;
      return option?.[value] ?? value;
    };
    const values = Array.isArray(value)
      ? value.map((v) => ({ id: v, name: getDisplayName(v, property) }))
      : [{ id: value, name: getDisplayName(value, property) }];
    searchValue.push({ id, name: property.name, values });
  });
  return searchValue;
};

/**
 * @param value search-select组件的model值
 * @param field 字段名
 * @param op 查询操作符
 * @returns {RulesItem} 构建search-select场景下的filter-rules，一般用于解决CS、JSON_CONTAINS等查询，保证value的准确性
 */
export const buildFilterRulesWithSearchSelect = (value: string | string[], field: string, op: QueryRuleOPEnum) => {
  let filterRules: RulesItem = { field, op, value };
  if (Array.isArray(value) && value.length > 1) {
    filterRules = { op: QueryRuleOPEnum.OR, rules: value.map((val) => ({ field, op, value: val })) };
  }
  if (Array.isArray(value) && value.length === 1) {
    filterRules = { field, op, value: value[0] };
  }
  return filterRules;
};

/**
 * @param value ip字符串 | ip字符串数组
 * @param networkType 网络类型 public | private
 * @returns {RulesItem} ip查询条件
 */
export const buildIPFilterRules = (value: string | string[], networkType: 'public' | 'private'): RulesItem => {
  const IPResult: RulesItem = { op: QueryRuleOPEnum.OR, rules: [] };
  const IPv4Set = new Set<string>();
  const IPv6Set = new Set<string>();

  const processIP = (ip: string) => {
    const { IPv4List, IPv6List } = parseIP(ip);
    IPv4List.forEach((ip) => IPv4Set.add(ip));
    IPv6List.forEach((ip) => IPv6Set.add(ip));
  };

  if (isString(value)) {
    processIP(value);
  } else if (Array.isArray(value)) {
    value.forEach(processIP);
  }

  if (IPv4Set.size) {
    IPResult.rules.push({
      field: `${networkType}_ipv4_addresses`,
      op: QueryRuleOPEnum.JSON_OVERLAPS,
      value: Array.from(IPv4Set),
    });
  }
  if (IPv6Set.size) {
    IPResult.rules.push({
      field: `${networkType}_ipv6_addresses`,
      op: QueryRuleOPEnum.JSON_OVERLAPS,
      value: Array.from(IPv6Set),
    });
  }

  return IPResult;
};

/**
 * 构建VIP查询条件
 * @param value ip字符串 | ip字符串数组
 * @returns {RulesItem} ip查询条件
 */
export const buildVIPFilterRules = (value: string | string[]): RulesItem => {
  // 无法辨别value是私有还是公网，因此都要作为条件来查询
  const privateIpResult = buildIPFilterRules(value, 'private');
  const publicIpResult = buildIPFilterRules(value, 'public');

  return { op: QueryRuleOPEnum.OR, rules: [...privateIpResult.rules, ...publicIpResult.rules] };
};

/**
 * 构建多值查询条件(精确查询)，如：多ID
 */
export const buildMultipleValueRulesItem = (field: string, value: string) => {
  const rulesItem: RulesItem = { field, op: QueryRuleOPEnum.EQ, value: '' };

  if (value) {
    const splitValue = value
      .trim()
      .split(/\n|;|；|,|，|\|/)
      .reduce((prev, curr) => {
        if (curr.trim()) {
          prev.push(curr.trim());
        }
        return prev;
      }, []);

    if (!splitValue.length) {
      // 如果没有分割值，直接使用原始值进行搜索
      Object.assign(rulesItem, { value });
    } else {
      Object.assign(rulesItem, { op: QueryRuleOPEnum.IN, value: splitValue });
    }
  }

  return rulesItem;
};
