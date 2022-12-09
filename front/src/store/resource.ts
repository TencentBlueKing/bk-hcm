// @ts-check
import http from '@/http';
import { defineStore } from 'pinia';

export const useResourceStore = defineStore({
  id: 'resourceStore',
  state: () => ({}),
  actions: {
    /**
     * @description: 获取资源列表
     * @param {any} data
     * @param {string} type
     * @return {*}
     */
    list(data: any, type: string) {
      return http.post(`/api/v1/cloud/${type}/list/`, data);
    },
  },
});
