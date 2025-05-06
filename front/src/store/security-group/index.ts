import { ref } from 'vue';
import { defineStore } from 'pinia';
import rollRequest from '@blueking/roll-request';
import http from '@/http';
import { VendorEnum } from '@/common/constant';
import { IListResData, IQueryResData, QueryBuilderType } from '@/typings';
import { enableCount } from '@/utils/search';
import { useWhereAmI } from '@/hooks/useWhereAmI';

export enum SecurityGroupManageType {
  BIZ = 'biz',
  PLATFORM = 'platform',
  UNKNOWN = '',
}

export enum SecurityGroupRelatedResourceName {
  CVM = 'CVM',
  CLB = 'CLB',
}

export type SecurityGroupMgmtAttrSingleType = 'manager' | 'bak_manager' | 'mgmt_biz_id' | 'usage_biz_ids';

export interface ISecurityGroupItem {
  id: string;
  vendor: VendorEnum;
  cloud_id: string;
  region: string;
  name: string;
  manager: string;
  bak_manager: string;
  usage_biz_ids: number[];
  mgmt_biz_id: number;
  mgmt_type: SecurityGroupManageType;
  memo: string;
  account_id: string;
  bk_biz_id: number;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  tags: Record<string, string>;
  cloud_created_time: string;
  cloud_update_time: string;
  account_managers?: string[];
  usage_biz_infos?: ISecurityGroupUsageBizMaintainerItem['usage_biz_infos'];
}

export interface ISecurityGroupRuleItem {
  id: string;
  type: string;
}

export interface ISecurityGroupDetail extends ISecurityGroupItem {
  cvm_count?: number; // tcloud、aws、huawei专属
  network_interface_count?: number; // azure专属
  subnet_count?: number; // azure专属
  extension?:
    | { cloud_project_id: string } // tcloud
    | {
        // aws
        vpc_id: string;
        cloud_vpc_id: string;
        cloud_owner_id: string;
      }
    | {
        // azure
        etag: string;
        flush_connection: string;
        resource_guid: string;
        provisioning_state: string;
        cloud_network_interface_ids: string[];
        cloud_subnet_ids: string[];
      }
    | {
        // huawei
        cloud_project_id: string;
        cloud_enterprise_project_id: string;
      };
  // ;
}

interface ISecurityGroupAssignPreviewItem {
  id: string;
  assignable: boolean;
  reason: string;
  assigned_biz_id: number;
}

type SecurityGroupRelType = 'cvm' | 'load_balancer' | string;
export interface IBizRelatedResource {
  bk_biz_id: number;
  res_count: number;
}
export type ISecurityGroupRelBusiness = Record<SecurityGroupRelType, IBizRelatedResource[]>;

type SecurityGroupRelCvmCommonFields = {
  id: string;
  cloud_id: string;
  name: string;
  vendor: VendorEnum;
  bk_biz_id: number;
  account_id: string;
  region: string;
  status: string;
  private_ipv4_addresses: string[];
  private_ipv6_addresses: string[];
  public_ipv4_addresses: string[];
  public_ipv6_addresses: string[];
};
export interface ISecurityGroupRelCvmByBizItem extends SecurityGroupRelCvmCommonFields {
  zone?: string;
  cloud_vpc_ids?: string[];
  cloud_subnet_ids?: string[];
}
export interface ISecurityGroupRelLoadBalancerByBizItem extends SecurityGroupRelCvmCommonFields {
  main_zones?: string[];
  backup_zones?: string[];
  cloud_vpc_id?: string;
  vpc_id?: string;
  network_type?: string;
  domain?: string;
  memo?: string;
}
export type SecurityGroupRelResourceByBizItem = (
  | ISecurityGroupRelCvmByBizItem
  | ISecurityGroupRelLoadBalancerByBizItem
) & {
  security_groups?: IResourceBoundSecurityGroupItem['security_groups'];
};

export interface ISecurityGroupRelResCountItem {
  id: string;
  resources?: Array<{
    res_name: SecurityGroupRelatedResourceName;
    count: number;
  }>;
  error?: string;
}

export type SecurityGroupRuleCountAndRelatedResourcesResult = {
  ruleCountMap: Record<string, number>;
  relatedResourcesList: ISecurityGroupRelResCountItem[];
};

// 安全组单个操作项的类型
export type ISecurityGroupOperateItem = ISecurityGroupItem &
  ISecurityGroupRelResCountItem & { rule_count?: number } & { [key: string]: any };

export interface IResourceBoundSecurityGroupItem {
  res_id: string;
  security_groups: { id: string; cloud_id: string; name: string }[];
}

export interface ISecurityGroupUsageBizMaintainerItem {
  id: string;
  managers: string[];
  usage_biz_infos: Array<{
    bk_biz_id: number;
    bk_biz_name: string;
    bk_biz_maintainer: string;
  }>;
}

export const useSecurityGroupStore = defineStore('security-group', () => {
  const { getBusinessApiPath } = useWhereAmI();

  const isFullListLoading = ref(false);
  const isFullRuleLoading = ref(false);

  const getFullList = async (params: QueryBuilderType) => {
    isFullListLoading.value = true;
    try {
      const list = await rollRequest({
        httpClient: http,
        pageEnableCountKey: 'count',
      }).rollReqUseCount<ISecurityGroupItem>(`/api/v1/cloud/${getBusinessApiPath()}security_groups/list`, params, {
        limit: 500,
        countGetter: (res) => res.data.count,
        listGetter: (res) => res.data.details,
      });

      return list;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isFullListLoading.value = false;
    }
  };

  const getFullRuleList = async (params: QueryBuilderType & { vendor: string; id: string }) => {
    const { vendor, id, ...data } = params;
    isFullRuleLoading.value = true;
    try {
      const list = await rollRequest({
        httpClient: http,
        pageEnableCountKey: 'count',
      }).rollReqUseCount<ISecurityGroupRuleItem>(
        `/api/v1/cloud/${getBusinessApiPath()}vendors/${vendor}/security_groups/${id}/rules/list`,
        data,
        {
          limit: 500,
          countGetter: (res) => res.data.count,
          listGetter: (res) => res.data.details,
        },
      );

      return list;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isFullRuleLoading.value = false;
    }
  };

  // 预览安全组分配到业务的结果，是否可分配
  const isAssignPreviewLoading = ref(false);
  const getAssignPreview = async (ids: string[]) => {
    isAssignPreviewLoading.value = true;
    try {
      const res: IQueryResData<ISecurityGroupAssignPreviewItem[]> = await http.post(
        '/api/v1/cloud/security_groups/assign/bizs/preview',
        { ids },
      );
      return res.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isAssignPreviewLoading.value = false;
    }
  };

  // 批量分配安全组到业务
  const isBatchAssignToBizLoading = ref(false);
  const batchAssignToBiz = async (ids: string[]) => {
    isBatchAssignToBizLoading.value = true;
    try {
      await http.post('/api/v1/cloud/security_groups/assign/bizs/batch', { ids });
    } finally {
      isBatchAssignToBizLoading.value = false;
    }
  };

  // 批量更新安全组管理属性，仅当所有管理属性均不存在时才允许编辑，所有管理属性都要提供
  // 注意：通过该接口更新的安全组会被默认设置为业务管理类型，不可再更改为平台管理类型
  const isBatchUpdateMgmtAttrLoading = ref(false);
  const batchUpdateMgmtAttr = async (
    security_groups: Array<{ id: string; manager: string; bak_manager: string; mgmt_biz_id: number }>,
  ) => {
    isBatchUpdateMgmtAttrLoading.value = true;
    try {
      await http.patch(`/api/v1/cloud/${getBusinessApiPath()}security_groups/mgmt_attrs/batch`, { security_groups });
    } finally {
      isBatchUpdateMgmtAttrLoading.value = false;
    }
  };

  // 查询安全组关联资源所属的业务列表，目前仅支持查询关联的CVM和CLB资源。
  // 返回的业务列表中，一定包含当前业务，且一定排在第一个
  const isQueryRelBusinessLoading = ref(false);
  const queryRelBusiness = async (security_group_id: string) => {
    isQueryRelBusinessLoading.value = true;
    try {
      const res: IQueryResData<ISecurityGroupRelBusiness> = await http.post(
        `/api/v1/cloud/${getBusinessApiPath()}security_groups/${security_group_id}/related_resources/bizs/list`,
        { security_group_id },
      );
      return res.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isQueryRelBusinessLoading.value = false;
    }
  };

  // 查询安全组关联的cvm列表，仅展示cvm摘要信息
  const isQueryRelCvmByBizLoading = ref(false);
  const queryRelCvmByBiz = async (sg_id: string, res_biz_id: number, payload: QueryBuilderType) => {
    isQueryRelCvmByBizLoading.value = true;
    const api = `/api/v1/cloud/${getBusinessApiPath()}security_groups/${sg_id}/related_resources/biz_resources/${res_biz_id}/cvms/list`;
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<ISecurityGroupRelCvmByBizItem[]>>, Promise<IListResData<ISecurityGroupRelCvmByBizItem[]>>]
      >([http.post(api, enableCount(payload, false)), http.post(api, enableCount(payload, true))]);
      const [{ details: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isQueryRelCvmByBizLoading.value = false;
    }
  };

  // 查询安全组关联的负载均衡列表，仅展示负载均衡摘要信息。
  const isQueryRelLoadBalancerByBizLoading = ref(false);
  const queryRelLoadBalancerByBiz = async (sg_id: string, res_biz_id: number, payload: QueryBuilderType) => {
    isQueryRelLoadBalancerByBizLoading.value = true;
    const api = `/api/v1/cloud/${getBusinessApiPath()}security_groups/${sg_id}/related_resources/biz_resources/${res_biz_id}/load_balancers/list`;
    try {
      const [listRes, countRes] = await Promise.all<
        [
          Promise<IListResData<ISecurityGroupRelLoadBalancerByBizItem[]>>,
          Promise<IListResData<ISecurityGroupRelLoadBalancerByBizItem[]>>,
        ]
      >([http.post(api, enableCount(payload, false)), http.post(api, enableCount(payload, true))]);
      const [{ details: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isQueryRelLoadBalancerByBizLoading.value = false;
    }
  };

  // 查询安全组关联的云上资源数量
  const isQueryRelatedResourcesCountLoading = ref(false);
  const queryRelatedResourcesCount = async (ids: string[]) => {
    isQueryRelatedResourcesCountLoading.value = true;
    try {
      const res: IListResData<ISecurityGroupRelResCountItem[]> = await http.post(
        `/api/v1/cloud/${getBusinessApiPath()}security_groups/related_resources/query_count`,
        { ids },
      );
      const { details = [] } = res.data ?? {};
      return details;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isQueryRelatedResourcesCountLoading.value = false;
    }
  };

  // 更新安全组管理属性
  const isUpdateMgmtAttrLoading = ref(false);
  const updateMgmtAttr = async (
    id: string,
    payload?: {
      mgmt_type?: string;
      manager?: string;
      bak_manager?: string;
      usage_biz_ids?: number[];
      mgmt_biz_id?: number;
    },
  ) => {
    isUpdateMgmtAttrLoading.value = true;
    try {
      await http.patch(`/api/v1/cloud/${getBusinessApiPath()}security_groups/${id}/mgmt_attrs`, payload);
    } finally {
      isUpdateMgmtAttrLoading.value = false;
    }
  };

  // 批量查询安全组规则数量
  const isBatchQueryRuleCountLoading = ref(false);
  const batchQueryRuleCount = async (security_group_ids: string[]) => {
    isBatchQueryRuleCountLoading.value = true;
    try {
      const res: IQueryResData<Record<string, number>> = await http.post(
        `/api/v1/cloud/${getBusinessApiPath()}security_groups/rules/count`,
        { security_group_ids },
      );
      return res.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isBatchQueryRuleCountLoading.value = false;
    }
  };

  const isQueryRuleCountAndRelatedResourcesLoading = ref(false);
  const queryRuleCountAndRelatedResources = async (ids: string[]) => {
    isQueryRuleCountAndRelatedResourcesLoading.value = true;
    try {
      const [ruleCountMap, relatedResourcesList] = await Promise.all([
        batchQueryRuleCount(ids),
        queryRelatedResourcesCount(ids),
      ]);
      return { ruleCountMap, relatedResourcesList };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isQueryRuleCountAndRelatedResourcesLoading.value = false;
    }
  };

  // 查询资源关联的安全组信息
  const isBatchQuerySecurityGroupByResIdsLoading = ref(false);
  const batchQuerySecurityGroupByResIds = async (res_type: 'cvm' | 'loadbalancer' | string, res_ids: string[]) => {
    isBatchQuerySecurityGroupByResIdsLoading.value = true;
    try {
      const res: IQueryResData<IResourceBoundSecurityGroupItem[]> = await http.post(
        `/api/v1/cloud/${getBusinessApiPath()}security_groups/res/${res_type}/batch`,
        { res_ids },
      );
      return res.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isBatchQuerySecurityGroupByResIdsLoading.value = false;
    }
  };

  // 查询资源关联的安全组信息，并合并
  const pullSecurityGroup = async (res_type: string, originList: SecurityGroupRelResourceByBizItem[]) => {
    if (!originList.length) return originList;
    const res_ids = originList.map((item) => item.id);
    const securityGroups = await batchQuerySecurityGroupByResIds(res_type, res_ids);
    return originList.map((item) => {
      return { ...item, security_groups: securityGroups.find((sg) => sg.res_id === item.id)?.security_groups || [] };
    });
  };

  // 安全组批量绑定CVM
  const isBatchAssociateCvmsLoading = ref(false);
  const batchAssociateCvms = async (data: { security_group_id: string; cvm_ids: string[] }) => {
    isBatchAssociateCvmsLoading.value = true;
    try {
      await http.post(`/api/v1/cloud/${getBusinessApiPath()}security_groups/associate/cvms/batch`, data);
    } finally {
      isBatchAssociateCvmsLoading.value = false;
    }
  };

  // 安全组批量解绑主机
  const isBatchDisassociateCvmsLoading = ref(false);
  const batchDisassociateCvms = async (data: { security_group_id: string; cvm_ids: string[] }) => {
    isBatchDisassociateCvmsLoading.value = true;
    try {
      await http.post(`/api/v1/cloud/${getBusinessApiPath()}security_groups/disassociate/cvms/batch`, data);
    } finally {
      isBatchDisassociateCvmsLoading.value = false;
    }
  };

  // 批量查询安全组使用业务负责人列表
  const isQueryUsageBizMaintainersLoading = ref(false);
  const queryUsageBizMaintainers = async (security_group_ids: string[]) => {
    isQueryUsageBizMaintainersLoading.value = true;
    try {
      const res: IQueryResData<ISecurityGroupUsageBizMaintainerItem[]> = await http.post(
        `/api/v1/cloud/${getBusinessApiPath()}security_groups/maintainers_info/list`,
        { security_group_ids },
      );
      return res.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isQueryUsageBizMaintainersLoading.value = false;
    }
  };

  return {
    isFullListLoading,
    getFullList,
    isFullRuleLoading,
    getFullRuleList,
    isAssignPreviewLoading,
    getAssignPreview,
    isBatchAssignToBizLoading,
    batchAssignToBiz,
    isBatchUpdateMgmtAttrLoading,
    batchUpdateMgmtAttr,
    isQueryRelBusinessLoading,
    queryRelBusiness,
    isQueryRelCvmByBizLoading,
    queryRelCvmByBiz,
    isQueryRelLoadBalancerByBizLoading,
    queryRelLoadBalancerByBiz,
    isQueryRelatedResourcesCountLoading,
    queryRelatedResourcesCount,
    isUpdateMgmtAttrLoading,
    updateMgmtAttr,
    isBatchQueryRuleCountLoading,
    batchQueryRuleCount,
    isQueryRuleCountAndRelatedResourcesLoading,
    queryRuleCountAndRelatedResources,
    isBatchQuerySecurityGroupByResIdsLoading,
    batchQuerySecurityGroupByResIds,
    pullSecurityGroup,
    isBatchAssociateCvmsLoading,
    batchAssociateCvms,
    isBatchDisassociateCvmsLoading,
    batchDisassociateCvms,
    isQueryUsageBizMaintainersLoading,
    queryUsageBizMaintainers,
  };
});
