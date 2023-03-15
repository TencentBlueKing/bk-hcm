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
    bizs: 0,
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
    async getAccountList(params: any, bizId?: number) {
      if (bizId > 0) {
        return await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/accounts/bizs/${bizId}`);
      }
      return await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/accounts/list`, params);
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
     * @description: 获取业务列表
     * @param {any}
     * @return {*}
     */
    async getBizList() {
      return await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/bk_bizs/list`);
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
    async updateBizsId(id: number) {
      this.bizs = id;
    },
  },
});
