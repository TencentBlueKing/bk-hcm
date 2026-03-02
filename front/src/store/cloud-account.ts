import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { IListResData, QueryBuilderType, QueryFilterType } from '@/typings';
import { VendorEnum } from '@/common/constant';
import { enableCount } from '@/utils/search';
import rollRequest from '@blueking/roll-request';
import {
  USE_MOCK,
  mockGetSubAccountSecretList,
  mockUpdateSubAccountSecretStatus,
  mockDeleteSubAccountSecret,
} from '@/views/cloud-account-manage/cloud-secret/mock';

// 二级账号项接口定义
export interface ISecondaryAccountItem {
  id: string;
  vendor: VendorEnum;
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
    login_flag?: string;
    action_flag?: string;
    console_login?: number;
    [k: string]: any;
  };
  [k: string]: any;
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

// 三级账号密钥项接口定义
export interface ISubAccountSecretItem {
  id: string;
  vendor: string;
  status: 'enabled' | 'disabled';
  account_id: string;
  sub_account_id: string;
  extension: {
    cloud_secret_id: string;
    cloud_main_account_id: string;
    cloud_sub_account_id: string;
    console_login?: number;
  };
  tenant_id?: string;
  cloud_created_at: string;
  disabled_time?: string;
  last_used_time?: string;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  sub_account_manager?: string;
  account_manager?: string;
}

// 更新密钥状态参数
export interface IUpdateSecretStatusParams {
  id: string;
  status: 'enabled' | 'disabled';
}

export const useCloudAccountStore = defineStore('cloudAccount', () => {
  const accountListLoading = ref(false);
  const secretListLoading = ref(false);
  const secretCheckLoading = ref(false);
  const subAccountSecretListLoading = ref(false);

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
   * 获取三级账号密钥列表
   * @param bk_biz_id 业务ID
   * @param vendor 云厂商
   * @param params 查询参数
   */
  const getSubAccountSecretList = async (
    bk_biz_id: number,
    vendor: string,
    params: { filter?: any; page: any } & Record<string, any>,
  ): Promise<{ list: ISubAccountSecretItem[]; count: number }> => {
    subAccountSecretListLoading.value = true;

    // 使用 Mock 数据
    if (USE_MOCK) {
      try {
        const result = await mockGetSubAccountSecretList(params as any);
        return result as { list: ISubAccountSecretItem[]; count: number };
      } finally {
        subAccountSecretListLoading.value = false;
      }
    }

    // 使用真实接口
    const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/sub_account_secrets/list`;
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

      // 处理数据，将 extension 中的字段提取到顶层便于展示
      const processedList = list.map((item: ISubAccountSecretItem) => ({
        ...item,
        cloud_secret_id: item.extension?.cloud_secret_id,
        cloud_main_account_id: item.extension?.cloud_main_account_id,
        cloud_sub_account_id: item.extension?.cloud_sub_account_id,
        console_login: item.extension?.console_login,
      }));

      return { list: processedList, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      subAccountSecretListLoading.value = false;
    }
  };

  /**
   * 启用或禁用三级账号密钥（创建申请）
   * @param bk_biz_id 业务ID
   * @param vendor 云厂商
   * @param params 密钥状态更新参数列表
   */
  const updateSubAccountSecretStatus = async (
    bk_biz_id: number,
    vendor: string,
    params: IUpdateSecretStatusParams[],
  ): Promise<{ ids: string[] }> => {
    // 使用 Mock 数据
    if (USE_MOCK) {
      return mockUpdateSubAccountSecretStatus(params);
    }

    // 使用真实接口
    try {
      const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/applications/types/update_sub_account_secret_status`;
      const res = await http.post(api, {
        sub_account_secrets: params,
      });
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 删除三级账号密钥（创建申请）
   * @param bk_biz_id 业务ID
   * @param vendor 云厂商
   * @param ids 密钥ID列表
   */
  const deleteSubAccountSecret = async (
    bk_biz_id: number,
    vendor: string,
    ids: string[],
  ): Promise<{ ids: string[] }> => {
    // 使用 Mock 数据
    if (USE_MOCK) {
      return mockDeleteSubAccountSecret(ids);
    }

    // 使用真实接口
    try {
      const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/applications/types/delete_sub_account_secret`;
      const res = await http.post(api, { ids });
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  return {
    accountListLoading,
    secretListLoading,
    secretCheckLoading,
    subAccountSecretListLoading,
    getSecondaryAccountList,
    getSecondaryAccountFullList,
    syncAccountResource,
    syncSecondaryAccounts,
    getAccountSecretList,
    createAccountSecret,
    updateAccountSecret,
    deleteAccountSecret,
    checkAccountSecret,
    createSecondaryAccount,
    updateSecondaryAccount,
    getSubAccountSecretList,
    updateSubAccountSecretStatus,
    deleteSubAccountSecret,
  };
});
