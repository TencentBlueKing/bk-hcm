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
    pageAuthData: [
      { type: 'account', action: 'find', id: 'account_find', path: '/resource/account' }, // 如果是列表查看权限 需要加上path
      { type: 'account', action: 'import', id: 'account_import' },
      { type: 'account', action: 'update', id: 'account_edit' },

      // 安全组
      { type: 'security_group', action: 'find', bk_biz_id: '2005000002', id: 'resource_find_security_group', path: '/business/security' },
      { type: 'security_group', action: 'create', id: 'iaas_resource_operate_security_group' },
    ],
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
