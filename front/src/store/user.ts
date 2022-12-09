import { defineStore } from 'pinia';
import http from '@/http/index';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
export const useUser = defineStore('user', {
  state: () => ({
    user: '',
    username: '',
  }),

  actions: {
    setUser(user: string) {
      this.user = user;
    },

    // 测试
    async test() {
      const res = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/mock/api/v4/organization/user_info/`);
      return res;
    },

    // 用户信息
    async userInfo() {
      const res = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/user`);
      this.username = res.data.username;
      return res;
    },
  },
});
