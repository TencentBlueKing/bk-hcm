import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { IListResData, QueryBuilderType, QueryFilterType } from '@/typings';
import { SecondaryAccountResourceTypeEnum, VendorEnum } from '@/common/constant';
import { enableCount, resolveBizApiPath } from '@/utils/search';
import rollRequest from '@blueking/roll-request';

// 二级账号项接口定义
export interface ISecondaryAccountItem {
  id: string;
  vendor: string;
  name: string;
  managers: string[];
  security_managers: string[];
  type: string;
  site: string;
  price: string;
  price_unit: string;
  memo: string;
  bk_biz_id: number;
  usage_biz_ids: number[];
  email: string;
  cloud_created_at: string;
  sync_status: string;
  sync_failed_reason: string;
  sub_account_count: number;
  account_secret_count: number;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  extension: {
    cloud_main_account_id: string;
    cloud_secret_id: string;
    cloud_sub_account_id: string;
    [k: string]: any;
  };
  [k: string]: any;
}

// 创建二级账号参数接口定义
export interface IAccountCreateParams {
  vendor: string;
  name: string;
  managers: string[];
  security_managers?: string[];
  type: string;
  site: string;
  bk_biz_id?: number;
  usage_biz_ids: number[];
  memo?: string;
  extension: Record<string, any>;
  remark?: string;
}

// 更新二级账号参数接口定义
export interface IAccountUpdateParams {
  name?: string;
  managers?: string[];
  security_managers?: string[];
  bk_biz_id?: number;
  usage_biz_ids?: number[];
  memo?: string;
  extension?: Record<string, any>;
}

// 账号密钥项接口定义
export interface IAccountSecretItem {
  id: string;
  vendor: string;
  type: string; // 密钥类型：resource(资源管理)、security(安全管理)
  status: string; // 密钥状态：normal(正常)、invalid(失效)
  account_id: string;
  extension: {
    cloud_secret_id: string;
    cloud_main_account_id?: string;
    cloud_sub_account_id?: string;
  };
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
}

// 密钥校验响应接口定义
export interface ISecretCheckResult {
  cloud_main_account_id: string;
  cloud_sub_account_id: string;
}

// 创建/更新密钥参数接口定义
export interface ISecretCreateParams {
  account_id: string;
  type: string;
  extension: {
    cloud_secret_id: string;
    cloud_secret_key: string;
  };
}

export interface ISecretUpdateParams {
  type?: string;
  extension?: {
    cloud_secret_id: string;
    cloud_secret_key: string;
  };
}

export interface ISecretCheckParams {
  account_id: string;
  type: string;
  extension: {
    cloud_secret_id: string;
    cloud_secret_key: string;
  };
}

export const useSecondaryAccountStore = defineStore('secondaryAccount', () => {
  const accountListLoading = ref(false);
  const secretListLoading = ref(false);
  const secretCheckLoading = ref(false);

  // 根据账号ID缓存二级账号列表
  const allSecondaryAccountCacheList = ref<Map<ISecondaryAccountItem['id'], ISecondaryAccountItem>>(new Map());

  /**
   * 获取二级账号列表
   * @param params 查询参数
   */
  const getSecondaryAccountList = async (params: QueryBuilderType & { bk_biz_id: number }) => {
    const { bk_biz_id, ...data } = params;
    accountListLoading.value = true;
    const api = `/api/v1/cloud/bizs/${bk_biz_id}/accounts/list`;
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<ISecondaryAccountItem[]>>, Promise<IListResData<ISecondaryAccountItem[]>>]
      >([http.post(api, enableCount(data, false)), http.post(api, enableCount(data, true))]);
      const [{ details: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      accountListLoading.value = false;
    }
  };

  /**
   * 根据账号ID获取二级账号列表，带缓存
   * @param accountIds 账号ID列表
   * @param vendor 云厂商
   * @param resType 资源类型：permission_policy_library | sub_account | sub_account_secret | permission_template
   * @param bizId 业务ID
   */
  const getSecondaryAccountListByAccountIds = async (
    accountIds: string[],
    vendor: VendorEnum,
    resType: SecondaryAccountResourceTypeEnum,
    bizId: number,
  ) => {
    const maxIdsLength = 100; // 该接口每次最大ID数目
    const api = `/api/v1/cloud/${resolveBizApiPath(bizId)}vendors/${vendor}/accounts/list/by/res_type`;
    const cachedIds = allSecondaryAccountCacheList.value.keys();
    const cachedIdSet = new Set(cachedIds);
    const newIds = accountIds.filter((id) => !cachedIdSet.has(`${id}@${resType}@${bizId}`));
    while (newIds.length) {
      const res = await http.post(api, {
        ids: newIds.splice(0, maxIdsLength),
        res_type: resType,
      });
      const list = res?.data?.details ?? [];
      for (const item of list) {
        allSecondaryAccountCacheList.value.set(`${item.id}@${resType}@${bizId}`, item);
      }
    }
    return accountIds.map((id) => allSecondaryAccountCacheList.value.get(`${id}@${resType}@${bizId}`)).filter(Boolean);
  };

  /**
   * 使用 rollRequest 获取二级账号全量列表（用于前端分页）
   * @param bk_biz_id 业务ID
   * @param filter 过滤条件
   * @param onProgress 进度回调，每批次数据返回时调用
   */
  const getSecondaryAccountFullList = async (
    bk_biz_id: number,
    filter: QueryFilterType,
    onProgress?: (list: ISecondaryAccountItem[], count: number) => void,
  ): Promise<ISecondaryAccountItem[]> => {
    accountListLoading.value = true;
    const api = `/api/v1/cloud/bizs/${bk_biz_id}/accounts/list`;
    const allList: ISecondaryAccountItem[] = [];

    try {
      const listGen = await rollRequest({ httpClient: http, pageEnableCountKey: 'count' }).rollReqUseCount<
        IListResData<ISecondaryAccountItem[]>
      >(
        api,
        { filter },
        {
          limit: 500, // 每批次拉取500条
          countGetter: (res) => res.data.count,
          listGetter: (res) => res.data.details,
          generator: true,
        },
        true,
      );

      // 串行迭代请求，避免一次性请求过多数据导致阻塞
      for await (const res of listGen) {
        const details = res.data?.details || [];
        allList.push(...details);
        // 回调通知进度
        onProgress?.(allList, res.data?.count || allList.length);
        // 完成第一次请求即关闭 loading 效果，其余请求静默处理
        accountListLoading.value = false;
      }

      return allList;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      accountListLoading.value = false;
    }
  };

  /**
   * 获取二级账号详情（通过列表接口按 id 查询单条）
   */
  const getSecondaryAccountDetail = async (bk_biz_id: number, id: string): Promise<ISecondaryAccountItem | null> => {
    const api = `/api/v1/cloud/bizs/${bk_biz_id}/accounts/list`;
    try {
      const res = await http.post(api, {
        filter: { rules: [{ field: 'id', op: 'eq', value: id }], op: 'and' },
        page: { count: false, start: 0, limit: 1 },
      });
      const list = res?.data?.details ?? [];
      return list.length > 0 ? list[0] : null;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 同步指定账号下指定资源
   * 接口文档：业务下同步指定账号下指定资源.md
   * @param bk_biz_id 业务ID
   * @param vendor 云厂商
   * @param account_id 账号ID
   * @param res 资源名称 (security_group | load_balancer | sub_account)
   * @param params 同步参数
   */
  const syncAccountResource = async (
    bk_biz_id: number,
    vendor: string,
    account_id: string,
    res: 'security_group' | 'load_balancer' | 'sub_account',
    params?: {
      regions?: string[];
      cloud_ids?: string[];
      tag_filters?: Record<string, string[]>;
      resource_group_names?: string[]; // Azure 专用
    },
  ) => {
    try {
      const response = await http.post(
        `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/accounts/${account_id}/resources/${res}/sync_by_cond`,
        params || {},
      );
      return response?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 批量同步多个二级账号的子账号资源
   * @param bk_biz_id 业务ID
   * @param vendor 云厂商
   * @param account_ids 账号ID列表
   */
  const syncSecondaryAccounts = async (bk_biz_id: number, vendor: string, account_ids: string[]) => {
    const results: { success: string[]; failed: { id: string; error: any }[] } = {
      success: [],
      failed: [],
    };

    // 并行同步所有账号
    await Promise.all(
      account_ids.map(async (account_id) => {
        try {
          await syncAccountResource(bk_biz_id, vendor, account_id, 'sub_account');
          results.success.push(account_id);
        } catch (error) {
          results.failed.push({ id: account_id, error });
        }
      }),
    );

    return results;
  };

  /**
   * 创建二级账号（提交申请）
   * @param bk_biz_id 业务ID
   * @param params 账号参数
   */
  const createSecondaryAccount = async (bk_biz_id: number, params: IAccountCreateParams) => {
    try {
      const res = await http.post(`/api/v1/cloud/bizs/${bk_biz_id}/applications/types/add_account`, params);
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 更新二级账号
   * @param bk_biz_id 业务ID
   * @param account_id 账号ID
   * @param params 更新参数
   */
  const updateSecondaryAccount = async (bk_biz_id: number, account_id: string, params: IAccountUpdateParams) => {
    try {
      const res = await http.patch(`/api/v1/cloud/bizs/${bk_biz_id}/accounts/${account_id}`, params);
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 获取账号密钥列表
   * @param bk_biz_id 业务ID
   * @param vendor 云厂商
   * @param params 查询参数
   */
  const getAccountSecretList = async (
    bk_biz_id: number,
    vendor: string,
    params: { filter: any; page: any },
  ): Promise<{ list: IAccountSecretItem[]; count: number }> => {
    secretListLoading.value = true;
    const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/account_secrets/list`;
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<IAccountSecretItem[]>>, Promise<IListResData<IAccountSecretItem[]>>]
      >([http.post(api, enableCount(params, false)), http.post(api, enableCount(params, true))]);
      const [{ details: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      secretListLoading.value = false;
    }
  };

  /**
   * 获取账号密钥详情（通过列表接口按 id 查询单条）
   */
  const getAccountSecretDetail = async (
    bk_biz_id: number,
    vendor: string,
    id: string,
  ): Promise<IAccountSecretItem | null> => {
    const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/account_secrets/list`;
    try {
      const res = await http.post(api, {
        filter: { rules: [{ field: 'id', op: 'eq', value: id }], op: 'and' },
        page: { count: false, start: 0, limit: 1 },
      });
      const list = res?.data?.details ?? [];
      return list.length > 0 ? list[0] : null;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 创建账号密钥
   * @param bk_biz_id 业务ID
   * @param params 密钥参数
   */
  const createAccountSecret = async (bk_biz_id: number, params: ISecretCreateParams) => {
    try {
      const res = await http.post(`/api/v1/cloud/bizs/${bk_biz_id}/account_secrets/create`, params);
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 更新账号密钥
   * @param bk_biz_id 业务ID
   * @param secretId 密钥ID
   * @param params 更新参数
   */
  const updateAccountSecret = async (bk_biz_id: number, secretId: string, params: ISecretUpdateParams) => {
    try {
      const res = await http.patch(`/api/v1/cloud/bizs/${bk_biz_id}/account_secrets/${secretId}`, params);
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 删除账号密钥
   * @param bk_biz_id 业务ID
   * @param vendor 云厂商
   * @param ids 密钥ID列表
   */
  const deleteAccountSecret = async (bk_biz_id: number, vendor: string, ids: string[]) => {
    try {
      const res = await http.delete(`/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/account_secrets/batch`, {
        data: { ids },
      });
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 校验账号密钥
   * @param bk_biz_id 业务ID
   * @param params 校验参数
   */
  const checkAccountSecret = async (bk_biz_id: number, params: ISecretCheckParams): Promise<ISecretCheckResult> => {
    secretCheckLoading.value = true;
    try {
      const res = await http.post(`/api/v1/cloud/bizs/${bk_biz_id}/account_secrets/check`, params);
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      secretCheckLoading.value = false;
    }
  };

  return {
    accountListLoading,
    secretListLoading,
    secretCheckLoading,
    allSecondaryAccountCacheList,
    getSecondaryAccountList,
    getSecondaryAccountDetail,
    getSecondaryAccountFullList,
    getSecondaryAccountListByAccountIds,
    syncAccountResource,
    syncSecondaryAccounts,
    createSecondaryAccount,
    updateSecondaryAccount,
    getAccountSecretList,
    getAccountSecretDetail,
    createAccountSecret,
    updateAccountSecret,
    deleteAccountSecret,
    checkAccountSecret,
  };
});
