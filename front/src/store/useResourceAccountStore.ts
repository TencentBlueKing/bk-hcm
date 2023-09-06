import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useResourceAccountStore = defineStore(
  'useResourceAccountStore',
  () => {
    const resourceAccount = ref({});
    const setResourceAccount = (val: Object) => {
      resourceAccount.value = val;
    };

    return {
      resourceAccount,
      setResourceAccount,
    };
  },
);
