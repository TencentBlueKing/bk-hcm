import { defineStore } from 'pinia';
import http from '@/http/index';

export const useUser = defineStore('user', {
  state: () => ({
    user: '',
  }),
  actions: {
    setUser(user: string) {
      this.user = user;
    },

    // 测试
    async test() {
      const res = await http.get('/api/v4/organization/user_info/');
      return res;
    },
  },
});
