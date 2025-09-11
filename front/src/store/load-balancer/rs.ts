import { ref } from 'vue';
import { defineStore } from 'pinia';
import { resolveApiPathByBusinessId } from '@/common/util';
import http from '@/http';
import { enableCount } from '@/utils/search';
import type { IListResData, IPageQuery } from '@/typings';
import { ILoadBalanceDeviceCondition } from '@/views/load-balancer/device/common';
import { VendorEnum } from '@/common/constant';

export interface IRsItem {
  inst_id: string;
  cloud_vpc_ids: string[];
  inst_type: string;
  ip: string;
  zone: string;
  targets: string[];
  target_count: number;
}

export const useLoadBalancerRsStore = defineStore('load-balancer-rs', () => {
  const getListLoading = ref(false);
  const getRsList = async (condition: ILoadBalanceDeviceCondition, page: IPageQuery, businessId: number) => {
    getListLoading.value = true;
    const { vendor } = condition;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', `vendors/${vendor}/targets/by_topo/list`, businessId);
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<IRsItem[]>>, Promise<IListResData<IRsItem[]>>]
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
      getListLoading.value = false;
    }
  };

  const batchUpdateWeightLoading = ref(false);
  const batchUpdateWeight = async (
    params: { account_id: string; target_ids: string[]; new_weight: number },
    businessId: number,
  ) => {
    batchUpdateWeightLoading.value = true;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', `targets/weight`, businessId);
    try {
      const res = await http.patch(api, params);
      return res;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      batchUpdateWeightLoading.value = false;
    }
  };

  const batchUnbindLoading = ref(false);
  const batchUnbind = async (params: { account_id: string; target_ids: string[] }, businessId: number) => {
    batchUnbindLoading.value = true;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', `target_groups/targets/batch`, businessId);
    try {
      const res = await http.delete(api, { data: params });
      return res;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      batchUnbindLoading.value = false;
    }
  };

  const batchExportLoading = ref(false);
  const batchExport = async (params: { target_ids: string[] }, businessId: number, vendor: VendorEnum) => {
    batchExportLoading.value = true;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', `vendors/${vendor}/targets/export`, businessId);
    try {
      const res = await http.download({
        url: api,
        data: params,
        globalError: false,
      });
      return res;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      batchExportLoading.value = false;
    }
  };

  return {
    getListLoading,
    getRsList,
    batchUpdateWeightLoading,
    batchUpdateWeight,
    batchUnbindLoading,
    batchUnbind,
    batchExportLoading,
    batchExport,
  };
});
