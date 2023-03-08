import http from '@/http';
import { defineStore } from 'pinia';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export const useBusinessStore = defineStore({
  id: 'businessStore',
  state: () => ({
  }),
  actions: {
    /**
     * @description: 新增安全组
     * @param {any} data
     * @return {*}
     */
    addSecurity(id: string, data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/bizs/${id}/security_groups/create`, data);
    },
  },
});
