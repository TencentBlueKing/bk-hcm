// @ts-check
import http from '@/http';
// import { Department } from '@/typings';
import { shallowRef } from 'vue';
import { defineStore } from 'pinia';

export const useAccountStore = defineStore({
  id: 'accountStore',
  state: () => ({
    fetching: false,
    list: shallowRef([]),
  }),
  actions: {
    /**
     * @description: 新增账号
     * @param {any} data
     * @return {*}
     */
    addAccount(data: any) {
      return http.post('/api/v1/cloud/accounts/create/', data);
    },
    /**
     * @description: 获取账号列表
     * @param {any} data
     * @return {*}
     */
    async getAccountList(params: any) {
      return await http.post('/api/v1/cloud/accounts/list/', params);
    },
    /**
     * @description: 获取账号详情
     * @param {any} data
     * @return {*}
     */
    async getAccountDetail(data: {id: number}) {
      return await http.post('/api/v1/cloud/accounts/retrieve/', data);
    },
    /**
     * @description: 测试云账号连接
     * @param {any} data
     * @return {*}
     */
    async testAccountConnection(data: any) {
      return await http.post('/api/v1/cloud/accounts/connection-test/', data);
    },
    /**
     * @description: 更新云账号
     * @param {any} data
     * @return {*}
     */
    async updateAccount(data: any) {
      return await http.post('/api/v1/cloud/accounts/update/', data);
    },
    /**
     * @description: 获取业务列表
     * @param {any}
     * @return {*}
     */
    async getBizList() {
      return await http.post('/api/v1/web/bk_bizs/list/');
    },
    /**
     * @description: 同步
     * @param {number} id
     * @return {*}
     */
    async accountSync(id: number) {
      await http.post('/mock/api/v4/sync/', id);
    },
    /**
     * @description: 删除
     * @param {number} id
     * @return {*}
     */
    async accountDelete(id: number) {
      return await http.post('/mock/api/v4/sync/', id);
    },
  },
});
