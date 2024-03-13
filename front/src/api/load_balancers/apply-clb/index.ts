/**
 * 负载均衡 - 购买
 */
// import http
import http from '@/http';
// import types
import { NetworkAccountTypeResp, ResourceOfCurrentRegionReqData, ResourceOfCurrentRegionResp } from './types';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

// 查询用户网络类型
// https://github.com/TencentBlueKing/bk-hcm/blob/feature-loadbalancer/docs/api-docs/web-server/docs/resource/clb/tcloud_describe_network_account_type.md
export const reqAccountNetworkType = async (account_id: string): Promise<NetworkAccountTypeResp> => {
  return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/tcloud/accounts/${account_id}/network_type`);
};

// 查询用户在当前地域支持可用区列表和资源列表
// https://github.com/TencentBlueKing/bk-hcm/blob/feature-loadbalancer/docs/api-docs/web-server/docs/resource/clb/tcloud_describe_resources.md
export const reqResourceListOfCurrentRegion = async (
  data: ResourceOfCurrentRegionReqData,
): Promise<ResourceOfCurrentRegionResp> => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/tcloud/load_balancers/resources/describe`, data);
};
