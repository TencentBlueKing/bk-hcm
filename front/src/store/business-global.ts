import { ref } from 'vue';
import { defineStore } from 'pinia';
import { IQueryResData } from '@/typings';
import http from '@/http';

export interface IBusinessItem {
  id: number;
  name: string;
}

export const useBusinessGlobalStore = defineStore('businessGlobal', () => {
  const businessFullList = ref<IBusinessItem[]>([]);
  const businessAuthorizedList = ref<IBusinessItem[]>([]);
  const businessFullListLoading = ref(false);
  const businessAuthorizedListLoading = ref(false);

  const getFullBusiness = async () => {
    businessFullListLoading.value = true;
    try {
      const { data: list = [] }: IQueryResData<IBusinessItem[]> = await http.post('/api/v1/web/bk_bizs/list');
      businessFullList.value = list;
      return list;
    } finally {
      businessFullListLoading.value = false;
    }
  };

  const getAuthorizedBusiness = async () => {
    businessAuthorizedListLoading.value = true;
    try {
      const { data: list = [] }: IQueryResData<IBusinessItem[]> = await http.post('/api/v1/web/authorized/bizs/list');
      businessAuthorizedList.value = list;
      return list;
    } finally {
      businessAuthorizedListLoading.value = false;
    }
  };

  const getBusinessFullList = async () => {
    if (businessFullList.value.length > 0) {
      return businessFullList.value;
    }
    const list = await getFullBusiness();
    return list;
  };

  const getBusinessAuthorizedList = async () => {
    if (businessAuthorizedList.value.length > 0) {
      return businessAuthorizedList.value;
    }
    const list = await getAuthorizedBusiness();
    return list;
  };

  const getFirstBizId = async () => {
    const list = await getBusinessFullList();
    return list?.[0]?.id;
  };

  const getFirstAuthorizedBizId = async () => {
    const list = await getBusinessAuthorizedList();
    return list?.[0]?.id;
  };

  const getBusinessNames = (id: IBusinessItem['id'] | IBusinessItem['id'][]) => {
    const ids = Array.isArray(id) ? id : [id];
    const names = [];
    for (const value of ids) {
      const name = businessFullList.value.find((item) => item.id === value)?.name;
      names.push(name);
    }
    return names;
  };

  const getBusinessIds = (name: IBusinessItem['name'] | IBusinessItem['name'][]) => {
    const names = Array.isArray(name) ? name : [name];
    const ids = [];
    for (const value of names) {
      const id = businessFullList.value.find((item) => item.name === value)?.id;
      ids.push(id);
    }
    return ids;
  };

  const getCacheSelected = (key: string) => {
    if (localStorage.getItem(key)) {
      const cacheValue = JSON.parse(localStorage.getItem(key));
      return cacheValue;
    }
  };

  return {
    businessFullList,
    businessAuthorizedList,
    businessFullListLoading,
    businessAuthorizedListLoading,
    getFullBusiness,
    getAuthorizedBusiness,
    getBusinessFullList,
    getBusinessAuthorizedList,
    getFirstBizId,
    getFirstAuthorizedBizId,
    getBusinessNames,
    getBusinessIds,
    getCacheSelected,
  };
});
