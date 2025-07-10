import http from '@/http';
import { defineStore } from 'pinia';
import rollRequest from '@blueking/roll-request';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { VendorEnum } from '@/common/constant';
import { FilterType } from '@/typings';
// import { json2Query } from '@/common/util';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
export interface GetAllSortParams {
  vendor: string | string[];
  id: string;
  filter: FilterType;
}
export interface SyncResourceParams {
  regions?: string[];
  cloud_ids?: string[];
  tag_filters?: Record<string, string[]>;
  resource_group_names?: string[]; // azure
}
export interface BatchBindSecurityInfoParams {
  cvm_id: string;
  security_group_ids: string[];
}

// 获取
const getBusinessApiPath = (type?: string) => {
  const { getBizsId } = useWhereAmI();
  if (location.href.includes('business') && type !== 'images') {
    return `bizs/${getBizsId()}/`;
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
    // 更新安全组规则排序
    updateRulesSort(data: any, type: string, id: string) {
      return http.put(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(
          type,
        )}vendors/${type}/security_groups/${id}/rules/batch/update`,
        data,
      );
    },
    // 安全组规则排序获取所有规则
    getAllSort(params: GetAllSortParams) {
      const { vendor, id, filter } = params;
      const fetchUrl = `vendors/${vendor}/security_groups/${id}/rules/list`;
      const list = rollRequest({
        httpClient: http,
        pageEnableCountKey: 'count',
      }).rollReqUseCount(
        `api/v1/cloud/${getBusinessApiPath()}${fetchUrl}`,
        {
          filter,
        },
        {
          limit: 500,
          countGetter: (res) => res.data.count,
          listGetter: (res) => res.data.details,
        },
      );
      return list;
    },
    /**
     * @description: 获取资源列表
     * @param {any} data
     * @param {string} type
     * @return {*}
     */
    list(data: any, type: string) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}${type}/list`, data, {
        cancelPrevious: true,
      });
    },
    detail(type: string, id: number | string, vendor?: string) {
      if (vendor) {
        return http.get(
          `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}vendors/${vendor}/${type}/${id}`,
        );
      }
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}${type}/${id}`);
    },
    delete(type: string, id: string | number) {
      return http.delete(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}${type}/${id}`);
    },
    deleteBatch(type: string, data: any, config?: any) {
      return http.delete(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}${type}/batch`,
        { data },
        config,
      );
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
    getAllCloudAreas() {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/all/cloud_areas/list`);
    },
    getRouteList(type: string, id: string, data: any) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(
          type,
        )}vendors/${type}/route_tables/${id}/routes/list`,
        data,
      );
    },
    // 分配到业务下
    assignBusiness(type: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}${type}/assign/bizs`, data);
    },
    // 新增
    add(type: string, data: any, config?: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}${type}`, data, config);
    },
    // 更新
    update(type: string, data: any, id: string | number, config?: any) {
      return http.put(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}${type}/${id}`, data, config);
    },
    // 获取
    countSubnetIps(id: string | number) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}subnets/${id}/ips/count`);
    },
    getEipListByCvmId(vendor: string, id: string) {
      return http.get(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/${vendor}/eips/cvms/${id}`,
      );
    },
    getDiskListByCvmId(vendor: string, id: string) {
      return http.get(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/${vendor}/disks/cvms/${id}`,
      );
    },
    // 获取根据主机安全组列表
    getSecurityGroupsListByCvmId(id: string) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}security_groups/res/cvm/${id}`);
    },
    // 操作主机相关
    cvmOperate(type: string, data: { ids: string[] }) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}cvms/batch/${type}`, data);
    },
    // 主机分配
    cvmAssignBizs(data: { cvm_ids: string[]; bk_biz_id: string }) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}cvms/assign/bizs`, data);
    },
    // 网络接口
    cvmNetwork(type: string, id: string) {
      return http.get(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(
          type,
        )}vendors/${type}/network_interfaces/cvms/${id}`,
      );
    },
    getCommonList(data: any, url: string) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}${url}`, data);
    },
    getNetworkList(type: string, id: string) {
      return http.get(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(
          type,
        )}vendors/${type}/network_interfaces/cvms/${id}`,
      );
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
      return http.delete(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}recycled/${type}/batch`, {
        data,
      });
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
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}security_groups/associate/${type}`,
        data,
      );
    },

    // 绑定主机安全组信息（主机批量关联安全组(仅支持: tcloud、aws)）
    batchBindSecurityInfo(params: BatchBindSecurityInfoParams) {
      const { cvm_id, security_group_ids } = params;
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}cvms/${cvm_id}/security_groups/batch_associate`,
        { security_group_ids },
      );
    },
    // 解绑主机安全组信息
    unBindSecurityInfo(type: string, data: any) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}security_groups/disassociate/${type}`,
        data,
      );
    },

    // 获取未绑定eip的网络接口列表
    getUnbindEipNetworkList(data: any) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}network_interfaces/associate/list`,
        data,
      );
    },

    // 获取未绑定disk的主机列表
    getUnbindDiskCvms(data: any) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}disk_cvm_rels/with/cvms/list`,
        data,
      );
    },

    // 获取未绑定主机的disk列表
    getUnbindCvmDisks(data: any) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}disk_cvm_rels/with/disks/without/cvm/list`,
        data,
      );
    },

    // 获取未绑定主机的eips列表
    getUnbindCvmEips(data: any) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}eip_cvm_rels/with/eips/without/cvm/list`,
        data,
      );
    },
    // 创建
    create(type: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath(type)}${type}/create`, data);
    },
    // 同步拉取资源
    syncResource(vendor: string, accountId: string, resourceName: string, params: SyncResourceParams, config?: any) {
      return http.post(
        `/api/v1/cloud/${getBusinessApiPath()}vendors/${vendor}/accounts/${accountId}/resources/${resourceName}/sync_by_cond`,
        params,
        config,
      );
    },
  },
});
