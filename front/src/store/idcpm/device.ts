import { ref } from 'vue';
import { defineStore } from 'pinia';
import rollRequest from '@blueking/roll-request';
import http from '@/http';
import { QueryBuilderType, QueryRuleOPEnum } from '@/typings';

export interface IIdcpmDevicetypeItem {
  id: string;
  device_type: string;
  cpu_core: number; // 原 cpu
  memory: number; // 原 mem
  raid: string;
  [k: string]: any;
}

export const useIdcpmDeviceStore = defineStore('idcpm-device', () => {
  const deviceTypeFullListLoading = ref(false);

  const getDeviceTypeFullList = async (params?: QueryBuilderType) => {
    deviceTypeFullListLoading.value = true;
    try {
      // 默认添加 disable: false 条件
      const defaultFilter = {
        op: QueryRuleOPEnum.AND,
        rules: [{ field: 'disable', op: QueryRuleOPEnum.EQ, value: false }],
      };

      const finalParams = params?.filter ? params : { ...params, filter: defaultFilter };

      const list = await rollRequest({
        httpClient: http,
        pageEnableCountKey: 'count',
      }).rollReqUseCount<IIdcpmDevicetypeItem>('/api/v1/woa/config/findmany/config/idcpm/devicetype', finalParams, {
        limit: 500,
        countGetter: (res) => res.data.count,
        listGetter: (res) => res.data.info,
      });
      return { list, count: list.length };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      deviceTypeFullListLoading.value = false;
    }
  };

  return {
    deviceTypeFullListLoading,
    getDeviceTypeFullList,
  };
});
