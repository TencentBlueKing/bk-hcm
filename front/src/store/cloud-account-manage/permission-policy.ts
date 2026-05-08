import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { IListResData, QueryRuleOPEnum, QueryBuilderType, IAPIResData } from '@/typings';
import rollRequest from '@blueking/roll-request';
import { resolveBizApiPath } from '@/utils/search';
import { VendorEnum } from '@/common/constant';
import { ListGeneratorFactory } from '@/components/form/list.vue';

// 权限策略列表
export interface IPermissionPolicyItem {
  id: string;
  name: string; // 策略库名称
  version: number; // 当前版本号
  bk_biz_ids: number[]; // 允许使用的业务ID列表
  memo: string; // 描述
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  policy_document: string; // 当前版本的策略JSON内容
  policy_hash: string; // 当前版本的策略HASH
  associated_account_count: number; // 关联二级账号数
}

// 已应用账号列表
export interface IPermissionAppliedItem {
  id: string; // 权限模版ID
  name: string; // 模板名称
  cloud_id: string; // 云上策略ID
  vendor: string; // 云厂商
  account_id: string; // 所属二级账号ID
  policy_library_id: string; // 应用时的策略库ID
  policy_library_version: number; // 应用时的策略库版本
  policy_library_sync_time: string; // 同步时间
  policy_document: string; // 策略JSON内容
  policy_hash: string; // 策略内容哈希值
  memo: string; // 描述
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  extension: { cloud_type: number }; // 云厂商扩展字段
  apply_status?: 'applied' | 'pending' | 'data_mismatch';
}

// 应用账号接口参数
export interface IAppliedParams {
  bizId?: number;
  vendor: VendorEnum;
  id: string; // 策略库ID
  selectedIds: string[]; // 已选择的ID
}

// 新增/编辑权限列表参数
export interface IOperationPermissionPolicyParams {
  id?: string;
  name: string;
  policy_document: string;
  bk_biz_ids: number[];
  memo: string;
}

export interface IApplyResultItem {
  account_id?: string; // 二级账号ID
  permission_template_id?: string; // 权限模版账号ID
  status: 'success' | 'failed';
  reason?: string; // status是failed返回
}

export const usePermissionPolicyStore = defineStore('permissionPolicy', () => {
  const permissionPolicyListLoading = ref(false);
  const unappliedAccountIdsListLoading = ref(false);
  const appliedAccountIdsListLoading = ref(false);

  /**
   * 创建权限策略库
   * @param vendor 云账户
   * @param params 权限策略库参数
   */
  const createPermissionPolicy = async (vendor: string, params: IOperationPermissionPolicyParams) => {
    try {
      const res = await http.post(`/api/v1/cloud/vendors/${vendor}/permission_policy_libraries/create`, params);
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 更新权限策略库
   * @param vendor 云账户
   * @param params 更新参数
   */
  const updatePermissionPolicy = async (vendor: string, params: IOperationPermissionPolicyParams) => {
    try {
      const { id, name, policy_document, bk_biz_ids, memo } = params;
      const res = await http.patch(`/api/v1/cloud/vendors/${vendor}/permission_policy_libraries/${id}`, {
        name,
        policy_document,
        bk_biz_ids,
        memo,
      });
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 获取权限策略库列表
   * @param bizId 业务ID
   * @param vendor 云厂商
   * @param params 查询参数
   */
  const getPermissionPolicyList = async (
    bizId: number,
    vendor: string,
    params: QueryBuilderType,
  ): Promise<{ list: IPermissionPolicyItem[]; count: number }> => {
    permissionPolicyListLoading.value = true;

    const api = `/api/v1/cloud/${resolveBizApiPath(bizId)}vendors/${vendor}/permission_policy_libraries/list`;
    try {
      // 构建请求参数
      const requestData = { ...params };

      // 获取列表数据
      const listRes = await http.post(api, {
        ...requestData,
        page: { ...requestData.page, count: false },
      });

      // 获取总数
      const countRes = await http.post(api, {
        ...requestData,
        page: { count: true, start: 0, limit: 0 },
      });

      const list = listRes?.data?.details || [];
      const count = countRes?.data?.count || 0;

      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      permissionPolicyListLoading.value = false;
    }
  };

  /**
   * 权限策略库关联的二级账号列表
   * @param bizId 业务ID
   * @param vendor 云账户
   * @param id 策略库ID
   * @param count 关联二级账号数
   */
  const getPermissionAssoAccountList = async (
    bizId: number,
    vendor: string,
    id: string,
    count: number,
  ): Promise<string[]> => {
    appliedAccountIdsListLoading.value = true;

    if (count === 0) return Promise.resolve([]);

    const api = `/api/v1/cloud/${resolveBizApiPath(
      bizId,
    )}vendors/${vendor}/permission_policy_libraries/${id}/account_ids`;
    try {
      // 获取列表数据
      const listRes = await http.get(api);

      const list = listRes?.data?.account_ids || [];

      return list;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      appliedAccountIdsListLoading.value = false;
    }
  };

  /**
   * 权限策略库应用到二级账号时未应用的列表
   * @param bizId 业务ID
   * @param vendor 云账户
   * @param id 策略库ID
   */
  const getUnappliedAccountIdsList = async (bizId: number, vendor: string, id: string): Promise<string[]> => {
    unappliedAccountIdsListLoading.value = true;

    const api = `/api/v1/cloud/${resolveBizApiPath(
      bizId,
    )}vendors/${vendor}/permission_policy_libraries/${id}/unapplied_account_ids`;
    try {
      // 获取列表数据
      const listRes = await http.get(api);

      const list = listRes?.data?.account_ids || [];

      return list;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      unappliedAccountIdsListLoading.value = false;
    }
  };

  /**
   * 权限策略库应用到二级账号时已经应用的列表
   * @param bizId 业务ID
   * @param vendor 云账户
   * @param id 策略库ID
   */
  const getAppliedAccountIdsList = async (
    bizId: number,
    vendor: string,
    id: string,
  ): Promise<IPermissionAppliedItem[]> => {
    appliedAccountIdsListLoading.value = true;

    const api = `/api/v1/cloud/${resolveBizApiPath(
      bizId,
    )}vendors/${vendor}/permission_policy_libraries/${id}/permission_templates`;
    try {
      // 获取列表数据
      const listRes = await http.get(api);

      const list = listRes?.data?.details || [];

      return list;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      appliedAccountIdsListLoading.value = false;
    }
  };

  /**
   * 权限策略库应用到二级账号确认应用（创建）-- 非业务下
   * @param params 参数 包含下面3个
   * @param vendor 云账户
   * @param id 策略库ID
   * @param accountIds 目标二级账号ID列表
   */
  const createAppliedAccount = async (params: IAppliedParams) => {
    const { vendor, id, selectedIds: accountIds } = params;
    const api = `/api/v1/cloud/vendors/${vendor}/permission_policy_libraries/${id}/apply`;
    try {
      const res: IAPIResData<{ results: IApplyResultItem[] }> = await http.post(api, {
        account_ids: accountIds,
      });

      const list = res?.data?.results || [];

      return list;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 权限策略库应用到二级账号确认应用（创建）-- 业务下
   * @param params 参数 包含下面4个
   * @param bizId 业务ID
   * @param vendor 云账户
   * @param id 策略库ID
   * @param accountIds 目标二级账号ID列表
   */
  const createAppliedAccountBiz = async (params: IAppliedParams) => {
    const { bizId, vendor, id, selectedIds: accountIds } = params;
    const api = `/api/v1/cloud/bizs/${bizId}/vendors/${vendor}/applications/types/apply_permission_policy_library_create`;
    try {
      const res: IAPIResData<{ ids: string[] }> = await http.post(api, {
        account_ids: accountIds,
        policy_library_id: id,
      });

      const list = res?.data?.ids || [];

      return list;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 权限策略库应用到二级账号确认应用（更新）-- 非业务下
   * @param params 参数 包含下面3个
   * @param vendor 云账户
   * @param id 策略库ID
   * @param templateIds 目前权限模板ID列表
   */
  const updateAppliedAccount = async (params: IAppliedParams) => {
    const { vendor, id, selectedIds: templateIds } = params;
    const api = `/api/v1/cloud/vendors/${vendor}/permission_policy_libraries/${id}/apply`;
    try {
      const res: IAPIResData<{ results: IApplyResultItem[] }> = await http.put(api, {
        permission_template_ids: templateIds,
      });

      const list = res?.data?.results || [];

      return list;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 权限策略库应用到二级账号确认应用（更新）-- 业务下
   * @param params 参数 包含下面4个
   * @param bizId 业务ID
   * @param vendor 云账户
   * @param id 策略库ID
   * @param templateIds 目前权限模板ID列表
   */
  const updateAppliedAccountBiz = async (params: IAppliedParams) => {
    const { bizId, vendor, id, selectedIds: templateIds } = params;
    const api = `/api/v1/cloud/bizs/${bizId}/vendors/${vendor}/applications/types/apply_permission_policy_library_update`;
    try {
      const res: IAPIResData<{ ids: string[] }> = await http.post(api, {
        permission_template_ids: templateIds,
        policy_library_id: id,
      });

      const list = res?.data?.ids || [];

      return list;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  const createPolicyLibraryListGenerator = (vendor: string, bizId?: number): ListGeneratorFactory => {
    return async function* (keywordOrOptions) {
      const api = `/api/v1/cloud/${resolveBizApiPath(bizId)}vendors/${vendor}/permission_policy_libraries/list`;
      const rules: Array<{ field: string; op: string; value: any }> = [];
      const keyword = typeof keywordOrOptions === 'string' ? keywordOrOptions : undefined;
      const options = typeof keywordOrOptions === 'object' ? keywordOrOptions : undefined;
      if (keyword) rules.push({ field: 'name', op: QueryRuleOPEnum.CS, value: keyword });
      if (options?.ids?.length) rules.push({ field: 'id', op: QueryRuleOPEnum.IN, value: options.ids });
      const filterParams = { op: QueryRuleOPEnum.AND, rules };

      const gen = await rollRequest({ httpClient: http, pageEnableCountKey: 'count' }).rollReqUseCount<
        IListResData<IPermissionPolicyItem[]>
      >(
        api,
        { filter: filterParams },
        { limit: 500, countGetter: (res) => res.data.count, listGetter: (res) => res.data.details, generator: true },
        true,
      );

      for (const promise of gen) {
        const res = await promise;
        yield res?.data?.details ?? [];
      }
    };
  };

  return {
    permissionPolicyListLoading,
    unappliedAccountIdsListLoading,
    appliedAccountIdsListLoading,
    createPermissionPolicy,
    updatePermissionPolicy,
    getPermissionPolicyList,
    createPolicyLibraryListGenerator,
    getUnappliedAccountIdsList,
    getAppliedAccountIdsList,
    createAppliedAccount,
    updateAppliedAccount,
    getPermissionAssoAccountList,
    createAppliedAccountBiz,
    updateAppliedAccountBiz,
  };
});
