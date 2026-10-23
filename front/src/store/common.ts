import { ref } from 'vue';
import { defineStore } from 'pinia';

export const useCommonStore = defineStore('common', () => {
  const isNoticeAlert = ref(false);

  const setIsNoticeAlert = (val: boolean) => {
    isNoticeAlert.value = val;
  };

  return {
    isNoticeAlert,
    setIsNoticeAlert,
  };
});
