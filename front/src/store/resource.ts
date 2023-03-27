import http from '@/http';
import { defineStore } from 'pinia';

import { useAccountStore } from '@/store';
// import { json2Query } from '@/common/util';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
// 获取
const getBusinessApiPath = (type?: string) => {
  const store = useAccountStore();
  if (location.href.includes('business') && type !== 'images') {
    return `bizs/${store.bizs}/`;
  }
  return '';
};

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
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}${type}/list`, data);
    },
    detail(type: string, id: number | string, vendor?: string) {
      if (vendor) {
        return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}vendors/${vendor}/${type}/${id}`);
      }
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}${type}/${id}`);
    },
    delete(type: string, id: string | number) {
      return http.delete(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}${type}/${id}`);
    },
    deleteBatch(type: string, data: any) {
      return http.delete(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}${type}/batch`, { data });
    },
    recyclBatch(type: string, data: any) {
      return http.delete(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/recycled/${getBusinessApiPath(type)}${type}/batch`, { data });
    },
    bindVPCWithCloudArea(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vpcs/bind/cloud_areas`, data);
    },
    getCloudAreas(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/cloud_areas/list`, data);
    },
    getRouteList(type: string, id: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}vendors/${type}/route_tables/${id}/routes/list`, data);
    },
    // 分配到业务下
    assignBusiness(type: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}${type}/assign/bizs`, data);
    },
    // 新增
    add(type: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}${type}`, data);
    },
    // 更新
    update(type: string, data: any, id: string | number) {
      return http.put(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}${type}/${id}`, data);
    },
    // 获取
    countSubnetIps(id: string | number) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}subnets/${id}/ips/count`);
    },
    getEipListByCvmId(vendor: string, id: string) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/${vendor}/eips/cvms/${id}`);
    },
    getDiskListByCvmId(vendor: string, id: string) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/${vendor}/disks/cvms/${id}`);
    },
    // 获取根据主机安全组列表
    getSecurityGroupsListByCvmId(id: string) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}security_groups/cvms/${id}`);
    },
    // 操作主机相关
    cvmOperate(type: string, data: {ids: string[]}) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}cvms/batch/${type}`, data);
    },
    // 主机分配
    cvmAssignBizs(data: {cvm_ids: string[], bk_biz_id: string}) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}cvms/assign/bizs`, data);
    },
    // 网络接口
    cvmNetwork(type: string, id: string) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}vendors/${type}/network_interfaces/cvms/${id}`);
    },
    getCommonList(data: any, url: string) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}${url}`, data);
    },
    getNetworkList(type: string, id: string) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}vendors/${type}/network_interfaces/cvms/${id}`);
    },
    attachDisk(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/disks/attach`, data);
    },
    detachDisk(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath('disks')}disks/detach`, data);
    },
    associateEip(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath('eips')}eips/associate`, data);
    },
    disassociateEip(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath('eips')}eips/disassociate`, data);
    },
    getCloudRegion(type: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/${type}/regions/list`, data);
    },
    // 销毁
    deleteRecycledData(type: string, data: any) {
      return http.delete(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}recycled/${type}/batch`, { data });
    },
    // 回收
    recoverRecycledData(type: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}${type}/recover`, data);
    },

    // 虚拟机回收
    recycledCvmsData(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath('cvms')}cvms/recycle`, data);
    },

    // 回收资源详情
    recycledResourceDetail(type: string, id: string) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}recycled/${type}/${id}`);
    },
  },
});
