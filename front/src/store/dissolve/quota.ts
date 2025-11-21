import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { IQueryResData } from '@/typings';
import { resolveBizApiPath } from '@/utils/search';

export interface ICpuCoreSummary {
  total_core: number;
  delivered_core: number;
}

export const useDissolveQuotaStore = defineStore('dissolve-quota', () => {
  const cpuCoreSummaryLoading = ref(false);
  const dissolveConfigLoading = ref(false);
  const upsertDissolveConfigLoading = ref(false);

  const getCpuCoreSummary = async (bizId: number, params: { bk_biz_id?: number } = {}) => {
    cpuCoreSummaryLoading.value = true;
    try {
      const api = `/api/v1/woa/${resolveBizApiPath(bizId)}dissolve/cpu_core/summary`;
      const res: IQueryResData<ICpuCoreSummary> = await http.post(api, params);
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      cpuCoreSummaryLoading.value = false;
    }
  };

  const getDissolveConfig = async () => {
    dissolveConfigLoading.value = true;
    try {
      const api = '/api/v1/woa/dissolve/config';
      const res: IQueryResData<{ host_apply_time: string; approval_limit: string }> = await http.get(api);
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      dissolveConfigLoading.value = false;
    }
  };

  const upsertDissolveConfig = async (params: { host_apply_time?: string; approval_limit?: string }) => {
    upsertDissolveConfigLoading.value = true;
    try {
      const api = '/api/v1/woa/dissolve/config/upsert';
      const res: IQueryResData<null> = await http.put(api, params);
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      upsertDissolveConfigLoading.value = false;
    }
  };

  return {
    cpuCoreSummaryLoading,
    dissolveConfigLoading,
    upsertDissolveConfigLoading,
    getCpuCoreSummary,
    getDissolveConfig,
    upsertDissolveConfig,
  };
});
