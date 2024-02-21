import http from '@/http';
import { defineStore } from 'pinia';

import { useAccountStore } from '@/store';
import { VendorEnum } from '@/common/constant';
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
    vendorOfCurrentResource: '' as VendorEnum,
  }),
  actions: {
    setSecurityRuleDetail(data: any) {
      this.securityRuleDetail = data;
    },
    setVendorOfCurrentResource(vendorName: VendorEnum) {
      this.vendorOfCurrentResource = vendorName;
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
      return http.delete(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/recycled/${type}/batch`, { data });
    },
    recycled(type: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}${type}/recycle`, data);
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
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath('disks')}disks/attach`, data);
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

    // 主机所关联资源(硬盘, eip)的个数
    getRelResByCvmIds(data: { ids: string[] }) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/cvms/rel_res/batch`, data);
    },

    // 虚拟机回收
    recycledCvmsData(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath('cvms')}cvms/recycle`, data);
    },

    // 回收资源详情
    recycledResourceDetail(type: string, id: string) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}recycled/${type}/${id}`);
    },

    // 获取azure默认数据
    getAzureDefaultList(type: string) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/azure/default/security_groups/rules/${type}`);
    },

    // 更新安全组信息
    updateSecurityInfo(id: string, data: any) {
      return http.patch(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}security_groups/${id}`, data);
    },

    // 绑定主机安全组信息
    bindSecurityInfo(type: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}security_groups/associate/${type}`, data);
    },

    // 解绑主机安全组信息
    unBindSecurityInfo(type: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}security_groups/disassociate/${type}`, data);
    },

    // 获取未绑定eip的网络接口列表
    getUnbindEipNetworkList(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}network_interfaces/associate/list`, data);
    },

    // 获取未绑定disk的主机列表
    getUnbindDiskCvms(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}disk_cvm_rels/with/cvms/list`, data);
    },

    // 获取未绑定主机的disk列表
    getUnbindCvmDisks(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}disk_cvm_rels/with/disks/without/cvm/list`, data);
    },

    // 获取未绑定主机的eips列表
    getUnbindCvmEips(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}eip_cvm_rels/with/eips/without/cvm/list`, data);
    },
    // 创建
    create(type: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}${type}/create`, data);
    },
  },
});
