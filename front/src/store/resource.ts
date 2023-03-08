import http from '@/http';
import { defineStore } from 'pinia';

import { useAccountStore } from '@/store';
// import { json2Query } from '@/common/util';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
// 获取
const getBusinessApiPath = () => {
  const store = useAccountStore();
  if (location.href.includes('business')) {
    return `bizs/${store.bizs}/`
  } else {
    return ''
  }
}

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
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}${type}/list`, data);
    },
    detail(type: string, id: number | string, vendor?: string) {
      if (vendor) {
        return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/${vendor}/${type}/${id}`);
      }
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}${type}/${id}`);
    },
    delete(type: string, id: string | number) {
      return http.delete(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}${type}/${id}`);
    },
    deleteBatch(type: string, data: any) {
      console.log('dataqq11', data);
      return http.delete(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}${type}/batch`, { data });
    },
    bindVPCWithCloudArea(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vpcs/bind/cloud_areas`, data);
    },
    getCloudAreas(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/cloud_areas/list`, data);
    },
    getRouteList(type: string, id: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/${type}/route_tables/${id}/routes/list`, data);
    },
    // 分配到业务下
    assignBusiness(type: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}${type}/assign/bizs`, data);
    },
    // 新增
    add(type: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}${type}`, data);
    },
    // 更新
    update(type: string, data: any, id: string | number) {
      return http.put(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}${type}/${id}`, data);
    },
    // 获取
    countSubnetIps(id: string | number) {
      return http.put(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}subnets/${id}/ips/count`);
    },
    getEipListByCvmId(vendor: string, id: string) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/${vendor}/eips/cvm/${id}`);
    },
    getDiskListByCvmId(vendor: string, id: string) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/${vendor}/disks/cvms/${id}`);
    },
    // 获取根据主机安全组列表
    getSecurityGroupsListByCvmId(id: string) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}security_groups/cvms/${id}`);
    },
    // 获取根据主机安全组列表
    cvmOperate(type: string, data: {ids: string[]}) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}cvms/batch/${type}`, data);
    },
    // 主机分配
    cvmAssignBizs(data: {cvm_ids: string[], bk_biz_id: string}) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}cvms/assign/bizs`, data);
    },
    // 网络接口
    cvmNetwork(type: string, id: string) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/${type}/network_interfaces/cvms/${id}`);
    },
    getCommonList(data: any, url: string, methodType?: string) {
      if (!methodType || methodType === 'post') {
        return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}${url}`, data);
      }
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}${url}`);
    },
    getNetworkList(type: string, id: string) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/${type}/network_interfaces/cvms/${id}`);
    },
    attachDisk(vendor: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/vendors/${vendor}/disks/attach`, data);
    },
    detachDisk(vendor: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/vendors/${vendor}/disks/detach`, data);
    },
    associateEip(vendor: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/vendors/${vendor}/eips/associate`, data);
    },
    disassociateEip(vendor: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/vendors/${vendor}/eips/disassociate`, data);
    },
  },
});
