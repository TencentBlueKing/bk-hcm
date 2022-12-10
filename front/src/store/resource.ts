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
    detail(type: string, id: number | string) {
      return http.get(`/api/v1/cloud/${type}/${id}/`);
    },
    delete(type: string, id: string | number) {
      return http.delete(`/api/v1/cloud/${type}/${id}`);
    },
    bindVPCWithCloudArea(data: any) {
      return http.post('/api/v1/cloud/vpc/bind/cloud_area', data);
    },
  },
});
