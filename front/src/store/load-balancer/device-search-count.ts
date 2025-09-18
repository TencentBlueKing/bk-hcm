import { ref } from 'vue';
import { defineStore } from 'pinia';
import { resolveApiPathByBusinessId } from '@/common/util';
import http from '@/http';
import { enableCount } from '@/utils/search';
import type { IListResData } from '@/typings';
import { ILoadBalanceDeviceCondition } from '@/views/load-balancer/device/typing';

export const useLoadBalancerCountStore = defineStore('load-balancer-count', () => {
  const getCountLoading = ref(false);
  const getCount = async (condition: ILoadBalanceDeviceCondition, businessId: number) => {
    getCountLoading.value = true;
    const { vendor } = condition;
    const listener = resolveApiPathByBusinessId(
      '/api/v1/cloud',
      `vendors/${vendor}/listeners/by_topo/list`,
      businessId,
    );
    const url = resolveApiPathByBusinessId('/api/v1/cloud', `vendors/${vendor}/url_rules/by_topo/list`, businessId);
    const rs = resolveApiPathByBusinessId('/api/v1/cloud', `vendors/${vendor}/targets/by_topo/count`, businessId);
    const rsIP = resolveApiPathByBusinessId('/api/v1/cloud', `vendors/${vendor}/targets/by_topo/list`, businessId);
    try {
      const res = await Promise.all<
        [
          Promise<IListResData<any[]>>,
          Promise<IListResData<any[]>>,
          Promise<IListResData<any[]>>,
          Promise<IListResData<any[]>>,
        ]
      >([
        http.post(listener, enableCount(condition, true)),
        http.post(url, enableCount(condition, true)),
        http.post(rs, enableCount(condition, true)),
        http.post(rsIP, enableCount(condition, true)),
      ]);
      const [listenerCount, urlCount, rsCount, rsIPCount] = res.map((res) => res.data.count);

      return { listenerCount, urlCount, rsCount, rsIPCount };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      getCountLoading.value = false;
    }
  };

  return {
    getCountLoading,
    getCount,
  };
});
