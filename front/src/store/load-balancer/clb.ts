import { ref } from 'vue';
import { defineStore } from 'pinia';
import type { IListResData, QueryBuilderType } from '@/typings';
import http from '@/http';
import rollRequest from '@blueking/roll-request';
import { enableCount } from '@/utils/search';
import { resolveApiPathByBusinessId } from '@/common/util';
import { LoadBalancerBatchImportPreviewDetails } from '@/views/load-balancer/clb/children/batch-import/typings';
import { VendorEnum } from '@/common/constant';
import { LoadBalancerBatchImportOperationType } from '@/views/load-balancer/constants';

type Tags = Record<string, any>;

export interface ILoadBalancerWithDeleteProtectionItem {
  id: string;
  cloud_id: string;
  name: string;
  vendor: VendorEnum;
  account_id: string;
  bk_biz_id: number;
  ip_version: string;
  lb_type: string;
  region: string;
  zones: string[];
  backup_zones: string[];
  vpc_id: string;
  cloud_vpc_id: string;
  subnet_id: string;
  cloud_subnet_id: string;
  private_ipv4_addresses: string[];
  private_ipv6_addresses: string[];
  public_ipv4_addresses: string[];
  public_ipv6_addresses: string[];
  domain: string;
  status: string;
  bandwidth: number;
  isp: string;
  cloud_created_time: string;
  cloud_status_time: string;
  cloud_expired_time: string;
  tags: Tags;
  memo: string;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  delete_protect: boolean;
  listener_count?: number; // 异步加载
  [key: string]: any;
}

export interface ILoadBalancerDetails {
  id: string;
  cloud_id: string;
  name: string;
  vendor: VendorEnum;
  account_id: string;
  bk_biz_id: number;
  ip_version: string;
  lb_type: string;
  region: string;
  zones: string[];
  backup_zones: string[];
  vpc_id: string;
  cloud_vpc_id: string;
  subnet_id: string;
  cloud_subnet_id: string;
  private_ipv4_addresses: string[];
  private_ipv6_addresses: string[];
  public_ipv4_addresses: string[];
  public_ipv6_addresses: string[];
  domain: string;
  status: string;
  cloud_created_time: string;
  cloud_status_time: string;
  cloud_expired_time: string;
  tags: Record<string, any>;
  memo: null;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  extension: {
    sla_type: string;
    charge_type: string;
    load_balancer_pass_to_target: boolean;
    internet_charge_type: string;
    snat: boolean;
    snat_pro: boolean;
    snat_ips: string[];
    delete_protect: boolean;
    egress: string;
    mix_ip_target: boolean;
    forward: number;
    target_vpc: string;
    target_region: string;
  };
  isp: string;
  bandwidth: number;
  sync_time: string;
}

export interface ILoadBalancerLockStatus {
  res_id: string;
  res_type: 'load_balancer';
  flow_id: string;
  status: 'executing' | 'success';
}

export const useLoadBalancerClbStore = defineStore('load-balancer-clb', () => {
  const loadBalancerListWithDeleteProtectionLoading = ref(false);
  const getLoadBalancerListWithDeleteProtection = async (
    payload: QueryBuilderType,
    businessId: number,
    useRollRequest = true,
  ) => {
    loadBalancerListWithDeleteProtectionLoading.value = true;

    const api = resolveApiPathByBusinessId('/api/v1/cloud', 'load_balancers/with/delete_protection/list', businessId);
    try {
      const result = { list: [] as ILoadBalancerWithDeleteProtectionItem[], count: 0 };

      if (useRollRequest) {
        const list = (await rollRequest({ httpClient: http, pageEnableCountKey: 'count' }).rollReqUseCount(
          api,
          payload as any,
          { limit: 500, countGetter: (res) => res.data.count, listGetter: (res) => res.data.details },
        )) as ILoadBalancerWithDeleteProtectionItem[];

        Object.assign(result, { list, count: list.length });
      } else {
        const [listRes, countRes] = await Promise.all<
          [
            Promise<IListResData<ILoadBalancerWithDeleteProtectionItem[]>>,
            Promise<IListResData<ILoadBalancerWithDeleteProtectionItem[]>>,
          ]
        >([http.post(api, enableCount(payload, false)), http.post(api, enableCount(payload, true))]);

        const list = listRes?.data?.details ?? [];
        const count = countRes?.data?.count ?? 0;

        Object.assign(result, { list, count });
      }

      return result;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      loadBalancerListWithDeleteProtectionLoading.value = false;
    }
  };

  const loadBalancerDetailsLoading = ref(false);
  const getLoadBalancerDetails = async (id: string, businessId?: number) => {
    loadBalancerDetailsLoading.value = true;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', `load_balancers/${id}`, businessId);
    try {
      const res = await http.get(api);
      return res.data as ILoadBalancerDetails;
    } catch (error) {
    } finally {
      loadBalancerDetailsLoading.value = false;
    }
  };

  const updateLoadBalancerLoading = ref(false);
  const updateLoadBalancer = async (
    vendor: VendorEnum,
    payload: {
      id: string;
      name?: string;
      internet_charge_type?: string;
      internet_max_bandwidth_out?: number;
      delete_protect?: boolean;
      load_balancer_pass_to_target?: boolean;
      snat_pro?: boolean;
      target_region?: string;
      target_vpc?: string;
      memo?: string;
    },
    businessId?: number,
  ) => {
    updateLoadBalancerLoading.value = true;
    const api = resolveApiPathByBusinessId(
      '/api/v1/cloud',
      `vendors/${vendor}/load_balancers/${payload.id}`,
      businessId,
    );
    try {
      const res = await http.patch(api, payload);
      return res;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      updateLoadBalancerLoading.value = false;
    }
  };

  const batchImportLoadBalancerLoading = ref(false);
  const batchImportLoadBalancer = async (
    vendor: VendorEnum,
    operation_type: LoadBalancerBatchImportOperationType,
    payload: {
      account_id: string;
      region_ids: string[];
      source: string;
      details: LoadBalancerBatchImportPreviewDetails;
    },
    businessId?: number,
  ) => {
    batchImportLoadBalancerLoading.value = true;
    const api = resolveApiPathByBusinessId(
      '/api/v1/cloud',
      `vendors/${vendor}/load_balancers/operations/${operation_type}/submit`,
      businessId,
    );
    try {
      const res = await http.post(api, payload);
      return res;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      batchImportLoadBalancerLoading.value = false;
    }
  };

  const batchDeleteLoadBalancerLoading = ref(false);
  const batchDeleteLoadBalancer = async (data: { ids: string[] }, businessId?: number) => {
    batchDeleteLoadBalancerLoading.value = true;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', 'load_balancers/batch', businessId);
    try {
      const res = await http.delete(api, { data });
      return res;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      batchDeleteLoadBalancerLoading.value = false;
    }
  };

  const loadBalancerListenerCountLoading = ref(false);
  const getListenerCountByLoadBalancerIds = async (lbIds: string[], businessId?: number) => {
    loadBalancerListenerCountLoading.value = true;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', `load_balancers/listeners/count`, businessId);
    try {
      const res = await http.post(api, { lb_ids: lbIds });
      return (res?.data?.details ?? []) as Array<{ lb_id: string; num: number }>;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      loadBalancerListenerCountLoading.value = false;
    }
  };

  const getLoadBalancerLockStatus = async (id: string, businessId?: number) => {
    const api = resolveApiPathByBusinessId('/api/v1/cloud', `load_balancers/${id}/lock/status`, businessId);
    try {
      const res = await http.get(api);
      return (res.data ?? {}) as ILoadBalancerLockStatus;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  return {
    loadBalancerListWithDeleteProtectionLoading,
    getLoadBalancerListWithDeleteProtection,
    loadBalancerDetailsLoading,
    getLoadBalancerDetails,
    updateLoadBalancerLoading,
    updateLoadBalancer,
    batchImportLoadBalancerLoading,
    batchImportLoadBalancer,
    batchDeleteLoadBalancerLoading,
    batchDeleteLoadBalancer,
    loadBalancerListenerCountLoading,
    getListenerCountByLoadBalancerIds,
    getLoadBalancerLockStatus,
  };
});
