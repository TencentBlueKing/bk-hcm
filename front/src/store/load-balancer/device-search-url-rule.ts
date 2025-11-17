import { ref } from 'vue';
import { defineStore } from 'pinia';
import { resolveApiPathByBusinessId } from '@/common/util';
import http from '@/http';
import { ILoadBalanceDeviceCondition } from '@/views/load-balancer/device/typing';

export interface IUrlRuleItem {
  id: string;
  ip: string[];
  lbl_protocols: string;
  lbl_port: number;
  rule_url: string;
  rule_domain: string[];
  target_count: number;
  listener_id: string;
}

export const useLoadBalancerUrlRuleStore = defineStore('load-balancer-url-rule', () => {
  const urlRuleListLoading = ref(false);
  const getUrlRuleList = async (condition: ILoadBalanceDeviceCondition, businessId: number) => {
    urlRuleListLoading.value = true;
    const { vendor } = condition;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', `vendors/${vendor}/url_rules/by_topo/list`, businessId);
    try {
      const res = await http.post(api, condition);

      const list = res?.data?.details ?? [];
      const count = res?.data?.count ?? 0;

      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      urlRuleListLoading.value = false;
    }
  };

  return {
    urlRuleListLoading,
    getUrlRuleList,
  };
});
