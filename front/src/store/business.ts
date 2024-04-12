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
  state: () => ({}),
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
    /**
     * 获取当前CLB绑定的安全组列表
     */
    async listCLBSecurityGroups(clb_id: string) {
      return http.get(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}security_groups/res/load_balancer/${clb_id}`,
      );
    },
    /**
     * 给当前负载均衡绑定安全组
     */
    async bindSecurityToCLB(data: { bk_biz_id: number; lb_id: string; security_group_ids: Array<string> }) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}security_groups/associate/load_balancers`,
        data,
      );
    },
    /**
     * 给当前负载均衡解绑指定的安全组
     */
    async unbindSecurityToCLB(data: { bk_biz_id: number; lb_id: string; security_group_id: string }) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}security_groups/disassociate/load_balancers`,
        data,
      );
    },
    /*
     * 新建目标组
     */
    createTargetGroups(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}target_groups/create`, data);
    },
    /**
     * 获取目标组详情（基本信息和健康检查）
     */
    getTargetGroupDetail(id: string) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}target_groups/${id}`);
    },
    /**
     * 批量删除目标组
     */
    deleteTargetGroups(data: { bk_biz_id: number; ids: string[] }) {
      return http.delete(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}target_groups/batch`, { data });
    },
    /**
     * 编辑目标组基本信息
     */
    editTargetGroups(data: any) {
      return http.patch(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}target_groups/${data.id}`, data);
    },
    /**
     * 目标组绑定RS列表
     */
    getRsList(
      tg_id: string,
      data: {
        bk_biz_id: string;
        tg_id: string;
        filter: Object;
        page: Object;
      },
    ) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}target_groups/${tg_id}/targets/list`,
        data,
      );
    },
    /**
     * 查询全量的RS列表
     */
    getAllRsList(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}cvms/list`, data);
    },
    /*
     * 业务下腾讯云监听器域名列表
     * @param id 监听器ID
     * @returns 域名列表
     */
    getDomainListByListenerId(id: string) {
      return http.post(`
        ${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/tcloud/listeners/${id}/domains/list
      `);
    },
    /**
     * 获取负载均衡基本信息
     */
    getLbDetail(id: string) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}load_balancers/${id}`);
    },
    /**
     * 更新负载均衡
     */
    updateLbDetail(data: {
      bk_biz_id?: string;
      id: string; // 负载均衡ID
      name?: string; // 名字
      internet_charge_type?: string; // 计费模式
      internet_max_bandwidth_out?: string; // 最大出带宽
      delete_protect?: boolean; // 删除
      load_balancer_pass_to_target?: boolean; // Target是否放通来自CLB的流量
      memo?: string; // 备注
    }) {
      return http.patch(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/tcloud/load_balancers/${data.id}`,
        data,
      );
    },
    /**
     * 新增监听器
     * @param data 监听器信息
     */
    createListener(data: any) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}load_balancers/${data.lb_id}/listeners/create`,
        data,
      );
    },
    /**
     * 更新监听器
     * @param data 监听器信息
     */
    updateListener(data: any) {
      return http.patch(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}listeners/${data.id}`, data);
    },
    /*
     * 新建域名、新建url
     */
    createRules(data: {
      bk_biz_id?: number; // 业务ID
      lbl_id: string; // 监听器id
      rules: Record<string, any>; // 待创建规则
    }) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/tcloud/listeners/${
          data.lbl_id
        }/rules/create`,
        data,
      );
    },
    /**
     * 更新域名
     */
    updateDomains(
      listenerId: string,
      data: {
        bk_biz_id?: number;
        lbl_id: string; // 监听器ID
        domain: string; // 新域名
        new_domain: string; // 新域名
        certificate?: string; // 证书信息
      },
    ) {
      return http.patch(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}listeners/${listenerId}/domains`,
        data,
      );
    },
    /**
     * 删除域名、URL
     */
    deleteRules(
      listenerId: string,
      data: {
        bk_biz_id?: number; // 业务ID
        lbl_id: string; // 监听器id
        rule_ids?: string[]; // URL规则ID数组
        domain?: string; // 按域名删除, 没有指定规则id的时候必填
        new_default_domain?: string; // 新默认域名,删除的域名是默认域名的时候需要指定
      },
    ) {
      return http.delete(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/tcloud/listeners/${listenerId}/rules/batch`,
        { data },
      );
    },
    /**
     * 更新URL规则
     */
    updateUrl(data: {
      bk_biz_id?: number; // 业务ID
      lbl_id: string; // 监听器id
      rule_id: string; // URL规则ID数组
      url: string; // 监听的url
      scheduler: string; // 均衡方式
      certificate?: Record<string, any>; // 证书信息，当协议为HTTPS时必传
    }) {
      return http.patch(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/tcloud/listeners/${data.lbl_id}/rules/${
          data.rule_id
        }`,
        data,
      );
    },
    /**
     * 业务下给指定目标组批量添加RS
     * @param target_group_id 目标组id
     * @param data rs列表
     */
    addRsToTargetGroup(target_group_id: string, data: any) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}target_groups/${target_group_id}/targets/create`,
        data,
      );
    },
    /**
     * 业务下批量修改RS端口
     * @param target_group_id 目标组id
     * @param data { target_ids, new_port }
     */
    batchUpdateRsPort(target_group_id: string, data: any) {
      return http.patch(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}target_groups/${target_group_id}/targets/port`,
        data,
      );
    },
    /**
     * 业务下批量修改RS权重
     * @param target_group_id 目标组id
     * @param data { target_ids, new_weight }
     */
    batchUpdateRsWeight(target_group_id: string, data: any) {
      return http.patch(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}target_groups/${target_group_id}/targets/weight`,
        data,
      );
    },
  },
});
