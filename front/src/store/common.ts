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

      // 业务访问权限
      { type: 'biz', action: 'access', id: 'biz_access' },

      // 目前主机、vpc、子网、安全组、云硬盘、网络接口、弹性IP、路由表、镜像等都当作iaas统一鉴权，为了方便，使用cvm当作整个iaas鉴权
      { type: 'cvm', action: 'find',  id: 'resource_find', path: ['/business/host', '/resource/resource'] },    // 业务 资源对应的路径
      { type: 'cvm', action: 'create', id: 'iaas_resource_operate' },    // iaas操作
      { type: 'cvm', action: 'delete', id: 'iaas_resource_delete' },    // iaas删除

      // // 安全组
      // eslint-disable-next-line max-len
      // { type: 'security_group', action: 'find',  id: 'resource_find_security', path: ['/business/security', '/resource/resource'] },    // 业务 资源对应的路径
      // { type: 'security_group', action: 'delete', id: 'iaas_resource_delete_security' },    // iaas删除
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
