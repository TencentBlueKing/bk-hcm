import { localStorageActions } from '@/common/util';
import { QueryRuleOPEnum, RulesItem } from '@/typings';
import { RouteLocationNormalizedLoaded } from 'vue-router';
import { decodeValueByAtob } from './common';

// 将输入的字符串形式的数字转换并格式化为指定精度的字符串表示
export function formatBillCost(value: string, fixed = 3): string {
  if (!value?.trim()) {
    return '0';
  }

  const num = parseFloat(value);
  if (isNaN(num)) {
    return '0';
  }

  return num % 1 === 0 ? num.toString() : num.toFixed(fixed);
}

export function calcBillData(last: string, current: string): string {
  const lastCost = parseFloat(formatBillCost(last));
  const currentCost = parseFloat(formatBillCost(current));
  const balance = currentCost - lastCost;
  const ratio = formatBillCost(String((balance / lastCost) * 100));
  if (balance === 0) return '0';
  if (lastCost === 0) return formatBillCost(String(currentCost * 100));
  return ratio;
}
// 环比计算
export function formatBillRatio(last: string, current: string): string {
  const ratio = calcBillData(last, current);
  return +ratio > 0 ? `+${ratio}%` : `${ratio}%`;
}

export function formatBillRatioClass(last: string, current: string): string {
  let className = 'red';
  const ratio = calcBillData(last, current);
  if (+ratio < 30 && +ratio > -30) {
    className = 'green';
  }
  return className;
}

// 账单查询规则类
export class BillSearchRules {
  rules = [] as RulesItem[];

  // 添加一条查询规则, 规则值从 url 或 local storage 中获取
  addRule(
    route: RouteLocationNormalizedLoaded,
    urlKey: string,
    field: string,
    op: QueryRuleOPEnum,
    valueParser = (value: string) => value && decodeValueByAtob(value),
  ) {
    const value = valueParser(route.query[urlKey] as string) || localStorageActions.get(urlKey, valueParser);
    value && this.rules.push({ field, op, value });
    // 支持链式调用
    return this;
  }
}
