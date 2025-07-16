import { ref } from 'vue';
import { defineStore } from 'pinia';
import { enableCount } from '@/utils/search';
import { resolveApiPathByBusinessId } from '@/common/util';
import type { IListResData, QueryBuilderType } from '@/typings';
import { ListenerProtocol, SessionType, SSLMode } from '@/views/load-balancer/constants';
import http from '@/http';
import { VendorEnum } from '@/common/constant';

export interface IListenerModel {
  id: string;
  account_id: string;
  lb_id: string;
  name: string;
  protocol: ListenerProtocol;
  port: number;
  scheduler: string;
  session_open?: boolean;
  session_type?: SessionType;
  session_expire?: number;
  target_group_id: string;
  domain?: string;
  url?: string;
  sni_switch: number;
  certificate: { ssl_mode: SSLMode; ca_cloud_id: string; cert_cloud_ids: string[] };
}

export interface IListenerItem extends IListenerModel {
  cloud_id: string;
  vendor: string;
  bk_biz_id: number;
  cloud_lb_id: string;
  default_domain: string;
  region: string;
  zones: string[];
  memo: string;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  end_port: number;
  domain_num: number;
  url_num: number;
  health_check: {
    health_switch: number;
    time_out: number;
    interval_time: number;
    health_num: number;
    un_health_num: number;
    http_code: number;
    check_type: string;
    http_check_path: string;
    http_check_domain: string;
    http_check_method: string;
    source_ip_type: number;
    extended_code: string;
  };
  binding_status: string;
}

export interface IListenersRsWeightStatItem {
  [key: string]: {
    non_zero_weight_count: number;
    zero_weight_count: number;
    total_count: number;
  };
}

export interface IListenerDetails extends IListenerItem {
  extension: {
    certificate: { ssl_mode: SSLMode; ca_cloud_id: string; cert_cloud_ids: string[] };
  };
  lbl_id: string;
  lbl_name: string;
  cloud_lbl_id: string;
  target_group_name: string;
  cloud_target_group_id: string;
}

export interface IListenerDomainsListResponseData {
  default_domain: string;
  domain_list: {
    domain: string;
    url_count: number;
  }[];
}

export interface IListenerRuleItem {
  id: string;
  cloud_id: string;
  name: string;
  rule_type: string;
  lb_id: string;
  cloud_lb_id: string;
  lbl_id: string;
  cloud_lbl_id: string;
  target_group_id: string;
  cloud_target_group_id: string;
  region: string;
  domain: string;
  url: string;
  scheduler: string;
  session_type: string;
  session_expire: number;
  health_check: IListenerItem['health_check'];
  certificate: IListenerModel['certificate'];
  memo: string;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
}

export const useLoadBalancerListenerStore = defineStore('load-balancer-listener', () => {
  const listenerListLoading = ref(false);
  const getListenerList = async (loadBalancerId: string, payload: QueryBuilderType, businessId: number) => {
    listenerListLoading.value = true;
    const api = resolveApiPathByBusinessId(
      '/api/v1/cloud',
      `load_balancers/${loadBalancerId}/listeners/list`,
      businessId,
    );
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<IListenerItem[]>>, Promise<IListResData<IListenerItem[]>>]
      >([http.post(api, enableCount(payload, false)), http.post(api, enableCount(payload, true))]);

      const list = listRes?.data?.details ?? [];
      const count = countRes?.data?.count ?? 0;

      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      listenerListLoading.value = false;
    }
  };

  const batchDeleteListenerLoading = ref(false);
  const batchDeleteListener = async (data: { ids: string[] }, businessId?: number) => {
    batchDeleteListenerLoading.value = true;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', 'listeners/batch', businessId);
    try {
      const res = await http.delete(api, { data });
      return res;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      batchDeleteListenerLoading.value = false;
    }
  };

  const listenerRsWeightStatLoading = ref(false);
  const getListenersRsWeightStat = async (ids: string[], businessId?: number) => {
    listenerRsWeightStatLoading.value = true;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', 'listeners/rs_weight_stat', businessId);
    try {
      const res = await http.post(api, { ids });
      return (res.data ?? {}) as IListenersRsWeightStatItem;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      listenerRsWeightStatLoading.value = false;
    }
  };

  const listenerDetailsLoading = ref(false);
  const getListenerDetails = async (listenerId: string, businessId?: number) => {
    listenerDetailsLoading.value = true;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', `listeners/${listenerId}`, businessId);
    try {
      const res = await http.get(api);
      return res.data as IListenerDetails;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      listenerDetailsLoading.value = false;
    }
  };

  const addListenerLoading = ref(false);
  const addListener = async (payload: IListenerModel, businessId?: number) => {
    addListenerLoading.value = true;
    const api = resolveApiPathByBusinessId(
      '/api/v1/cloud',
      `load_balancers/${payload.lb_id}/listeners/create`,
      businessId,
    );
    try {
      const res = await http.post(api, payload);
      return res;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      addListenerLoading.value = false;
    }
  };

  const updateListenerLoading = ref(false);
  const updateListener = async (data: Partial<IListenerDetails>, businessId?: number) => {
    updateListenerLoading.value = true;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', `listeners/${data.id}`, businessId);
    try {
      const res = await http.patch(api, data);
      return res;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      updateListenerLoading.value = false;
    }
  };

  const domainListByListenerIdLoading = ref(false);
  const getDomainListByListenerId = async (vendor: VendorEnum, listenerId: string, businessId?: number) => {
    domainListByListenerIdLoading.value = true;
    const api = resolveApiPathByBusinessId(
      '/api/v1/cloud',
      `vendors/${vendor}/listeners/${listenerId}/domains/list`,
      businessId,
    );
    try {
      const res = await http.post(api);
      return res.data as IListenerDomainsListResponseData;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      domainListByListenerIdLoading.value = false;
    }
  };

  const updateDomainLoading = ref(false);
  const updateDomain = async (
    payload: {
      lbl_id: string;
      domain: string;
      new_domain?: string;
      certificate?: IListenerModel['certificate'];
      default_server?: boolean;
    },
    businessId?: number,
  ) => {
    updateDomainLoading.value = true;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', `listeners/${payload.lbl_id}/domains`, businessId);
    try {
      const res = await http.patch(api, payload);
      return res;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      updateDomainLoading.value = false;
    }
  };

  const ruleListByListenerIdLoading = ref(false);
  const getRuleListByListenerId = async (
    vendor: VendorEnum,
    listenerId: string,
    payload: QueryBuilderType,
    businessId?: number,
  ) => {
    ruleListByListenerIdLoading.value = true;
    const api = resolveApiPathByBusinessId(
      '/api/v1/cloud',
      `vendors/${vendor}/listeners/${listenerId}/rules/list`,
      businessId,
    );
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<IListenerRuleItem[]>>, Promise<IListResData<IListenerRuleItem[]>>]
      >([http.post(api, enableCount(payload, false)), http.post(api, enableCount(payload, true))]);

      const list = listRes?.data?.details ?? [];
      const count = countRes?.data?.count ?? 0;

      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      ruleListByListenerIdLoading.value = false;
    }
  };

  return {
    listenerListLoading,
    getListenerList,
    batchDeleteListenerLoading,
    batchDeleteListener,
    listenerRsWeightStatLoading,
    getListenersRsWeightStat,
    listenerDetailsLoading,
    getListenerDetails,
    addListenerLoading,
    addListener,
    updateListenerLoading,
    updateListener,
    domainListByListenerIdLoading,
    getDomainListByListenerId,
    updateDomainLoading,
    updateDomain,
    ruleListByListenerIdLoading,
    getRuleListByListenerId,
  };
});
