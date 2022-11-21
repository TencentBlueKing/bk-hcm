import { defineStore } from 'pinia';

export const useUser = defineStore('user', {
  state: () => ({
    user: '',
  }),
  actions: {
    setUser(user) {
      this.user = user;
    },
  },
});
