import { ref } from 'vue';
import { defineStore } from 'pinia';
import { resolveApiPathByBusinessId } from '@/common/util';
import http from '@/http';
import type { IListResData } from '@/typings';
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

  const rsList = ref<IRsItem[]>();
  const rsListCount = ref(0);
  const rsCount = ref(0);

  // 获取设备检索-RS列表
  const getRsList = async (condition: ILoadBalanceDeviceCondition, businessId: number) => {
    getListLoading.value = true;
    const { vendor } = condition;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', `vendors/${vendor}/targets/by_topo/list`, businessId);
    const rs = resolveApiPathByBusinessId('/api/v1/cloud', `vendors/${vendor}/targets/by_topo/count`, businessId);
    try {
      const [rsRes, rsCountRes] = await Promise.all<
        [Promise<IListResData<IRsItem[]>>, Promise<IListResData<IRsItem[]>>]
      >([http.post(api, condition, { globalError: false }), http.post(rs, condition, { globalError: false })]);

      rsList.value = rsRes?.data?.details ?? [];
      rsListCount.value = rsRes?.data?.count ?? 0;
      rsCount.value = rsCountRes?.data?.count ?? 0;

      return { list: rsList.value, count: rsListCount.value, rsCount: rsCount.value };
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
