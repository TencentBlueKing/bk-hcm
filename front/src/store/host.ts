import { defineStore } from 'pinia';

export const useHostStore = defineStore('host', {
  state: () => ({
    // 云地域
    regionList: [],
  }),
});
