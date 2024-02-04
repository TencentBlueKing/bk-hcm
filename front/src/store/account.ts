// @ts-check
import http from '@/http';
// import { Department } from '@/typings';
import { shallowRef } from 'vue';
import { defineStore } from 'pinia';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export const useAccountStore = defineStore({
  id: 'accountStore',
  state: () => ({
    fetching: false,
    list: shallowRef([]),
    bizs: 0 as number,
    accountList: shallowRef([]),
  }),
  actions: {
    /**
     * @description: 新增账号
     * @param {any} data
     * @return {*}
     */
    addAccount(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/accounts/create`, data);
    },
    /**
     * @description: 获取账号列表
     * @param {any} data
     * @param {number} bizId
     * @return {*}
     */
    async getAccountList(data: any, bizId?: number | string, isRes?: boolean) {
      if (isRes) {
        return await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/accounts/resources/accounts/list`, data);
      }
      if (bizId > 0) {
        return await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/accounts/bizs/${bizId}`, { params: data.params });
      }
      return await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/accounts/list`, data);
    },
    /**
     * @description: 获取账号详情
     * @param {any} data
     * @return {*}
     */
    async getAccountDetail(id: string | string[]) {
      return await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/accounts/${id}`);
    },
    /**
     * @description: 创建时测试云账号连接
     * @param {any} data
     * @return {*}
     */
    async testAccountConnection(data: any) {
      return await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/accounts/check`, data);
    },
    /**
     * @description: 更新时测试云账号连接
     * @param {any} data
     * @return {*}
     */
    async updateTestAccount(data: any) {
      const { id } = data;
      delete data.id;
      return await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/accounts/${id}/check`, data);
    },
    /**
     * @description: 更新云账号
     * @param {any} data
     * @return {*}
     */
    async updateAccount(data: any) {
      const { id } = data;
      delete data.id;
      return await http.patch(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/accounts/${id}`, data);
    },
    /**
     * @description: 获取全量业务列表
     * @param {any}
     * @return {*}
     */
    async getBizList() {
      return await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/bk_bizs/list`);
    },
    /**
     * @description: 获取全量管控区域数据
     * @param {any}
     * @return {*}
     */
    async getAllCloudAreas() {
      return await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/all/cloud_areas/list`);
    },
    /**
     * @description: 根据账号id获取业务id
     * @param {any}
     * @return {*}
     */
    async getBizIdWithAccountId(id: string) {
      return await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/accounts/${id}`);
    },
    /**
     * @description: 获取有权限的业务列表
     * @param {any}
     * @return {*}
     */
    async getBizListWithAuth() {
      return await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/authorized/bizs/list`);
    },
    /**
     * @description: 获取审计下有权限的业务列表
     * @param {any}
     * @return {*}
     */
    async getBizAuditListWithAuth() {
      return await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/authorized/audit/bizs/list`);
    },
    /**
     * @description: 获取部门信息
     * @param {any}
     * @return {*}
     */
    async getDepartmentInfo(departmentId: number) {
      return await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/departments/${departmentId}`);
    },
    /**
     * @description: 同步
     * @param {number} id
     * @return {*}
     */
    async accountSync(id: number) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/accounts/${id}/sync`);
    },
    /**
     * @description: 删除
     * @param {number} id
     * @return {*}
     */
    async accountDelete(id: number) {
      return http.delete(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/accounts/${id}`);
    },
    /**
     * @description: 删除前校验
     * @param {number} id
     * @return {*}
     */
    async accountDeleteValidate(id: number) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/accounts/${id}/delete/validate`);
    },
    /**
     * @description: 申请账号
     * @param {number} data
     * @return {*}
     */
    async applyAccount(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/applications/types/add_account`, data);
    },
    /**
     * @description: 查询申请账号列表
     * @param {number} data
     * @return {*}
     */
    async getApplyAccountList(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/applications/list`, data);
    },
    /**
     * @description: 查询申请账号列表
     * @param {number} data
     * @return {*}
     */
    async getApplyAccountDetail(id: string) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/applications/${id}`);
    },
    /**
     * @description: 撤销申请
     * @param {number} id
     * @return {*}
     */
    async cancelApplyAccount(id: string) {
      return http.patch(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/applications/${id}/cancel`);
    },
    /**
     * @description: 更新业务id
     * @param {number} id
     * @return {*}
     */
    async updateBizsId(id: number | string) {
      this.bizs = id;
    },

    async updateAccountList(data: any) {
      console.log('data', data);
      this.accountList = data?.map(({ id, name }: { id: string; name: string }) => ({
        id,
        name,
      }));
      console.log('this.accountList', this.accountList);
    },

    /**
     * 获取我的审批列表
     * @param data
     */
    async getApprovalList(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/tickets/types/my_approval/list`, data);
    },

    /**
     * 拒绝/通过审批单据
     * @param data
     */
    async approveTickets(data: any) {
      http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/tickets/approve`, data);
    },
  },
});
