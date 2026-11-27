import { LocationQuery } from 'vue-router';
import { CLB_SPECS } from '@/common/constant';
import { QueryRuleOPEnum, RulesItem } from '@/typings';
import {
  ActiveQueryKey,
  ClbDetailsTabKey,
  LOAD_BALANCER_INSTANCE_SPEC_NAME,
  LoadBalancerInstanceSpec,
} from './constants';

type InstanceSpecSource = { exclusive?: number; sla_type?: string };

/**
 * 实例规格展示值
 * 云上把规格拆成 exclusive(是否独占型) 与 sla_type(性能容量型档位) 两个正交字段，后端只返回原始值，由前端合成一列
 */
export const getLoadBalancerInstanceSpecName = (data: InstanceSpecSource) => {
  if (data?.exclusive === 1) return LOAD_BALANCER_INSTANCE_SPEC_NAME[LoadBalancerInstanceSpec.EXCLUSIVE];
  if (data?.sla_type) return CLB_SPECS[data.sla_type] ?? data.sla_type;
  if (data?.exclusive === 0) return LOAD_BALANCER_INSTANCE_SPEC_NAME[LoadBalancerInstanceSpec.SHARED];
  return '--';
};

/**
 * 实例规格筛选项
 * 先给独占型 / 共享型两个粗粒度分类，再列各性能容量型档位
 */
export const LOAD_BALANCER_INSTANCE_SPEC_SEARCH_NAME: Record<string, string> = {
  ...LOAD_BALANCER_INSTANCE_SPEC_NAME,
  ...CLB_SPECS,
};

// 档位与共享型都只对非独占实例成立：展示时 exclusive === 1 优先判独占型，筛选不叠加这条会筛出列里显示「独占型」的行
const buildNonExclusiveRule = (slaTypeRule: RulesItem): RulesItem => ({
  op: QueryRuleOPEnum.AND,
  rules: [{ field: 'extension.exclusive', op: QueryRuleOPEnum.JSON_EQ, value: 0 }, slaTypeRule],
});

/**
 * 实例规格查询条件
 * 展示值来自 exclusive / sla_type，筛选则统一走 extension 内字段
 */
export const buildLoadBalancerInstanceSpecFilterRules = (value: string | string[]): RulesItem => {
  const specs = (Array.isArray(value) ? value : [value]).filter(Boolean);
  const slaTypes = specs.filter((spec) => spec in CLB_SPECS);
  const rules: RulesItem[] = [];

  if (specs.includes(LoadBalancerInstanceSpec.EXCLUSIVE)) {
    rules.push({ field: 'extension.exclusive', op: QueryRuleOPEnum.JSON_EQ, value: 1 });
  }
  if (specs.includes(LoadBalancerInstanceSpec.SHARED)) {
    rules.push(buildNonExclusiveRule({ field: 'extension.sla_type', op: QueryRuleOPEnum.JSON_EQ, value: '' }));
  }
  // 档位不论选几个都合成一条 json_in：pkg/runtime/filter 每层 rules 上限 10 条，逐档 json_eq 全选时会被撑满
  if (slaTypes.length) {
    rules.push(buildNonExclusiveRule({ field: 'extension.sla_type', op: QueryRuleOPEnum.JSON_IN, value: slaTypes }));
  }

  return rules.length === 1 ? rules[0] : { op: QueryRuleOPEnum.OR, rules };
};

const CLB_DETAILS_TAB_VALUES = new Set<string>(Object.values(ClbDetailsTabKey));

export const resolveClbDetailsTab = (tab: unknown): ClbDetailsTabKey => {
  return typeof tab === 'string' && CLB_DETAILS_TAB_VALUES.has(tab)
    ? (tab as ClbDetailsTabKey)
    : ClbDetailsTabKey.LISTENER;
};

/**
 * 仅保留 CLB 详情页跨实例 / 跨 tab 切换时必要的公共路由参数（当前 tab 标识与返回来源标记），
 * 避免分页、筛选、排序、一次性跳转标记等单 tab 业务参数残留或串扰。
 */
export const pickClbDetailsBaseQuery = (
  query: LocationQuery,
  options?: { tab?: ClbDetailsTabKey },
): Record<string, string> => {
  const nextQuery: Record<string, string> = {};

  if (options?.tab) {
    nextQuery[ActiveQueryKey.DETAILS] = options.tab;
  } else if (
    typeof query?.[ActiveQueryKey.DETAILS] === 'string' &&
    CLB_DETAILS_TAB_VALUES.has(query[ActiveQueryKey.DETAILS])
  ) {
    nextQuery[ActiveQueryKey.DETAILS] = query[ActiveQueryKey.DETAILS];
  }

  if (query && Object.hasOwn(query, '_f')) {
    const historyFlag = Array.isArray(query._f) ? query._f[0] : query._f;
    if (historyFlag !== null && historyFlag !== undefined) {
      nextQuery._f = String(historyFlag);
    }
  }

  return nextQuery;
};
