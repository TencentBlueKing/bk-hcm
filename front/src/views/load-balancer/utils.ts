import { CLB_SPECS } from '@/common/constant';
import { QueryRuleOPEnum, RulesItem } from '@/typings';
import { LOAD_BALANCER_INSTANCE_SPEC_NAME, LoadBalancerInstanceSpec } from './constants';

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
