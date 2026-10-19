import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { enableCount } from '@/utils/search';
import type { IListResData, QueryBuilderType } from '@/typings';
import type { VendorEnum } from '@/common/constant';

const EXCLUSIVE_CLUSTER_API_PATH = '/api/v1/cloud/load_balancers/exclusive_clusters';

export enum ExclusiveClusterType {
  TGW = 'TGW',
  STGW = 'STGW',
}

export interface IExclusiveClusterItem {
  id: string;
  cloud_id: string;
  name: string;
  vendor: VendorEnum;
  account_id: string;
  bk_biz_id: number;
  region: string;
  zone: string;
  cluster_type: ExclusiveClusterType;
  cluster_tag: string;
  network: string;
  isp: string;
  egress: string;
  ip_version: string;
  max_conn: number | null;
  clb_resource_count: number;
  extension: Record<string, any>;
  memo: string;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
}

export const useLoadBalancerExclusiveClusterStore = defineStore('load-balancer-exclusive-cluster', () => {
  const exclusiveClusterListLoading = ref(false);
  const getExclusiveClusterList = async (payload: QueryBuilderType) => {
    exclusiveClusterListLoading.value = true;

    const api = `${EXCLUSIVE_CLUSTER_API_PATH}/list`;
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<IExclusiveClusterItem[]>>, Promise<IListResData<IExclusiveClusterItem[]>>]
      >([http.post(api, enableCount(payload, false)), http.post(api, enableCount(payload, true))]);

      return { list: listRes?.data?.details ?? [], count: countRes?.data?.count ?? 0 };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      exclusiveClusterListLoading.value = false;
    }
  };

  const assignExclusiveClusterToBizLoading = ref(false);
  // 仅未分配（bk_biz_id 为 -1）的集群可分配，含已分配集群时后端整批拒绝
  const assignExclusiveClusterToBiz = async (clusterIds: string[], bkBizId: number) => {
    assignExclusiveClusterToBizLoading.value = true;
    try {
      return await http.post(`${EXCLUSIVE_CLUSTER_API_PATH}/assign/bizs`, {
        cluster_ids: clusterIds,
        bk_biz_id: bkBizId,
      });
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      assignExclusiveClusterToBizLoading.value = false;
    }
  };

  return {
    exclusiveClusterListLoading,
    getExclusiveClusterList,
    assignExclusiveClusterToBizLoading,
    assignExclusiveClusterToBiz,
  };
});
