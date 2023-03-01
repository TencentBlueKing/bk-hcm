import http from '@/http';
import { defineStore } from 'pinia';
// import { json2Query } from '@/common/util';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export const useResourceStore = defineStore({
  id: 'resourceStore',
  state: () => ({
    securityRuleDetail: {},
  }),
  actions: {
    setSecurityRuleDetail(data: any) {
      this.securityRuleDetail = data;
    },
    /**
     * @description: 获取资源列表
     * @param {any} data
     * @param {string} type
     * @return {*}
     */
    list(data: any, type: string) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${type}/list`, data);
    },
    detail(type: string, id: number | string, vendor?: string) {
      if (vendor) {
        return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/${vendor}/${type}/${id}`);
      }
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${type}/${id}`);
    },
    delete(type: string, id: string | number) {
      return http.delete(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${type}/${id}`);
    },
    deleteBatch(type: string, data: any) {
      console.log('dataqq11', data);
      return http.delete(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${type}/batch`, { data });
    },
    bindVPCWithCloudArea(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vpcs/bind/cloud_areas`, data);
    },
    getCloudAreas(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/cloud_areas/list`, data);
    },
    getRouteList(type: string, id: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/${type}/route_tables/${id}/routes/list`, data);
    },
    // 分配到业务下
    assignBusiness(type: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${type}/assign/bizs`, data);
    },
    // 新增
    add(type: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${type}`, data);
    },
    // 更新
    update(type: string, data: any, id: string | number) {
      return http.put(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${type}/${id}`, data);
    },
    // 获取
    countSubnetIps(id: string | number) {
      return http.put(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/subnets/${id}/ips/count`);
    },
    getEipListByCvmId(vendor: string, id: string) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/${vendor}/eips/cvm/${id}`);
    },
    getDiskListByCvmId(vendor: string, id: string) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/${vendor}/disks/cvms/${id}`);
    }
  },
});
