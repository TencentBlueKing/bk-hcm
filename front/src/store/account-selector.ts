import { ref } from 'vue';
import { defineStore } from 'pinia';
import rollRequest from '@blueking/roll-request';
import http from '@/http';
import { IAccountItem, QueryBuilderType } from '@/typings';

export const useAccountSelectorStore = defineStore('account-selector', () => {
  const businessAccountList = ref<IAccountItem[]>([]);
  const resourceAccountList = ref<IAccountItem[]>([]);
  const authorizedResourceAccountList = ref<IAccountItem[]>([]);
  const businessAccountLoading = ref(false);
  const resourceAccountLoading = ref(false);
  const authorizedResourceAccountLoading = ref(false);

  const getBusinessAccountList = async (params: { bizId: number; account_type: 'resource' }) => {
    const { bizId, ...query } = params;
    businessAccountLoading.value = true;
    try {
      const { data: list = [] } = await http.get(`/api/v1/cloud/accounts/bizs/${bizId}`, { params: query });
      businessAccountList.value = list;

      return list;
    } catch {
      businessAccountList.value = [];
    } finally {
      businessAccountLoading.value = false;
    }
  };

  const getResourceAccountList = async (data: QueryBuilderType) => {
    resourceAccountLoading.value = true;
    try {
      const list = await rollRequest({
        httpClient: http,
        pageEnableCountKey: 'count',
      }).rollReqUseCount<IAccountItem>('/api/v1/cloud/accounts/resources/accounts/list', data, {
        limit: 500,
        countGetter: (res) => res.data.count,
        listGetter: (res) => res.data.details,
      });

      resourceAccountList.value = list as IAccountItem[];

      return list;
    } catch {
      resourceAccountList.value = [];
    } finally {
      resourceAccountLoading.value = false;
    }
  };

  const getAuthorizedResourceAccountList = async (data: QueryBuilderType) => {
    authorizedResourceAccountLoading.value = true;
    try {
      const list = await rollRequest({
        httpClient: http,
        pageEnableCountKey: 'count',
      }).rollReqUseCount<IAccountItem>('/api/v1/cloud/accounts/list', data, {
        limit: 500,
        countGetter: (res) => res.data.count,
        listGetter: (res) => res.data.details,
      });

      authorizedResourceAccountList.value = list as IAccountItem[];

      return list;
    } catch {
      authorizedResourceAccountList.value = [];
    } finally {
      authorizedResourceAccountLoading.value = false;
    }
  };

  return {
    businessAccountList,
    resourceAccountList,
    authorizedResourceAccountList,
    businessAccountLoading,
    resourceAccountLoading,
    authorizedResourceAccountLoading,
    getBusinessAccountList,
    getResourceAccountList,
    getAuthorizedResourceAccountList,
  };
});
