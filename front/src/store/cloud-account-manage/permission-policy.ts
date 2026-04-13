import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { IListResData, QueryFilterType, QueryRuleOPEnum } from '@/typings';
import rollRequest from '@blueking/roll-request';

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

// 新增/编辑权限列表参数
export interface IOperationPermissionPolicyParams {
  id?: string;
  name: string;
  policy_document: string;
  bk_biz_ids: number[];
  memo: string;
}

export const usePermissionPolicyStore = defineStore('permissionPolicy', () => {
  const permissionPolicyListLoading = ref(false);

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
   * @param bk_biz_id 业务ID
   * @param account_id 账号ID
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
   * @param bk_biz_id 业务ID
   * @param vendor 云厂商
   * @param params 查询参数
   */
  const getPermissionPolicyList = async (
    bk_biz_id: number,
    vendor: string,
    params: { filter?: any; page: any } & Record<string, any>,
  ): Promise<{ list: IPermissionPolicyItem[]; count: number }> => {
    permissionPolicyListLoading.value = true;

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

      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      permissionPolicyListLoading.value = false;
    }
  };

  /**
   * 使用 rollRequest 获取权限策略库全量列表（用于前端分页）
   * @param vendor 云账户
   * @param filter 过滤条件
   * @param onProgress 进度回调，每批次数据返回时调用
   */
  const getPermissionPolicyFullList = async (
    vendor: string,
    filter: QueryFilterType,
    onProgress?: (list: IPermissionPolicyItem[], count: number) => void,
  ): Promise<IPermissionPolicyItem[]> => {
    permissionPolicyListLoading.value = true;
    const api = `/api/v1/cloud/vendors/${vendor}/permission_policy_libraries/list`;
    const allList: IPermissionPolicyItem[] = [];

    try {
      const listGen = await rollRequest({ httpClient: http, pageEnableCountKey: 'count' }).rollReqUseCount<
        IListResData<IPermissionPolicyItem[]>
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
        permissionPolicyListLoading.value = false;
      }

      return allList;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      permissionPolicyListLoading.value = false;
    }
  };

  const createPolicyLibraryListGenerator = (vendor: string) => {
    return async function* (keyword?: string) {
      const api = `/api/v1/cloud/vendors/${vendor}/permission_policy_libraries/list`;
      const filter = keyword
        ? { op: QueryRuleOPEnum.AND, rules: [{ field: 'name', op: QueryRuleOPEnum.CS, value: keyword }] }
        : {};

      const gen = await rollRequest({ httpClient: http, pageEnableCountKey: 'count' }).rollReqUseCount<
        IListResData<IPermissionPolicyItem[]>
      >(
        api,
        { filter },
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
    createPermissionPolicy,
    updatePermissionPolicy,
    getPermissionPolicyList,
    getPermissionPolicyFullList,
    createPolicyLibraryListGenerator,
  };
});
