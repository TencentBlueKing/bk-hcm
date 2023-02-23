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
    async authVerify(data: any, action: string[]) {
      const res = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/auth/verify`, data);
      res.data.results = res.data.results.reduce((p: any, e: any, i: number) => {    // 将数组转成对象
        p[`${action[i]}_authorized`] = e.authorized;
        return p;
      }, {});
      this.authVerifyData = res.data;
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
  },
});
