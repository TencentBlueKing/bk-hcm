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
    async addAccount(data: any) {
      try {
        return await http.post('/mock/api/v4/add/', data);
      } catch (error) {
        console.error(error);
      }
    },
    /**
     * @description: 获取账号列表
     * @param {any} data
     * @return {*}
     */
    async getAccountList(params: any) {
      try {
        return await http.get('/mock/api/v4/get/', params);
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
