import { ref } from 'vue';
import { defineStore } from 'pinia';
import { resolveApiPathByBusinessId } from '@/common/util';
import http from '@/http';
import type { IListResData, IPageQuery } from '@/typings';
import { localSort } from '@/utils/search';
import { ILoadBalanceDeviceCondition } from '@/views/load-balancer/device/typing';
import { IListenerItem } from './listener';

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

export const useLoadBalancerDeviceSearchStore = defineStore('load-balancer-device-search', () => {
  const overviewListLoading = ref(false);
  const listenerListLoading = ref(false);
  const urlRuleListLoading = ref(false);

  const listenerList = ref<IListenerItem[]>();
  const urlRuleList = ref<IUrlRuleItem[]>();

  const listenerCount = ref(0);
  const urlCount = ref(0);
  const rsCount = ref(0);

  const getOverviewList = async (condition: ILoadBalanceDeviceCondition, businessId: number) => {
    overviewListLoading.value = true;

    const { vendor } = condition;

    const listener = resolveApiPathByBusinessId(
      '/api/v1/cloud',
      `vendors/${vendor}/listeners/by_topo/list`,
      businessId,
    );
    const url = resolveApiPathByBusinessId('/api/v1/cloud', `vendors/${vendor}/url_rules/by_topo/list`, businessId);
    const rs = resolveApiPathByBusinessId('/api/v1/cloud', `vendors/${vendor}/targets/by_topo/count`, businessId);

    try {
      const [listenerRes, urlRuleRes, rsRes] = await Promise.all<
        [Promise<IListResData<any[]>>, Promise<IListResData<any[]>>, Promise<IListResData<any[]>>]
      >([
        http.post(listener, condition, { globalError: false }),
        http.post(url, condition, { globalError: false }),
        http.post(rs, condition, { globalError: false }),
      ]);

      if (listenerRes?.code === 0) {
        listenerList.value = localSort(listenerRes?.data?.details ?? [], {
          column: { field: 'created_at' },
          type: 'DESC',
        });
        listenerCount.value = listenerRes?.data?.count ?? 0;
      } else if (listenerRes?.code === 2000026) {
        listenerList.value = [];
        listenerCount.value = Number.MAX_SAFE_INTEGER;
      }

      if (urlRuleRes?.code === 0) {
        urlRuleList.value = localSort(urlRuleRes?.data?.details ?? [], {
          column: { field: 'created_at' },
          type: 'DESC',
        });
        urlCount.value = urlRuleRes?.data?.count ?? 0;
      } else if (urlRuleRes?.code === 2000026) {
        urlRuleList.value = [];
        urlCount.value = Number.MAX_SAFE_INTEGER;
      }

      if (rsRes?.code === 0) {
        rsCount.value = rsRes?.data?.count ?? 0;
      } else if (rsRes?.code === 2000026) {
        rsCount.value = Number.MAX_SAFE_INTEGER;
      }

      return {
        listenerList: listenerList.value,
        urlRuleList: urlRuleList.value,
        listenerCount: listenerCount.value,
        urlCount: urlCount.value,
        rsCount: rsCount.value,
      };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      overviewListLoading.value = false;
    }
  };

  const getListenerList = async (condition: ILoadBalanceDeviceCondition, page: IPageQuery, businessId: number) => {
    if (listenerList.value) {
      return {
        list: listenerList.value,
        count: listenerCount.value,
      };
    }
    const { vendor } = condition;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', `vendors/${vendor}/listeners/by_topo/list`, businessId);
    listenerListLoading.value = true;
    try {
      const res: IListResData<IListenerItem[]> = await http.post(api, condition, { globalError: false });
      if (res?.code === 0) {
        listenerList.value = localSort(res?.data?.details ?? [], {
          column: { field: page.sort || 'created_at' },
          type: page.order || 'DESC',
        });
        listenerCount.value = res?.data?.count ?? 0;
      } else if (res?.code === 2000026) {
        listenerList.value = [];
        listenerCount.value = Number.MAX_SAFE_INTEGER;
      }
      return {
        list: listenerList.value,
        count: listenerCount.value,
      };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      listenerListLoading.value = false;
    }
  };

  const getUrlRuleList = async (condition: ILoadBalanceDeviceCondition, page: IPageQuery, businessId: number) => {
    if (urlRuleList.value) {
      return {
        list: urlRuleList.value,
        count: urlCount.value,
      };
    }
    const { vendor } = condition;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', `vendors/${vendor}/url_rules/by_topo/list`, businessId);
    urlRuleListLoading.value = true;
    try {
      const res: IListResData<IUrlRuleItem[]> = await http.post(api, condition, { globalError: false });
      if (res?.code === 0) {
        urlRuleList.value = localSort(res?.data?.details ?? [], {
          column: { field: page.sort || 'created_at' },
          type: page.order || 'DESC',
        });
        urlCount.value = res?.data?.count ?? 0;
      } else if (res?.code === 2000026) {
        urlRuleList.value = [];
        urlCount.value = Number.MAX_SAFE_INTEGER;
      }
      return {
        list: urlRuleList.value,
        count: urlCount.value,
      };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      urlRuleListLoading.value = false;
    }
  };

  return {
    overviewListLoading,
    getOverviewList,
    getListenerList,
    listenerListLoading,
    getUrlRuleList,
    urlRuleListLoading,
  };
});
