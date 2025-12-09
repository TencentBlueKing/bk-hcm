import http from '@/http';
import { defineStore } from 'pinia';
import {
  IBizResourcesTicketsParam,
  IOpResourcesTicketsParam,
  IOpResourcesTicketsResult,
  IBizResourcesTicketsResult,
  ResourcePlanTicketByIdResult,
  IPlanTicket,
  IOpProductsResult,
  IPlanProductsResult,
  IBizsByOpProductResult,
  IListResourcesDemandsParam,
  IListResourcesDemandsResult,
  IPlanDemandResult,
  IListChangeLogsParam,
  IListChangeLogsResult,
  ResourcePlanTicketAuditResData,
} from '@/typings/resourcePlan';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export const useResourcePlanStore = defineStore('resourcePlanStore', {
  state: () => ({}),
  actions: {
    getDiskTypes() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/meta/disk_type/list`);
    },
    // 查询OBS项目类型列表。
    getObsProjects() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/meta/obs_project/list`);
    },
    // 业务视角查询资源预测单据。
    getBizResourcesTicketsList(bizId: number, data: IBizResourcesTicketsParam): Promise<IBizResourcesTicketsResult> {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/bizs/${bizId}/plans/resources/tickets/list`, data);
    },
    // 管理员查询资源预测单据。
    getOpResourcesTicketsList(data: IOpResourcesTicketsParam): Promise<IOpResourcesTicketsResult> {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plans/resources/tickets/list`, data);
    },
    // 获取 资源管理 资源预测申请单据详情。
    getBizResourcesTicketsById(bizId: number, id: string): Promise<ResourcePlanTicketByIdResult> {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/bizs/${bizId}/plans/resources/tickets/${id}`);
    },
    // 获取  服务请求 资源预测申请单据详情。
    getOpResourcesTicketsById(id: string): Promise<ResourcePlanTicketByIdResult> {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plans/resources/tickets/${id}`);
    },
    // 查询资源预测单据的审批流，包括审批状态、当前审批阶段等（业务下）
    getBizResourcesTicketsAuditById(bizId: number, id: string): Promise<ResourcePlanTicketAuditResData> {
      return http.get(`/api/v1/woa/bizs/${bizId}/plans/resources/tickets/${id}/audit`);
    },
    // 查询资源预测单据的审批流，包括审批状态、当前审批阶段等（服务请求下）
    getOpResourcesTicketsAuditById(id: string): Promise<ResourcePlanTicketAuditResData> {
      return http.get(`/api/v1/woa/plans/resources/tickets/${id}/audit`);
    },
    createBizPlan(data: IPlanTicket, bk_biz_id: number): { data: { id: string } } {
      return http.post(`/api/v1/woa/bizs/${bk_biz_id}/plans/resources/tickets/create`, data);
    },
    getBizOrgRelation(bizId: number) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/bizs/${bizId}/org/relation`);
    },
    // 获取预测类型列表
    getDemandClasses() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plan/demand_class/list`);
    },
    // 获取城市列表
    getRegions() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/meta/region/list`);
    },
    // 获取可用区列表
    getZones(region_ids?: string[]) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/meta/zone/list`, { region_ids });
    },
    getSources() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plan/demand_source/list`);
    },
    // 获取机型规格列表
    getDeviceClasses() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/meta/device_class/list`);
    },
    // 获取机型类型列表
    getDeviceTypes(device_classes?: string[], device_types?: string[]) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/meta/device_type/list`, { device_classes, device_types });
    },
    // 查询运营产品列表。
    getOpProductsList(): Promise<IOpProductsResult> {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/metas/op_products/list`);
    },
    // 查询规划产品列表。
    getPlanProductsList(): Promise<IPlanProductsResult> {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/metas/plan_products/list`);
    },
    // 根据运营产品ID查询业务列表。
    getBizsByOpProductList(data: { op_product_id: number }): Promise<IBizsByOpProductResult> {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/metas/bizs/by/op_product/list`, data);
    },
    // 获取预测类型列表
    getPlanTypes() {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/metas/plan_types/list`);
    },
    // 查询业务下资源预测需求列表
    getResourcesDemandsList(bk_biz_id: number, data: IListResourcesDemandsParam): Promise<IListResourcesDemandsResult> {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/bizs/${bk_biz_id}/plans/resources/demands/list`, data);
    },
    // 查询业务下资源预测需求详情信息
    getPlanDemand(bk_biz_id: number, demand_id: string): Promise<IPlanDemandResult> {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/bizs/${bk_biz_id}/plans/demands/${demand_id}`);
    },
    // 查询资源预测需求单的变更历史
    getListChangeLogs(bk_biz_id: number, data: IListChangeLogsParam): Promise<IListChangeLogsResult> {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/bizs/${bk_biz_id}/plans/demands/change_logs/list`, data);
    },
    // 查询管理下资源预测需求列表
    getResourcesDemandsListByOrg(data: IListResourcesDemandsParam): Promise<IListResourcesDemandsResult> {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plans/resources/demands/list`, data);
    },
    // 查询管理下资源预测需求详情信息
    getPlanDemandByOrg(demand_id: string): Promise<IPlanDemandResult> {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plans/demands/${demand_id}`);
    },
    // 查询管理下资源预测需求单的变更历史
    getListChangeLogsByOrg(data: IListChangeLogsParam): Promise<IListChangeLogsResult> {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plans/demands/change_logs/list`, data);
    },
    // 批量取消资源预测需求
    cancelResourcesDemands(
      bk_biz_id: number,
      data: {
        cancel_demands: { demand_id: string; remained_cpu_core: number }[];
      },
    ) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/bizs/${bk_biz_id}/plans/resources/demands/cancel`, data);
    },
  },
});
