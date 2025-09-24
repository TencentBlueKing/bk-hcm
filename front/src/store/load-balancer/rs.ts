import { ref } from 'vue';
import { defineStore } from 'pinia';
import { resolveApiPathByBusinessId } from '@/common/util';
import http from '@/http';
import { enableCount } from '@/utils/search';
import type { IListResData, IPageQuery } from '@/typings';
import { ILoadBalanceDeviceCondition } from '@/views/load-balancer/device/typing';
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
  // 获取设备检索-RS列表
  const getRsList = async (condition: ILoadBalanceDeviceCondition, page: IPageQuery, businessId: number) => {
    getListLoading.value = true;
    const { vendor } = condition;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', `vendors/${vendor}/targets/by_topo/list`, businessId);
    const rs = resolveApiPathByBusinessId('/api/v1/cloud', `vendors/${vendor}/targets/by_topo/count`, businessId);
    try {
      const [listRes, countRes, rsCountRes] = await Promise.all<
        [Promise<IListResData<IRsItem[]>>, Promise<IListResData<IRsItem[]>>, Promise<IListResData<IRsItem[]>>]
      >([
        http.post(api, enableCount({ ...condition, page }, false)),
        http.post(api, enableCount({ ...condition, page }, true)),
        http.post(rs, enableCount(condition, true)),
      ]);

      const list = listRes?.data?.details ?? [];
      const count = countRes?.data?.count ?? 0;
      const rsCount = rsCountRes?.data?.count ?? 0;

      return { list, count, rsCount };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      getListLoading.value = false;
    }
  };

  const batchUpdateWeightLoading = ref(false);
  // 单个/批量修改RS权重
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

  const batchUpdatePortLoading = ref(false);
  // 单个/批量修改RS端口
  const batchUpdatePort = async (
    target_group_id: string,
    params: { target_ids: any[]; new_port: number },
    businessId: number,
  ) => {
    batchUpdatePortLoading.value = true;
    const api = resolveApiPathByBusinessId(
      '/api/v1/cloud',
      `target_groups/${target_group_id}/targets/port`,
      businessId,
    );
    try {
      const res = await http.patch(api, params);
      return res;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      batchUpdatePortLoading.value = false;
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
    batchUpdatePortLoading,
    batchUpdatePort,
    batchUnbindLoading,
    batchUnbind,
    batchExportLoading,
    batchExport,
  };
});
