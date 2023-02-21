import http from '@/http';
import { defineStore } from 'pinia';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export const useAuditStore = defineStore({
  id: 'audit',
  state: () => ({
  }),
  actions: {
    list(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/audits/list`, data);
    },
    detail(id: number) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/audits/${id}`);
    },
  },
});
