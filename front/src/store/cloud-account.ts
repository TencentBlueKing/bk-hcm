import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { IListResData, QueryBuilderType, QueryFilterType } from '@/typings';
import { VendorEnum } from '@/common/constant';
import { QueryRuleOPEnum } from '@/typings/common';
import { enableCount } from '@/utils/search';
import rollRequest from '@blueking/roll-request';
import type { ICloudSecretItem } from '@/views/cloud-account-manage/cloud-secret/typings';

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

// ISubAccountSecretItem 已统一为 ICloudSecretItem（来自 cloud-secret/typings.ts），两者共用同一套接口
// 保留别名以保持向后兼容
export { type ICloudSecretItem as ISubAccountSecretItem } from '@/views/cloud-account-manage/cloud-secret/typings';

// 三级账号项接口定义
export interface ISubAccountItem {
  id: string;
  cloud_id: string;
  name: string;
  vendor: string;
  site: string;
  account_id: string;
  managers: string[];
  bk_biz_ids: number[];
  memo: string;
  email: string;
  phone_num: string;
  country_code: string;
  cloud_created_at: string;
  sub_account_secret_count: number;
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
  operable?: boolean;
  [k: string]: any;
}

// 创建三级账号参数
export interface ISubAccountCreateParams {
  account_id: string;
  name: string;
  receive_email: string;
  email?: string;
  phone_num?: string;
  country_code?: string;
  managers?: string[];
  memo?: string;
  extension: {
    console_login: number; // 0=编程账号，1=控制台账号
  };
}

// 更新三级账号参数
export interface ISubAccountUpdateParams {
  id: string;
  name?: string;
  email?: string;
  phone_num?: string;
  bk_biz_id?: number;
  country_code?: string;
  managers?: string[];
  memo?: string;
}

// 更新密钥状态参数
export interface IUpdateSecretStatusParams {
  id: string;
  status: 'enabled' | 'disabled';
}

export interface ISubAccountSecretParams {
  status?: string;
  account_ids?: string[];
  sub_account_ids?: string[];
  account_managers?: string[];
  sub_account_managers?: string[];
  extension?: {
    cloud_secret_ids?: string[];
    cloud_main_account_ids?: string[];
    cloud_sub_account_ids?: string[];
  };
  page: any;
}

export const useCloudAccountStore = defineStore('cloudAccount', () => {
  const accountListLoading = ref(false);
  const secretListLoading = ref(false);
  const secretCheckLoading = ref(false);
  const subAccountSecretListLoading = ref(false);
  const subAccountListLoading = ref(false);

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
   * @param bizId 业务ID
   */
  const getSecondaryAccountListByAccountIds = async (accountIds: string[], bizId: number) => {
    const api = `/api/v1/cloud/bizs/${bizId}/accounts/list`;
    const cachedIds = allSecondaryAccountCacheList.value.keys();
    const cachedIdSet = new Set(cachedIds);
    const newIds = accountIds.filter((id) => !cachedIdSet.has(id));
    if (newIds.length > 0) {
      const list = await rollRequest({
        httpClient: http,
        pageEnableCountKey: 'count',
      }).rollReqUseCount<ISecondaryAccountItem>(
        api,
        {
          filter: { op: QueryRuleOPEnum.AND, rules: [{ field: 'id', op: QueryRuleOPEnum.IN, value: newIds }] },
        },
        {
          limit: 500,
          countGetter: (res) => res.data.count,
          listGetter: (res) => res.data.details,
        },
      );
      for (const item of list) {
        allSecondaryAccountCacheList.value.set(item.id, item);
      }
    }
    return accountIds.map((id) => allSecondaryAccountCacheList.value.get(id)).filter(Boolean);
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
   * 获取三级账号详情（通过列表接口按 id 查询单条）
   */
  const getSubAccountDetail = async (
    bk_biz_id: number,
    vendor: string,
    id: string,
  ): Promise<ISubAccountItem | null> => {
    const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/sub_accounts/list`;
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
   * 获取三级账号密钥列表
   * @param bk_biz_id 业务ID
   * @param vendor 云厂商
   * @param params 查询参数
   */
  const getSubAccountSecretList = async (
    bk_biz_id: number,
    vendor: string,
    params: ISubAccountSecretParams,
  ): Promise<{ list: ICloudSecretItem[]; count: number }> => {
    subAccountSecretListLoading.value = true;

    // 使用真实接口
    const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/sub_account_secrets/list`;
    try {
      // 构建请求参数（去除 page 后的查询条件）
      const { page, ...queryParams } = params;

      // 获取列表数据
      const listRes = await http.post(api, {
        ...queryParams,
        page: { ...page, count: false },
      });

      // 获取总数
      const countRes = await http.post(api, {
        ...queryParams,
        page: { count: true, start: 0, limit: 0 },
      });

      const list = listRes?.data?.details || [];
      const count = countRes?.data?.count || 0;

      // 处理数据，将 extension 中的字段提取到顶层便于展示
      const processedList = list.map((item: ICloudSecretItem) => ({
        ...item,
        ...item.extension,
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
   * 获取三级账号密钥详情（通过 sub_account_secrets/list 接口按 id 查询单条）
   */
  const getSubAccountSecretDetail = async (
    bk_biz_id: number,
    vendor: string,
    id: string,
  ): Promise<ICloudSecretItem | null> => {
    const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/sub_account_secrets/list`;
    try {
      const res = await http.post(api, {
        ids: [id],
        page: { count: false, start: 0, limit: 1 },
      });
      const list = (res?.data as { details: ICloudSecretItem[] })?.details ?? [];
      if (list.length === 0) return null;
      const item = list[0];
      // 将 extension 中的字段提取到顶层，与列表数据处理保持一致
      return { ...item, ...item.extension };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
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
    try {
      const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/applications/types/delete_sub_account_secret`;
      const res = await http.post(api, { ids });
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 获取三级账号全量列表（用于前端分页）
   * @param bk_biz_id 业务ID
   * @param vendor 云厂商
   * @param filter 过滤条件
   * @param onProgress 进度回调
   */
  const getSubAccountFullList = async (
    bk_biz_id: number,
    vendor: string,
    filter: QueryFilterType,
    onProgress?: (list: ISubAccountItem[], count: number) => void,
  ): Promise<ISubAccountItem[]> => {
    subAccountListLoading.value = true;
    const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/sub_accounts/list`;
    const allList: ISubAccountItem[] = [];

    try {
      const listGen = await rollRequest({ httpClient: http, pageEnableCountKey: 'count' }).rollReqUseCount<
        IListResData<ISubAccountItem[]>
      >(
        api,
        { filter },
        {
          limit: 500,
          countGetter: (res) => res.data.count,
          listGetter: (res) => res.data.details,
          generator: true,
        },
        true,
      );

      for await (const res of listGen) {
        const details = res.data?.details || [];
        allList.push(...details);
        onProgress?.(allList, res.data?.count || allList.length);
        subAccountListLoading.value = false;
      }

      return allList;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      subAccountListLoading.value = false;
    }
  };

  /**
   * 创建三级账号（提交申请）
   */
  const createSubAccount = async (
    bk_biz_id: number,
    vendor: string,
    subAccounts: ISubAccountCreateParams[],
  ): Promise<{ ids: string[] }> => {
    try {
      const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/applications/types/add_sub_account`;
      const res = await http.post(api, { sub_accounts: subAccounts });
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 更新三级账号（提交申请）
   */
  const updateSubAccount = async (
    bk_biz_id: number,
    vendor: string,
    subAccounts: ISubAccountUpdateParams[],
  ): Promise<{ ids: string[] }> => {
    try {
      const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/applications/types/update_sub_account`;
      const res = await http.post(api, { sub_accounts: subAccounts });
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 删除三级账号（提交申请）
   */
  const deleteSubAccount = async (bk_biz_id: number, vendor: string, ids: string[]): Promise<{ ids: string[] }> => {
    try {
      const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/applications/types/delete_sub_account`;
      const res = await http.post(api, { ids });
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 新增三级账号密钥
   */
  const createSubAccountSecret = async (
    bk_biz_id: number,
    vendor: string,
    id: string,
  ): Promise<{ id: string; extension: { cloud_secret_id: string; cloud_secret_key: string } }> => {
    try {
      const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/sub_account_secrets/create`;
      const res = await http.post(api, { id });
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 获取三级账号数量（纯计数查询）
   * @param bk_biz_id 业务ID
   * @param vendor 云厂商
   * @param filter 过滤条件
   */
  const getSubAccountCount = async (bk_biz_id: number, vendor: string, filter: QueryFilterType): Promise<number> => {
    try {
      const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/sub_accounts/list`;
      const res = await http.post(api, enableCount({ filter }, true));
      return res?.data?.count ?? 0;
    } catch (error) {
      console.error(error);
      return 0;
    }
  };

  return {
    accountListLoading,
    secretListLoading,
    secretCheckLoading,
    subAccountSecretListLoading,
    subAccountListLoading,
    getSecondaryAccountList,
    getSecondaryAccountDetail,
    getSecondaryAccountFullList,
    getSecondaryAccountListByAccountIds,
    syncAccountResource,
    syncSecondaryAccounts,
    getAccountSecretList,
    getAccountSecretDetail,
    createAccountSecret,
    updateAccountSecret,
    deleteAccountSecret,
    checkAccountSecret,
    createSecondaryAccount,
    updateSecondaryAccount,
    getSubAccountSecretList,
    getSubAccountSecretDetail,
    updateSubAccountSecretStatus,
    deleteSubAccountSecret,
    allSecondaryAccountCacheList,
    getSubAccountDetail,
    getSubAccountFullList,
    getSubAccountCount,
    createSubAccount,
    updateSubAccount,
    deleteSubAccount,
    createSubAccountSecret,
  };
});
