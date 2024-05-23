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
    authVerifyData: null as any,
    authVerifyParams: null as any,
    pageAuthData: [
      // { type: 'cloud_selection_scheme', action: 'create', id: 'cloud_selection_recommend' },
      {
        type: 'cloud_selection_scheme',
        action: 'create',
        id: 'cloud_selection_recommend',
        path: '/scheme/recommendation',
      },
      { type: 'cloud_selection_scheme', action: 'find', id: 'cloud_selection_find', path: '/scheme/deployment/list' },
      { type: 'cloud_selection_scheme', action: 'update', id: 'cloud_selection_edit', path: '/scheme/recommendation' },
      {
        type: 'cloud_selection_scheme',
        action: 'delete',
        id: 'cloud_selection_delete',
        path: '/scheme/recommendation',
      },

      { type: 'account', action: 'find', id: 'account_find', path: '/resource/account' }, // 如果是列表查看权限 需要加上path
      { type: 'account', action: 'import', id: 'account_import', path: '/resource/resource' },
      { type: 'account', action: 'update', id: 'account_edit' },

      // 业务访问权限
      { type: 'biz', action: 'access', id: 'biz_access' },

      // 目前资源下主机、vpc、子网、安全组、云硬盘、网络接口、弹性IP、路由表、镜像等都当作iaas统一鉴权，为了方便，使用cvm当作整个iaas鉴权
      { type: 'cvm', action: 'find', id: 'resource_find', path: ['/resource/resource'] }, // 业务 资源对应的路径
      { type: 'cvm', action: 'create', id: 'iaas_resource_create' }, // iaas创建
      { type: 'cvm', action: 'update', id: 'iaas_resource_operate' }, // iaas编辑更新
      { type: 'cvm', action: 'delete', id: 'iaas_resource_delete' }, // iaas删除

      // // 安全组
      // eslint-disable-next-line max-len
      // { type: 'security_group', action: 'find',  id: 'resource_find_security', path: ['/business/security', '/resource/resource'] },    // 业务 资源对应的路径
      // { type: 'security_group', action: 'delete', id: 'iaas_resource_delete_security' },    // iaas删除

      // 目前业务下主机、vpc、子网、安全组、云硬盘、网络接口、弹性IP、路由表、镜像等都当作iaas统一鉴权，为了方便，使用cvm当作整个业务iaas鉴权
      { type: 'cvm', action: 'find', id: 'resource_find', bk_biz_id: 0 }, // 业务 资源对应的路径
      { type: 'cvm', action: 'create', id: 'biz_iaas_resource_create', bk_biz_id: 0 }, // 业务iaas创建
      { type: 'cvm', action: 'update', id: 'biz_iaas_resource_operate', bk_biz_id: 0 }, // 业务iaas编辑更新
      { type: 'cvm', action: 'delete', id: 'biz_iaas_resource_delete', bk_biz_id: 0 }, // 业务iaas删除

      { type: 'biz_audit', action: 'find', id: 'resource_audit_find' }, // 审计查看权限

      { type: 'recycle_bin', action: 'find', id: 'recycle_bin_find', path: '/resource/recyclebin' }, // 回收站查看权限
      { type: 'recycle_bin', action: 'recycle', id: 'recycle_bin_manage' }, // 回收站管理

      // 证书权限
      { type: 'cert', action: 'create', id: 'cert_resource_create' }, // 资源 证书上传
      { type: 'cert', action: 'create', id: 'biz_cert_resource_create', bk_biz_id: 0 }, // 业务 证书上传
      { type: 'cert', action: 'delete', id: 'cert_resource_delete' }, // 资源 证书删除
      { type: 'cert', action: 'delete', id: 'biz_cert_resource_delete', bk_biz_id: 0 }, // 业务 证书删除
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

    async updatePageAuthData(data: any) {
      this.pageAuthData = data;
    },
  },
});
