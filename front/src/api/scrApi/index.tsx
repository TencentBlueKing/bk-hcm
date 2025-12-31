import http from '@/http';
import { ICvmDeviceCreateModel } from '@/views/ziyanScr/cvm-model/CreateDevice';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

/**
 * 获取区域列表
 */
const getAreas = async () => {
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/cvm/tablelist/arealist`);
  return data;
};
/**
 * 获取可用区
 * @param {String} vendor 供应方，idc|qcloud
 * @param {Array} region
 * @returns {Promise}
 */
const getZones = async ({ vendor, region = [], isCmdbRegion = false }, config) => {
  const params = {};

  if (vendor === 'qcloud') {
    if (isCmdbRegion) {
      params.cmdb_region_name = region;
    } else {
      params.region = region.length ? region : undefined;
    }
  } else if (vendor === 'idc') {
    params.cmdb_region_name = region;
  }

  const { data } = await http.post(
    `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/findmany/config/${vendor}/zone`,
    params,
    {
      removeEmptyFields: true,
      ...config,
    },
  );
  return data;
};
/**
 * 获取 CVM 机型列表
 * @param  {String} zone 可用区 id
 */
const getCvmTypes = async (zone: string, area: any) => {
  const params = {
    zone,
    area,
  };
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/cvm/tablelist/cvmtype`, { params });
  return data;
};
/**
 * 获取镜像列表
 * @param  {String} area 区域 id
 */
const getImages = async (region: string[]) => {
  const { data } = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/findmany/config/cvm/image`, { region });
  return data;
};
/**
 * 获取数据盘类型列表
 */
const getDiskTypes = async () => {
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/cvm/disktype`);
  return data;
};
const getAntiAffinityLevels = async (resourceType: any, hasZone: any, config: any) => {
  const { data } = await http.post(
    `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/affinity`,
    {
      resource_type: resourceType,
      has_zone: hasZone,
    },
    {
      config,
    },
  );
  return data;
};
/**
 * 获取回收单据中的主机
 * @returns {Promise}
 */
const getRecycleHosts = async (params) => {
  const { data } = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/findmany/recycle/host`, params, {
    transformFields: true,
    removeEmptyFields: true,
  });
  return data;
};
/** 资源回收单据预览 */
const getPreRecycleList = async ({ ips, remark, returnPlan: { cvm, pm, skipConfirm } }) => {
  const { data } = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/preview/recycle/order`, {
    ips,
    remark,
    return_plan: { cvm, pm },
    skip_confirm: skipConfirm,
  });
  return data;
};
const getRecyclableHosts = async ({ bk_biz_id, ips, asset_ids, bk_host_ids }: any) => {
  const { data } = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/findmany/recycle/recyclability`, {
    bk_biz_id,
    ips,
    asset_ids,
    bk_host_ids,
  });
  return data;
};

/** 业务待回收主机列表查询接口 */
const getRecycleList = async ({ bkBizId, page }) => {
  const { data } = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/find/recycle/biz/host`, {
    bk_biz_id: bkBizId,
    page,
  });
  return data;
};
/**
 * 搜索栏业务列表
 */
const getBusinesses = async () => {
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/auth/apps`);
  return data;
};
/**
 * 搜索栏业务列表新增业务跳转URL
 */
const getAuthApplyUrl = async (permission) => {
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/auth/apply`, {
    ...permission,
  });
  return data;
};
/**
 * 搜索栏需求类型列表
 */
const getRequireTypes = async () => {
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/requirement`);
  return data;
};
/**
 * 获取地域列表
 * @param {String} vendor 供应方，idc|qcloud
 * @returns {Promise}
 */
const getRegions = async (vendor: string): Promise<any> => {
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/${vendor}/region`);
  return data;
};

/**
 * 获取 cpu mem disk 可选项
 * @returns {Promise}
 */
const getRestrict = async () => {
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/cvm/devicerestrict`);
  return data;
};
/**
 * 获取物理机机型
 */
const getIDCPMDeviceTypes = async () => {
  const { data } = await http.post(
    `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/findmany/config/idcpm/devicetype`,
    { page: {} },
    { simpleConditions: true },
  );
  return data;
};
/**
 * 获取物理机操作系统
 */
const getOsTypes = async () => {
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/idcpm/ostype`);
  return data;
};
/**
 * 获取物理机运营商
 */
const getIsps = async () => {
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/idcpm/isp`);
  return data;
};
// VPC列表
const getVpcs = async (region: any) => {
  const { data } = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/findmany/config/cvm/vpc`, {
    region,
  });
  return data;
};
// 子网列表
const getSubnets = async ({ region, zone, vpc }) => {
  const { data } = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/findmany/config/cvm/subnet`, {
    region,
    zone,
    vpc,
  });
  return data;
};
const updateCvmDeviceTypeConfigs = async ({ ids, properties }) => {
  const { data } = await http.put(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/updatemany/config/cvm/device/property`, {
    ids,
    properties,
  });
  return data;
};
/**
 * CVM机型配置信息创建接口
 * @returns {Promise}
 */
const createCvmDevice = async (params: ICvmDeviceCreateModel & { force_create: boolean }, config?: any) => {
  const res = await http.post(
    `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/createmany/config/cvm/device`,
    params,
    config,
  );
  return res;
};
/** 资源回收单据执行接口 */
const startRecycleList = async (suborder_id: string[]) => {
  const data = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/start/recycle/order`, {
    suborder_id,
  });
  return data;
};
/**
 * 获取申请单据详情
 */
const getOrderDetail = async (orderId) => {
  const { data } = await http.post(
    `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/get/apply/ticket`,
    {
      order_id: orderId,
    },
    {
      transformFields: true,
    },
  );
  return data;
};

/**
 * 获取申请单据列表
 */
const getOrders = async ({ bk_biz_id, suborder_id }) => {
  const { data } = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/findmany/apply`, {
    bk_biz_id,
    suborder_id,
  });
  return data;
};
/**
 * cpu、内存默认值的查询接口
 * @param {*} param0
 * @param {*} config
 * @returns
 */
const getAvailDevices = async ({ filter, page }) => {
  const rules = [
    filter.region?.length && { field: 'region', operator: 'in', value: filter.region },
    filter.zone?.length && { field: 'zone', operator: 'in', value: filter.zone },
    filter.require_type && { field: 'require_type', operator: 'equal', value: filter.require_type },
    filter.device_type && { field: 'device_type', operator: 'in', value: filter.device_type },
    filter.cpu && { field: 'cpu', operator: 'equal', value: filter.cpu },
    filter.mem && { field: 'mem', operator: 'equal', value: filter.mem },
    filter.disk && { field: 'disk', operator: 'equal', value: filter.disk },
    filter.device_group && {
      field: 'label.device_group',
      operator: typeof filter.device_group === 'string' ? '=' : 'in',
      value: filter.device_group,
    },
  ].filter(Boolean);
  const { data } = await http.post(
    `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/findmany/config/cvm/device/detail/avail`,
    {
      filter: {
        condition: 'AND',
        rules,
      },
      page: {
        start: page.start,
        limit: page.limit,
        sort: page.sort,
      },
    },
    {
      simpleConditions: true,
    },
  );
  return data;
};
/**
 * 获取资源最大申领量
 * @param {Object} params 参数
 * @param {String} params.bk_biz_id CC 业务 ID
 * @param {String} params.require_type 申领类型
 * @param {String} params.region 地域
 * @param {String} params.zone 园区
 * @param {String} params.vpc VPC
 * @param {String} params.subnet 子网
 * @param {String} params.charge_type 计费模式
 * @returns {Promise}
 */
const getCapacity = async ({ require_type, region, zone, device_type, vpc, subnet, charge_type }) => {
  const { data } = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/cvm/capacity`, {
    require_type,
    region,
    zone,
    device_type,
    vpc,
    subnet,
    charge_type,
  });
  return data;
};

// 主机申请业务列表拉取接口
const getCvmApplyAuthBizList = async () => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/authorized/cvm/apply/bizs/list`);
};

// 主机回收业务列表拉取接口
const getCvmRecycleAuthBizList = async () => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/authorized/cvm/recycle/bizs/list`);
};

export default {
  getCapacity,
  getAreas,
  getZones,
  getCvmTypes,
  getImages,
  getDiskTypes,
  getAntiAffinityLevels,
  getRecyclableHosts,
  getRecycleList,
  getBusinesses,
  getAuthApplyUrl,
  getRequireTypes,
  getRegions,
  getVpcs,
  getSubnets,
  getIDCPMDeviceTypes,
  getOsTypes,
  getIsps,
  getRestrict,
  getRecycleHosts,
  getPreRecycleList,
  updateCvmDeviceTypeConfigs,
  getOrderDetail,
  startRecycleList,
  createCvmDevice,
  getOrders,
  getAvailDevices,
  getCvmApplyAuthBizList,
  getCvmRecycleAuthBizList,
};
