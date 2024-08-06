import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useGlobalPermissionDialog = defineStore('useGlobalPermissionDialog', () => {
  const isShow = ref(false);

  const setShow = (val: boolean) => {
    isShow.value = val;
  };

  return {
    isShow,
    setShow,
  };
});
