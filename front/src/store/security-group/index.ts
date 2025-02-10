import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { VendorEnum } from '@/common/constant';
import { IListResData, IQueryResData, QueryBuilderType } from '@/typings';
import { enableCount } from '@/utils/search';
import { useWhereAmI } from '@/hooks/useWhereAmI';

export interface ISgItem {
  id: string;
  vendor: VendorEnum;
  cloud_id: string;
  region: string;
  name: string;
  manager: string;
  bak_manager: string;
  usage_biz_ids: number[];
  mgmt_biz_id: number;
  mgmt_type: string;
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
}

interface ISgAssignBizsPreviewItem {
  id: string;
  assignable: boolean;
  reason: string;
  assigned_biz_id: number;
}

type SgRelType = 'cvm' | 'load_balancer';
interface IBizResourceCount {
  bk_biz_id: number;
  res_count: number;
}
type ISgRelBusiness = Record<SgRelType, IBizResourceCount[]>;

type SgRelCvmCommonFields = {
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
interface ISgRelCvmByBizItem extends SgRelCvmCommonFields {
  zone: string;
  cloud_vpc_ids: string[];
  cloud_subnet_ids: string[];
}
interface ISgRelLoadBalancerByBizItem extends SgRelCvmCommonFields {
  main_zones: string[];
  backup_zones: string[];
  cloud_vpc_id: string;
  vpc_id: string;
  network_type: string;
  domain: string;
  memo: string;
}

interface ISgRelResCountItem {
  id: string;
  resources: Array<{
    res_name: 'cvm' | 'load_balancer' | 'db' | 'container';
    count: number;
  }>;
}
// 安全组单个操作项的类型
export type ISgOperateItem = ISgItem & ISgRelResCountItem & { rule_count?: number } & { [key: string]: any };

export const useSecurityGroupStore = defineStore('security-group', () => {
  const { getBusinessApiPath } = useWhereAmI();

  // 预览安全组分配到业务的结果，是否可分配
  const isQueryAssignBizsPreviewLoading = ref(false);
  const queryAssignBizsPreview = async (ids: string[]) => {
    isQueryAssignBizsPreviewLoading.value = true;
    try {
      const res: IQueryResData<ISgAssignBizsPreviewItem[]> = await http.post(
        '/api/v1/cloud/security_groups/assign/bizs/preview',
        { ids },
      );
      return res.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isQueryAssignBizsPreviewLoading.value = false;
    }
  };

  // 批量分配安全组到业务
  const isBatchAssignSgToBizLoading = ref(false);
  const batchAssignSgToBiz = async (sg_ids: string[]) => {
    isBatchAssignSgToBizLoading.value = true;
    try {
      await http.post('/api/v1/cloud/security_groups/assign/bizs', { sg_ids });
    } finally {
      isBatchAssignSgToBizLoading.value = false;
    }
  };

  // 批量更新安全组管理属性，仅当所有管理属性均不存在时才允许编辑，所有管理属性都要提供
  // 注意：通过该接口更新的安全组会被默认设置为业务管理类型，不可再更改为平台管理类型
  const isBatchUpdateSgMgmtAttrLoading = ref(false);
  const batchUpdateSgMgmtAttr = async (
    security_groups: Array<{ id: string; manager: string; bak_manager: string[]; mgmt_biz_id: number }>,
  ) => {
    isBatchUpdateSgMgmtAttrLoading.value = true;
    try {
      await http.patch('/api/v1/cloud/security_groups/mgmt_attrs/batch', { security_groups });
    } finally {
      isBatchUpdateSgMgmtAttrLoading.value = false;
    }
  };

  // 查询安全组关联资源所属的业务列表，目前仅支持查询关联的CVM和CLB资源。
  // 返回的业务列表中，一定包含当前业务，且一定排在第一个
  const isQuerySgRelBusinessLoading = ref(false);
  const querySgRelBusiness = async (security_group_id: string) => {
    isQuerySgRelBusinessLoading.value = true;
    try {
      const res: IQueryResData<ISgRelBusiness> = await http.post(
        `/api/v1/cloud/security_groups/${security_group_id}/related_resources/bizs/list`,
        { security_group_id },
      );
      return res.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isQuerySgRelBusinessLoading.value = false;
    }
  };

  // 查询安全组关联的cvm列表，仅展示cvm摘要信息
  const isQuerySgRelCvmByBizLoading = ref(false);
  const querySgRelCvmByBiz = async (sg_id: string, res_biz_id: number, payload: QueryBuilderType) => {
    isQuerySgRelCvmByBizLoading.value = true;
    const api = `/security_groups/${sg_id}/related_resources/biz_resources/${res_biz_id}/cvms/list`;
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<ISgRelCvmByBizItem[]>>, Promise<IListResData<ISgRelCvmByBizItem[]>>]
      >([http.post(api, enableCount(payload, false)), http.post(api, enableCount(payload, true))]);
      const [{ details: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isQuerySgRelCvmByBizLoading.value = false;
    }
  };

  // 查询安全组关联的负载均衡列表，仅展示负载均衡摘要信息。
  const isQuerySgRelLoadBalancerByBizLoading = ref(false);
  const querySgRelLoadBalancerByBiz = async (sg_id: string, res_biz_id: number, payload: QueryBuilderType) => {
    isQuerySgRelLoadBalancerByBizLoading.value = true;
    const api = `/security_groups/${sg_id}/related_resources/biz_resources/${res_biz_id}/load_balancers/list`;
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<ISgRelLoadBalancerByBizItem[]>>, Promise<IListResData<ISgRelLoadBalancerByBizItem[]>>]
      >([http.post(api, enableCount(payload, false)), http.post(api, enableCount(payload, true))]);
      const [{ details: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isQuerySgRelLoadBalancerByBizLoading.value = false;
    }
  };

  // 查询安全组关联的云上资源数量
  const isQuerySgRelatedResourcesLoading = ref(false);
  const querySgRelatedResources = async (ids: string[]) => {
    isQuerySgRelatedResourcesLoading.value = true;
    try {
      const res: IQueryResData<ISgRelResCountItem[]> = await http.post(
        '/api/v1/cloud/security_groups/related_resources/query_count',
        { ids },
      );
      return res.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isQuerySgRelatedResourcesLoading.value = false;
    }
  };

  // 更新安全组管理属性
  const isUpdateSgMgmtAttrLoading = ref(false);
  const updateSgMgmtAttr = async (
    id: string,
    payload?: Array<{
      mgmt_type?: string;
      manager?: string;
      bak_manager?: string;
      usage_biz_ids?: number[];
      mgmt_biz_id?: number;
    }>,
  ) => {
    isUpdateSgMgmtAttrLoading.value = true;
    try {
      await http.patch(`/api/v1/cloud/security_groups/${id}/mgmt_attrs`, payload);
    } finally {
      isUpdateSgMgmtAttrLoading.value = false;
    }
  };

  // 批量查询安全组规则数量
  const isBatchQuerySgRuleCountLoading = ref(false);
  const batchQuerySgRuleCount = async (security_group_ids: string[]) => {
    isBatchQuerySgRuleCountLoading.value = true;
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
      isBatchQuerySgRuleCountLoading.value = false;
    }
  };

  return {
    isQueryAssignBizsPreviewLoading,
    queryAssignBizsPreview,
    isBatchAssignSgToBizLoading,
    batchAssignSgToBiz,
    isBatchUpdateSgMgmtAttrLoading,
    batchUpdateSgMgmtAttr,
    isQuerySgRelBusinessLoading,
    querySgRelBusiness,
    isQuerySgRelCvmByBizLoading,
    querySgRelCvmByBiz,
    isQuerySgRelLoadBalancerByBizLoading,
    querySgRelLoadBalancerByBiz,
    isQuerySgRelatedResourcesLoading,
    querySgRelatedResources,
    isUpdateSgMgmtAttrLoading,
    updateSgMgmtAttr,
    isBatchQuerySgRuleCountLoading,
    batchQuerySgRuleCount,
  };
});
