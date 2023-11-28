import http from '@/http';
import { defineStore } from 'pinia';
import { QueryFilterType, IPageQuery } from '@/typings/common';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;


// 资源选型模块相关状态管理和接口定义
export const useSchemeStore = defineStore({
  id: 'schemeStore',
  state: () => ({}),
  actions: {
    /**
     * 获取资源选型方案列表
     * @param filter 过滤参数
     * @param page 分页参数
     * @returns 
     */
    listCloudSelectionScheme (filter: QueryFilterType, page: IPageQuery) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/selections/schemes/list`, { filter, page });
    },
    /**
     * 获取收藏的资源选型方案列表
     * @returns 
     */
    listCollection () {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/collections/cloud_selection_scheme/list`);
    }
  },
});
