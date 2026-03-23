import { ref } from 'vue';
import { defineStore } from 'pinia';
import rollRequest from '@blueking/roll-request';
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

/** 创建 GPU 需求主单 - 子单明细 */
export interface IGpuDemandCreateDetail {
  demand_type: string;
  demand_year: number;
  demand_month: number;
  gpu_num: number;
  qpm_max: number;
  extension: any[];
}

/** 创建 GPU 需求主单 - 请求参数 */
export interface IGpuDemandCreateParams {
  op_product_id: number;
  op_product_name: string;
  details: IGpuDemandCreateDetail[];
}

/** GPU 需求子单数据 */
export interface IGpuDemandSubOrder {
  id: string;
  order_id: string;
  bk_biz_id: number;
  op_product_id: number;
  op_product_name: string;
  demand_type: string;
  demand_year: number;
  demand_month: number;
  gpu_num: number;
  qpm_max: number;
  status: string;
  comment: string[];
  extension: any[];
  remark: string;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
}

/** 子单列表接口 tpl_config 中的 header 定义 */
export interface ITplHeader {
  name: string;
  field: string;
  db_field?: string;
  type: string;
  formula?: string;
  required?: boolean;
  readonly?: boolean;
  hidden?: boolean;
  value?: (string | number)[];
  /** 大于 (exclusive minimum) */
  gt?: number;
  /** 大于等于 (inclusive minimum) */
  gte?: number;
  /** 小于 (exclusive maximum) */
  lt?: number;
  /** 小于等于 (inclusive maximum) */
  lte?: number;
}

/** 子单列表接口 tpl_config 中的 sheet 定义 */
export interface ITplSheet {
  name: string;
  head_row: number;
  row_start: number;
  fixed_headers: ITplHeader[];
  headers: ITplHeader[];
}

/** 子单列表接口 tpl_config */
export interface ITplConfig {
  id: string;
  sheets: ITplSheet[];
}

/** 子单批量更新项 */
export interface IGpuSubOrderUpdateItem {
  suborder_id: string;
  status?: string;
  comment?: string[];
  demand_year?: number;
  demand_month?: number;
  gpu_num?: number;
  qpm_max?: number;
  extension?: any[];
}

/** 子单列表查询参数 */
export interface IGpuSubOrderListParams {
  filter: QueryFilterType | QueryFilterTypeLegacy;
}

/** 子单列表接口返回的 data 结构 */
export interface IGpuSubOrderListData {
  count: number;
  details: IGpuDemandSubOrder[];
  tpl_config: ITplConfig[];
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

  /** 创建 GPU 需求提报主单（同时批量创建子单） */
  const createGpuDemandOrder = async (params: IGpuDemandCreateParams) => {
    const { getBizsId } = useWhereAmI();
    const bizId = getBizsId();
    const api = `/api/v1/woa/bizs/${bizId}/plans/resources/gpu/order/create`;
    return http.post(api, params);
  };

  /** 覆盖上传（重新导入）GPU 需求主单 */
  const overwriteGpuDemandOrder = async (params: { order_id: string; details: IGpuDemandCreateDetail[] }) => {
    const { getBizsId } = useWhereAmI();
    const bizId = getBizsId();
    const api = `/api/v1/woa/bizs/${bizId}/plans/resources/gpu/order/overwrite`;
    return http.patch(api, params);
  };

  const detailLoading = ref(false);

  /** 获取 GPU 需求子单列表（含 tpl_config），使用 rollRequest 滚动拉取全量数据 */
  const getGpuSubOrderList = async (params: IGpuSubOrderListParams): Promise<IGpuSubOrderListData> => {
    detailLoading.value = true;
    const api = `/api/v1/woa/${getBusinessApiPath()}plans/resources/gpu/demands/suborders/list`;
    let capturedTplConfig: ITplConfig[] = [];
    try {
      const details = (await rollRequest({
        httpClient: http,
        pageEnableCountKey: 'count',
      }).rollReqUseCount<IGpuDemandSubOrder>(api, params, {
        limit: 500,
        countGetter: (res) => res.data.count,
        listGetter: (res) => {
          if (res.data.tpl_config?.length && !capturedTplConfig.length) {
            capturedTplConfig = res.data.tpl_config;
          }
          return res.data.details;
        },
      })) as IGpuDemandSubOrder[];
      return { count: details.length, details, tpl_config: capturedTplConfig };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      detailLoading.value = false;
    }
  };

  /** 获取需求主单详情（复用列表接口通过 ID 过滤） */
  const getGpuDemandDetail = async (orderId: string): Promise<IGpuDemandItem | null> => {
    const api = `/api/v1/woa/${getBusinessApiPath()}plans/resources/gpu/demands/orders/list`;
    try {
      const res = await http.post<IListResData<IGpuDemandItem[]>>(api, {
        filter: { op: 'and', rules: [{ field: 'id', op: 'eq', value: orderId }] },
        page: { count: false, start: 0, limit: 1 },
      });
      const list = res?.data?.details ?? [];
      return list.length > 0 ? list[0] : null;
    } catch (error) {
      console.error(error);
      return null;
    }
  };

  /** 业务下 GPU 需求子单批量终止 */
  const batchTerminateSubOrders = async (params: { suborder_ids: string[] }) => {
    const { getBizsId } = useWhereAmI();
    const bizId = getBizsId();
    const api = `/api/v1/woa/bizs/${bizId}/plans/resources/gpu/demands/suborders/batch/terminate`;
    return http.post(api, params);
  };

  /** 服务请求 - 子单批量更新（评审/驳回/编辑） */
  const batchUpdateSubOrders = async (params: { suborder_data: IGpuSubOrderUpdateItem[] }) => {
    const api = `/api/v1/woa/${getBusinessApiPath()}plans/resources/gpu/demands/suborders/batch`;
    return http.post(api, params);
  };

  /** 资源下 GPU 需求子单批量更新状态（评审通过/驳回） */
  const batchUpdateSubOrderStatus = async (params: { suborder_ids: string[]; status: string; comment?: string[] }) => {
    const api = '/api/v1/woa/plans/resources/gpu/demands/suborders/batch/status';
    return http.post(api, params);
  };

  return {
    listLoading,
    detailLoading,
    getGpuDemandList,
    batchPendingOrders,
    batchRejectOrders,
    batchTerminateOrders,
    batchUpdateSubOrders,
    batchUpdateSubOrderStatus,
    batchTerminateSubOrders,
    createGpuDemandOrder,
    overwriteGpuDemandOrder,
    getGpuSubOrderList,
    getGpuDemandDetail,
  };
});
