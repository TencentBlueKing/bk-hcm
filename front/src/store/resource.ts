import http from '@/http';
import { defineStore } from 'pinia';
import { json2Query } from '@/common/util';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

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
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${type}/list/`, data);
      // return http.post(`http://9.135.119.6:9602/api/v1/cloud/${type}/list/`, data);
    },
    detail(type: string, id: number | string) {
      return http.get(`/api/v1/cloud/${type}/${id}/`);
    },
    delete(type: string, id: string | number) {
      return http.delete(`/api/v1/cloud/${type}/${id}`);
    },
    deleteBatch(type: string, data: any) {
      return http.delete(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${type}/batch?${json2Query(data)}`);
    },
    bindVPCWithCloudArea(data: any) {
      return http.post('/api/v1/cloud/vpc/bind/cloud_area', data);
    },
    // 分配到业务下
    assignBusiness(type: string, data: any) {
      return http.post(`/api/v1/cloud/${type}/assign/bizs`, data);
    },
    // 新增
    add(type: string, data: any) {
      return http.post(`/api/v1/cloud/${type}`, data);
    },
    // 更新
    update(type: string, data: any, id: string | number) {
      return http.patch(`/api/v1/cloud/${type}/${id}`, data);
    },
  },
});
