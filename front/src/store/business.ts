import http from '@/http';
import { defineStore } from 'pinia';

import { useAccountStore } from '@/store';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
// 获取
const getBusinessApiPath = () => {
  const store = useAccountStore();
  if (location.href.includes('business')) {
    return `bizs/${store.bizs}/`;
  }
  return '';
};

export const useBusinessStore = defineStore({
  id: 'businessStore',
  state: () => ({
  }),
  actions: {
    /**
     * @description: 新增安全组
     * @param {any} data
     * @return {*}
     */
    addSecurity(data: any, isRes = false) {
      if (isRes) return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/security_groups/create`, data);
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}security_groups/create`, data);
    },
    addEip(id: number, data: any, isRes = false) {
      if (isRes) return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/eips/create`, data);
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/bizs/${id}/eips/create`, data);
    },
    /**
     * @description: 获取可用区列表
     * @param {any} data
     * @return {*}
     */
    getZonesList({ vendor, region, data }: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/${vendor}/regions/${region}/zones/list`, data);
    },
    /**
     * @description: 获取资源组列表
     * @param {any} data
     * @return {*}
     */
    getResourceGroupList(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/azure/resource_groups/list`, data);
    },
    /**
     * @description: 获取路由表列表
     * @param {any} data
     * @return {*}
     */
    getRouteTableList(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}route_tables/list`, data);
    },
    /**
     * @description: 创建子网
     * @param {any} data
     * @return {*}
     */
    createSubnet(bizs: number | string, data: any, isRes = false) {
      if (isRes) return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/subnets/create`, data);
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/bizs/${bizs}/subnets/create`, data);
    },
  },
});
