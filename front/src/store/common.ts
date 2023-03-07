// @ts-check
import http from '@/http';
// import { Department } from '@/typings';
import { shallowRef } from 'vue';
import { defineStore } from 'pinia';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export const useCommonStore = defineStore({
  id: 'commonStore',
  state: () => ({
    list: shallowRef([]),
    authVerifyData: null,
    authVerifyParams: null,
  }),
  actions: {
    /**
     * @description: 权限鉴权
     * @param {any} data
     * @return {*}
     */
    async authVerify(data: any) {
      const res = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/auth/verify`, data);
      return res;
    },

    /**
     * @description: 权限操作鉴权
     * @param {any} data
     * @return {*}
     */
    async authActionUrl(data: any) {
      const res = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/auth/find/apply_perm_url`, data);
      return res.data;
    },

    async addAuthVerifyParams(data: any) {
      this.authVerifyParams = data;
    },
    /**
     * @description: 管理全局的按钮是否disabled和获取跳转链接的参数
     * @param {any} data
     * @return {*}
     */
    async addAuthVerifyData(data: any) {
      this.authVerifyData = data;
    },
  },
});
