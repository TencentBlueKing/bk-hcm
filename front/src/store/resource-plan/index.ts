import { ref } from 'vue';
import { defineStore } from 'pinia';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import http from '@/http';
import { enableCount } from '@/utils/search';
import { IOverviewListResData, IPageQuery, IListResData } from '@/typings';

export enum ResourcesDemandStatus {
  CAN_APPLY = 'can_apply',
  NOT_READY = 'not_ready',
  EXPIRED = 'expired',
  SPENT_ALL = 'spent_all',
  LOCKED = 'locked',
}

export interface IResourcesDemandOverview {
  total_cpu_core: number;
  total_applied_core: number;
  in_plan_cpu_core: number;
  in_plan_applied_cpu_core: number;
  out_plan_cpu_core: number;
  out_plan_applied_cpu_core: number;
  expiring_cpu_core: number;
}

export interface IResourcesDemandItem {
  demand_id: string;
  bk_biz_id: number;
  bk_biz_name: string;
  op_product_id: number;
  op_product_name: string;
  status: ResourcesDemandStatus;
  status_name: string;
  demand_class: string;
  available_year_month: string;
  expect_time: string;
  device_class: string;
  device_type: string;
  total_os: string;
  applied_os: string;
  remained_os: string;
  total_cpu_core: number;
  applied_cpu_core: number;
  remained_cpu_core: number;
  total_memory: number;
  applied_memory: number;
  remained_memory: number;
  total_disk_size: number;
  applied_disk_size: number;
  remained_disk_size: number;
  region_id: string;
  region_name: string;
  zone_id: string;
  zone_name: string;
  plan_type: string;
  obs_project: string;
  generation_type: string;
  device_family: string;
  core_type: string;
  disk_type: string;
  disk_type_name: string;
  disk_io: number;
}

export interface IListResourcesDemandsParams {
  bk_biz_ids?: number[];
  op_product_ids?: string[];
  plan_product_ids?: string[];
  demand_ids?: string[];
  obs_projects?: string[];
  demand_classes?: string[];
  device_classes?: string[];
  device_types?: string[];
  region_ids?: string[];
  zone_ids?: string[];
  plan_types?: string[];
  expiring_only?: boolean;
  expect_time_range?: {
    start: string;
    end: string;
  };
  statuses?: ResourcesDemandStatus[];
  page: IPageQuery;
}

export interface IResourcePlanOpProductItem {
  op_product_id: string;
  op_product_name: string;
}

export interface IResourcePlanPlanProductItem {
  plan_product_id: string;
  plan_product_name: string;
}

export const useResourcePlanStore = defineStore('resource-plan', () => {
  const { getBusinessApiPath } = useWhereAmI();

  const opProductList = ref<IResourcePlanOpProductItem[]>();
  const planProductList = ref<IResourcePlanPlanProductItem[]>();

  const demandListLoading = ref(false);
  const getDemandList = async (params: IListResourcesDemandsParams) => {
    demandListLoading.value = true;
    const api = `/api/v1/woa/${getBusinessApiPath()}plans/resources/demands/list`;
    try {
      const [listRes, countRes] = await Promise.all<
        [
          Promise<IOverviewListResData<IResourcesDemandItem[], IResourcesDemandOverview>>,
          Promise<IOverviewListResData<IResourcesDemandItem[], IResourcesDemandOverview>>,
        ]
      >([http.post(api, enableCount(params, false)), http.post(api, enableCount(params, true))]);
      const [{ details: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      demandListLoading.value = false;
    }
  };

  const getOpProductList = async () => {
    if (opProductList.value) {
      return opProductList.value;
    }
    try {
      const res: IListResData<IResourcePlanOpProductItem[]> = await http.post('/api/v1/woa/metas/op_products/list');
      opProductList.value = res.data.details ?? [];
      return opProductList.value;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  const getPlanProductList = async () => {
    if (planProductList.value) {
      return planProductList.value;
    }
    try {
      const res: IListResData<IResourcePlanPlanProductItem[]> = await http.post('/api/v1/woa/metas/plan_products/list');
      planProductList.value = res.data.details ?? [];
      return planProductList.value;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  const getBizOrgRelation = async (bizId: number) => {
    return http.get(`/api/v1/woa/bizs/${bizId}/org/relation`);
  };

  const getBizsByOpProductList = async (params: { op_product_id: number }) => {
    const res = await http.post('/api/v1/woa/metas/bizs/by/op_product/list', params);
    return (res?.data?.details ?? []) as Array<{ bk_biz_id: number; bk_biz_name: string }>;
  };

  return {
    demandListLoading,
    getDemandList,
    getOpProductList,
    getPlanProductList,
    getBizOrgRelation,
    getBizsByOpProductList,
  };
});
