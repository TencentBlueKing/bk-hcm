import { ref } from 'vue';
import { defineStore } from 'pinia';
import { resolveApiPathByBusinessId } from '@/common/util';
import http from '@/http';
import { enableCount } from '@/utils/search';
import type { IListResData, IPageQuery } from '@/typings';
import { ILoadBalanceDeviceCondition } from '@/views/load-balancer/device/common';

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
  const getUrlRuleList = async (condition: ILoadBalanceDeviceCondition, page: IPageQuery, businessId: number) => {
    urlRuleListLoading.value = true;
    const { vendor } = condition;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', `vendors/${vendor}/url_rules/by_topo/list`, businessId);
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<IUrlRuleItem[]>>, Promise<IListResData<IUrlRuleItem[]>>]
      >([
        http.post(api, enableCount({ ...condition, page }, false)),
        http.post(api, enableCount({ ...condition, page }, true)),
      ]);

      const list = listRes?.data?.details ?? [];
      const count = countRes?.data?.count ?? 0;

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
