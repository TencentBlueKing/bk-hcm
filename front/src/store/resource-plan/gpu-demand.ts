import { ref } from 'vue';
import { defineStore } from 'pinia';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import http from '@/http';
import { enableCount } from '@/utils/search';
import type { IListResData, QueryFilterType, QueryFilterTypeLegacy, IPageQuery } from '@/typings';

export const GPU_DEMAND_STATUS = {
  INIT: 'INIT',
  PENDING: 'PENDING',
  DONE: 'DONE',
  REJECT: 'REJECT',
  REJECT_ALL: 'REJECT_ALL',
  TERMINATE: 'TERMINATE',
} as const;

export type GpuDemandStatus = (typeof GPU_DEMAND_STATUS)[keyof typeof GPU_DEMAND_STATUS];

export const GPU_DEMAND_STATUS_MAP: Record<GpuDemandStatus, string> = {
  [GPU_DEMAND_STATUS.INIT]: '待评审',
  [GPU_DEMAND_STATUS.PENDING]: '评审中',
  [GPU_DEMAND_STATUS.DONE]: '已评审',
  [GPU_DEMAND_STATUS.REJECT]: '部分已驳回',
  [GPU_DEMAND_STATUS.REJECT_ALL]: '全部已驳回',
  [GPU_DEMAND_STATUS.TERMINATE]: '已终止',
};

export interface IGpuDemandItem {
  id: string;
  bk_biz_id: number;
  op_product_id: number;
  op_product_name: string;
  template_id: string;
  status: GpuDemandStatus;
  remark: string;
  total_gpu_num: number;
  total_qpm_max: number;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
}

export interface IGpuDemandListParams {
  filter: QueryFilterType | QueryFilterTypeLegacy;
  page: IPageQuery;
}

export const useGpuDemandStore = defineStore('gpu-demand', () => {
  const { getBusinessApiPath } = useWhereAmI();

  const listLoading = ref(false);

  const getGpuDemandList = async (params: IGpuDemandListParams) => {
    listLoading.value = true;
    const api = `/api/v1/woa/${getBusinessApiPath()}plans/resources/gpu/demands/orders/list`;
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<IGpuDemandItem[]>>, Promise<IListResData<IGpuDemandItem[]>>]
      >([http.post(api, enableCount(params, false)), http.post(api, enableCount(params, true))]);
      const [{ details: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      listLoading.value = false;
    }
  };

  const batchPendingOrders = async (params: { order_ids: string[] }) => {
    const api = '/api/v1/woa/plans/resources/gpu/demands/orders/batch/pending';
    return http.post(api, params);
  };

  const batchRejectOrders = async (params: { order_ids: string[] }) => {
    const api = '/api/v1/woa/plans/resources/gpu/demands/orders/batch/reject';
    return http.post(api, params);
  };

  const batchTerminateOrders = async (params: { order_ids: string[] }) => {
    const api = `/api/v1/woa/${getBusinessApiPath()}plans/resources/gpu/demands/orders/batch/terminate`;
    return http.post(api, params);
  };

  return {
    listLoading,
    getGpuDemandList,
    batchPendingOrders,
    batchRejectOrders,
    batchTerminateOrders,
  };
});
