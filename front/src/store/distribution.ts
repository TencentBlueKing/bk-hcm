import { defineStore } from 'pinia';
import { computed, ref } from 'vue';

export const useDistributionStore = defineStore('distribution', () => {
  const cloudAccountId = ref('');

  const computedCloudAccountId = computed(() => {
    return cloudAccountId;
  });

  const setCloudAccountId = (val: string) => {
    cloudAccountId.value = val;
  };

  return {
    cloudAccountId,
    computedCloudAccountId,
    setCloudAccountId,
  };
});
