import { ref } from 'vue';
import { defineStore } from 'pinia';
import { enableCount } from '@/utils/search';
import { resolveApiPathByBusinessId } from '@/common/util';
import type { IListResData, QueryBuilderType, IPageQuery } from '@/typings';
import http from '@/http';
import { ListenerProtocol, Scheduler, SessionType, SSLMode } from '@/views/load-balancer/constants';
import { VendorEnum } from '@/common/constant';
import rollRequest from '@blueking/roll-request';
import { ILoadBalanceDeviceCondition } from '@/views/load-balancer/device/typing';

export interface IListenerModel {
  id: string;
  account_id: string;
  lb_id: string;
  name: string;
  protocol: ListenerProtocol;
  port: number;
  scheduler: Scheduler;
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
  targets?: any[];
  cloud_id: string;
  vendor: VendorEnum;
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
  // 异步加载字段
  rs_num: number;
  non_zero_weight_count: number;
  zero_weight_count: number;
  non_zero_weight_target_count?: number;
  target_count?: number;
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

export interface IListenerDomainInfoItem {
  domain: string;
  url_count: number;
  // 前端交互字段
  key?: string;
  displayConfig?: { isNew?: boolean; originDomain?: string; isExpand?: boolean };
}
export interface IListenerDomainsListResponseData {
  default_domain: string;
  domain_list: IListenerDomainInfoItem[];
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
  scheduler: IListenerModel['scheduler'];
  session_type: IListenerModel['session_type'];
  session_expire: number;
  health_check: IListenerItem['health_check'];
  certificate: IListenerModel['certificate'];
  memo: string;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  // 异步加载字段
  rs_num: number;
  binding_status: string;
  // 前端交互字段
  displayConfig?: {
    isNew?: boolean;
  };
}

export interface IListenerRuleModel {
  id?: string;
  url: string;
  target_group_id?: string;
  domains?: string[]; // create
  domain?: string; // update
  session_expire_time?: number;
  scheduler?: Scheduler;
  forward_type?: string;
  default_server?: boolean;
  http2?: boolean;
  target_type?: string;
  quic?: boolean;
  trpc_func?: string;
  trpc_callee?: string;
  certificate?: IListenerModel['certificate']; // https必传
}

export interface IListenerRuleCreateResponseData {
  unknown_cloud_ids: string[];
  success_cloud_ids: string[];
  failed_cloud_ids: string[];
  success_ids: string[];
  failed_message: string;
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

  // 设备检索监听器列表
  const deviceListenerListLoading = ref(false);
  const getDeviceListenerList = async (
    condition: ILoadBalanceDeviceCondition,
    page: IPageQuery,
    businessId: number,
    loadAll = false,
  ) => {
    const { vendor } = condition;
    deviceListenerListLoading.value = true;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', `vendors/${vendor}/listeners/by_topo/list`, businessId);
    try {
      if (loadAll) {
        const list = (await rollRequest({ httpClient: http, pageEnableCountKey: 'count' }).rollReqUseCount(
          api,
          condition as any,
          { limit: 500, countGetter: (res) => res.data.count, listGetter: (res) => res.data.details },
        )) as any[];

        const data = Object.assign({ list: [], count: 0 }, { list, count: list.length });
        return data;
      }

      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<IListenerRuleItem[]>>, Promise<IListResData<IListenerRuleItem[]>>]
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
      deviceListenerListLoading.value = false;
    }
  };

  const batchDeleteListenerLoading = ref(false);
  const batchDeleteListener = async (data: { ids: string[]; account_id: string }, businessId?: number) => {
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

  const domainListLoading = ref(false);
  const getDomainListByListenerId = async (vendor: VendorEnum, listenerId: string, businessId?: number) => {
    domainListLoading.value = true;
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
      domainListLoading.value = false;
    }
  };

  const updateDomainLoading = ref(false);
  const updateDomain = async (
    listenerId: string,
    payload: {
      domain: string;
      new_domain?: string;
      certificate?: IListenerModel['certificate'];
      default_server?: boolean;
    },
    businessId?: number,
  ) => {
    updateDomainLoading.value = true;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', `listeners/${listenerId}/domains`, businessId);
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

  const batchDeleteDomainLoading = ref(false);
  const batchDeleteDomain = async (
    vendor: VendorEnum,
    listenerId: string,
    payload: { domains: string[]; new_default_domain?: string },
    businessId?: number,
  ) => {
    batchDeleteDomainLoading.value = true;
    const api = resolveApiPathByBusinessId(
      '/api/v1/cloud',
      `vendors/${vendor}/listeners/${listenerId}/rules/by/domains/batch`,
      businessId,
    );
    try {
      const res = await http.delete(api, { data: payload });
      return res;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      batchDeleteDomainLoading.value = false;
    }
  };

  const ruleListLoading = ref(false);
  const getRuleListByListenerId = async (
    vendor: VendorEnum,
    listenerId: string,
    payload: QueryBuilderType,
    businessId?: number,
  ) => {
    ruleListLoading.value = true;
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
      ruleListLoading.value = false;
    }
  };

  const rulesBindingStatusListLoading = ref(false);
  const getRulesBindingStatusList = async (
    vendor: VendorEnum,
    listenerId: string,
    payload: { rule_ids: string[] },
    businessId?: number,
  ) => {
    rulesBindingStatusListLoading.value = true;
    const api = resolveApiPathByBusinessId(
      '/api/v1/cloud',
      `vendors/${vendor}/listeners/${listenerId}/rules/binding_status/list`,
      businessId,
    );
    try {
      const res: IListResData<{ rule_id: string; binding_status: string }[]> = await http.post(api, payload);
      return res.data?.details || [];
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      rulesBindingStatusListLoading.value = false;
    }
  };

  const createRulesLoading = ref(false);
  const createRules = async (
    vendor: VendorEnum,
    listenerId: string,
    payload: IListenerRuleModel,
    businessId?: number,
  ) => {
    createRulesLoading.value = true;
    const api = resolveApiPathByBusinessId(
      '/api/v1/cloud',
      `vendors/${vendor}/listeners/${listenerId}/rules/create`,
      businessId,
    );
    try {
      const res = await http.post(api, payload);
      return res.data as IListenerRuleCreateResponseData;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      createRulesLoading.value = false;
    }
  };

  const updateRuleLoading = ref(false);
  const updateRule = async (
    vendor: VendorEnum,
    listenerId: string,
    ruleId: string,
    payload: IListenerRuleModel,
    businessId?: number,
  ) => {
    updateRuleLoading.value = true;
    const api = resolveApiPathByBusinessId(
      '/api/v1/cloud',
      `vendors/${vendor}/listeners/${listenerId}/rules/${ruleId}`,
      businessId,
    );
    try {
      const res = await http.patch(api, payload);
      return res;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      updateRuleLoading.value = false;
    }
  };

  const batchDeleteRuleLoading = ref(false);
  const batchDeleteRule = async (
    vendor: VendorEnum,
    listenerId: string,
    payload: { rule_ids: string[] },
    businessId?: number,
  ) => {
    batchDeleteRuleLoading.value = true;
    const api = resolveApiPathByBusinessId(
      '/api/v1/cloud',
      `vendors/${vendor}/listeners/${listenerId}/rules/batch`,
      businessId,
    );
    try {
      const res = await http.delete(api, { data: payload });
      return res;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      batchDeleteRuleLoading.value = false;
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
    domainListLoading,
    getDomainListByListenerId,
    updateDomainLoading,
    updateDomain,
    batchDeleteDomainLoading,
    batchDeleteDomain,
    ruleListLoading,
    getRuleListByListenerId,
    rulesBindingStatusListLoading,
    getRulesBindingStatusList,
    createRulesLoading,
    createRules,
    updateRuleLoading,
    updateRule,
    batchDeleteRuleLoading,
    batchDeleteRule,
    getDeviceListenerList,
    deviceListenerListLoading,
  };
});
