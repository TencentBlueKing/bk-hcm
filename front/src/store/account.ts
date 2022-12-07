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
      try {
        return await http.post('/api/v1/cloud/accounts/list/', params);
      } catch (error) {
        console.error(error);
      }
    },
    /**
     * @description: 获取账号详情
     * @param {any} data
     * @return {*}
     */
    async getAccountDetail(id: string | number) {
      try {
        return await http.post('/api/v1/cloud/accounts/retrieve/', id);
      } catch (error) {
        console.error(error);
      }
    },
    /**
     * @description: 测试云账号连接
     * @param {any} data
     * @return {*}
     */
    async testAccountConnection(data: any) {
      try {
        return await http.post('/api/v1/cloud/accounts/connection-test/', data);
      } catch (error) {
        console.error(error);
      }
    },
    /**
     * @description: 更新云账号
     * @param {any} data
     * @return {*}
     */
    async updateAccount(data: any) {
      try {
        return await http.post('/api/v1/cloud/accounts/update/', data);
      } catch (error) {
        console.error(error);
      }
    },
    /**
     * @description: 获取业务列表
     * @param {any}
     * @return {*}
     */
    async getBizList() {
      try {
        return await http.post('/api/v1/web/bk_bizs/list/');
      } catch (error) {
        console.error(error);
      }
    },
    /**
     * @description: 同步
     * @param {number} id
     * @return {*}
     */
    async accountSync(id: number) {
      try {
        return await http.post('/mock/api/v4/sync/', id);
      } catch (error) {
        console.error(error);
      }
    },
    /**
     * @description: 删除
     * @param {number} id
     * @return {*}
     */
    async accountDelete(id: number) {
      try {
        return await http.post('/mock/api/v4/sync/', id);
      } catch (error) {
        console.error(error);
      }
    },
  },
});
