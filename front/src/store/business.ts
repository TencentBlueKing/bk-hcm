import http from '@/http';
import { defineStore } from 'pinia';
import { AsyncTaskDetailResp, ClbQuotasResp, LbPriceInquiryResp } from '@/typings';
import { VendorEnum } from '@/common/constant';
import { useWhereAmI } from '@/hooks/useWhereAmI';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export interface ICloneSecurityParams {
  id: string;
  name: string;
  manager: string;
  bak_manager: string;
}

export const useBusinessStore = defineStore('business', () => {
  const { getBusinessApiPath } = useWhereAmI();

  return {
    /**
     * @description: 获取资源列表 - 业务下
     * @param {any} data
     * @param {string} type
     * @return {*}
     */
    list(data: any, type: string, config = {}) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}${type}/list`, data, config);
    },
    getCommonList(data: any, url: string, config = {}) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}${url}`, data, config);
    },
    /**
     * 根据id获取对应资源详情信息
     * @param type 资源类型
     * @param id 资源id
     * @returns 资源详情信息
     */
    detail(type: string, id: number | string, config = {}) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}${type}/${id}`, config);
    },
    /**
     * common-批量删除资源
     * @param type 资源类型
     * @param data 资源ids
     */
    deleteBatch(type: string, data: { ids: string[] }) {
      return http.delete(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}${type}/batch`, { data });
    },
    /**
     * @description: 新增安全组
     * @param {any} data
     * @return {*}
     */
    addSecurity(data: any, isRes = false) {
      if (isRes) return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/security_groups/create`, data);
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}security_groups/create`, data);
    },
    /**
     * @description: 克隆安全组
     * @param {ICloneSecurityParams} data
     * @return {*}
     */
    cloneSecurity(data: ICloneSecurityParams) {
      const { id, name, manager, bak_manager } = data;
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}security_groups/${id}/clone`, {
        name,
        manager,
        bak_manager,
      });
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
    getRsList(tg_id: string, data: { filter: Object; page: Object }) {
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
     * 业务下监听器域名列表
     */
    getDomainListByListenerId(vendor: VendorEnum, listenerId: string) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/${vendor}/listeners/${listenerId}/domains/list`,
      );
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
    updateLbDetail(
      vendor: VendorEnum,
      data: {
        bk_biz_id?: string;
        id: string;
        name?: string;
        internet_charge_type?: string;
        internet_max_bandwidth_out?: number;
        delete_protect?: boolean;
        load_balancer_pass_to_target?: boolean;
        snat_pro?: boolean;
        memo?: string;
      },
    ) {
      return http.patch(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/${vendor}/load_balancers/${data.id}`,
        data,
      );
    },
    /**
     * 新增监听器
     */
    createListener(data: any) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}load_balancers/${data.lb_id}/listeners/create`,
        data,
      );
    },
    /**
     * 更新监听器
     */
    updateListener(data: any) {
      return http.patch(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}listeners/${data.id}`, data);
    },
    /*
     * 新建域名、新建url
     */
    createRules(data: {
      bk_biz_id?: number;
      lbl_id: string;
      url?: string;
      scheduler?: string;
      target_group_id?: string;
      vendor: VendorEnum;
      domains: string[];
    }) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/${data.vendor}/listeners/${
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
        lbl_id: string;
        domain: string;
        new_domain?: string;
        certificate?: Object;
        default_server?: boolean;
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
      vendor: VendorEnum,
      listenerId: string,
      data: {
        bk_biz_id?: number;
        lbl_id: string;
        rule_ids?: string[];
        domain?: string;
        new_default_domain?: string;
      },
    ) {
      return http.delete(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/${vendor}/listeners/${listenerId}/rules/batch`,
        { data },
      );
    },
    /**
     * 批量删除域名
     */
    batchDeleteDomains(data: {
      bk_biz_id?: number;
      vendor: VendorEnum;
      lbl_id: string;
      domains: string[];
      new_default_domain?: string;
    }) {
      return http.delete(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/${data.vendor}/listeners/${
          data.lbl_id
        }/rules/by/domains/batch`,
        { data },
      );
    },
    /**
     * 更新URL规则
     */
    updateUrl(data: {
      bk_biz_id?: number;
      vendor: VendorEnum;
      lbl_id: string;
      rule_id: string;
      url: string;
      scheduler: string;
      certificate?: Record<string, any>;
      target_group_id?: string;
    }) {
      return http.patch(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/${data.vendor}/listeners/${
          data.lbl_id
        }/rules/${data.rule_id}`,
        data,
      );
    },
    /**
     * 业务下给指定目标组批量添加RS
     */
    batchAddTargets(data: {
      account_id: string;
      target_groups: {
        target_group_id: string;
        targets: { inst_type: string; cloud_inst_id: string; port: number; weight: number }[];
      }[];
    }) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}target_groups/targets/create`,
        data,
      );
    },
    /**
     * 查询操作记录异步记录指定批次的子任务列表
     */
    getAsyncTaskList(data: { audit_id: string; flow_id: string; action_id: string }) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}audits/async_task/list`, data);
    },
    /**
     * 查询操作记录异步任务进度流
     */
    getAsyncFlowList(data: { audit_id: number; flow_id: string }) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}audits/async_flow/list`, data);
    },
    /**
     * 更新目标组健康检查
     */
    updateHealthCheck(data: {
      id: string;
      health_check: {
        health_switch: 0 | 1;
        time_out: number;
        interval_time: number;
        health_num: number;
        un_health_num: number;
        check_port: number;
        check_type: 'TCP' | 'HTTP' | 'HTTPS' | 'GRPC' | 'PING' | 'CUSTOM';
        http_code: string;
        http_version: string;
        http_check_path?: string;
        http_check_domain?: string;
        http_check_method?: 'HEAD' | 'GET';
        source_ip_type: 0 | 1;
        context_type: 'HEX' | 'TEXT';
      };
    }) {
      return http.patch(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}target_groups/${data.id}/health_check`,
        data,
      );
    },
    /**
     * 获取目标组列表
     */
    getTargetGroupList(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}target_groups/list`, data);
    },
    /*
     * 业务下给指定目标组移除RS
     */
    batchDeleteTargets(data: { account_id: string; target_ids: string[] }) {
      return http.delete(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}target_groups/targets/batch`, {
        data,
      });
    },
    /**
     * 查询指定的目标组绑定的负载均衡下的端口健康信息
     */
    asyncGetTargetsHealth(target_group_id: string, data: { cloud_lb_ids: string[] }) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}target_groups/${target_group_id}/targets/health`,
        data,
      );
    },
    /**
     * 查询指定的负载均衡下的监听器数量
     */
    asyncGetListenerCount(data: { lb_ids: string[] }) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}load_balancers/listeners/count`,
        data,
      );
    },
    /**
     * 重试异步任务
     */
    retryAsyncTask(data: { bk_biz_id?: number; lb_id: string; flow_id: string; task_id: string }) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}load_balancers/${data.lb_id}/async_tasks/retry`,
        data,
      );
    },
    /**
     * 终止指定异步任务操作
     */
    endTask(data: { bk_biz_id?: number; lb_id: string; flow_id: string }) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}load_balancers/${
          data.lb_id
        }/async_flows/terminate`,
        data,
      );
    },
    /**
     * 获取任务终止后rs的状态
     */
    getFlowResults(data: { bk_biz_id?: number; lb_id?: string; flow_id: string }) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}load_balancers/${data.lb_id}/async_tasks/result`,
        data,
      );
    },
    /**
     * 复制flow参数重新执行
     */
    excuteTask(data: { bk_biz_id?: number; lb_id: string; flow_id: string }) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}load_balancers/${data.lb_id}/async_flows/clone`,
        data,
      );
    },
    /**
     * 获取腾讯云账号负载均衡的配额
     */
    getClbQuotas(data: { account_id: string; region: string }): Promise<ClbQuotasResp> {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}load_balancers/quotas`, data);
    },
    /**
     * 查询负载均衡价格
     */
    lbPricesInquiry(data: any): Promise<LbPriceInquiryResp> {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}load_balancers/prices/inquiry`,
        data,
      );
    },
    /**
     * 查询负载均衡状态锁定详情
     */
    getLBLockStatus(id: string) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}load_balancers/${id}/lock/status`);
    },
    /**
     * 查询异步任务详情
     */
    getAsyncTaskDetail(flowId: string): Promise<AsyncTaskDetailResp> {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/async_task/flows/${flowId}`);
    },
    /**
     * 业务下新增腾讯云负载均衡SNAT IP
     */
    createSnatIps(lb_id: string, vendor: VendorEnum, data: { snat_ips: { subnet_id: string; ip?: string }[] }) {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/${vendor}/load_balancers/${lb_id}/snat_ips/create`,
        data,
      );
    },
    /**
     * 业务下删除腾讯云负载均衡SNAT IP
     */
    deleteSnatIps(lb_id: string, vendor: VendorEnum, data: { delete_ips: string[] }) {
      return http.delete(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/${vendor}/load_balancers/${lb_id}/snat_ips`,
        {
          data,
        },
      );
    },
    /**
     * 查询监听器绑定的rs的权重情况
     */
    getListenersRsWeightStat(ids: string[]) {
      return http.post(`/api/v1/cloud/${getBusinessApiPath()}listeners/rs_weight_stat`, { ids });
    },
    // 主机所关联资源(硬盘, eip)的个数
    getRelResByCvmIds(data: { ids: string[] }) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}cvms/rel_res/batch`, data);
    },
    // 虚拟机回收
    recycledCvmsData(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}cvms/recycle`, data);
    },
    // 操作主机相关
    cvmOperate(type: string, data: { ids: string[] }) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}cvms/batch/${type}`, data);
    },
  };
});
