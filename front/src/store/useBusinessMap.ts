import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { useAccountStore } from './account';

export const useBusinessMapStore = defineStore('businessMapStore', () => {
  const businessMap = ref<Map<number, string>>(new Map());
  const businessMapSize = computed(() => businessMap.value.size);

  const accountStore = useAccountStore();
  const updateBusinessMap = async () => {
    const { data } = await accountStore.getBizList();
    if (data && data.length > 0) {
      businessMap.value = new Map();
      for (const { id, name } of data) {
        businessMap.value.set(id, name);
      }
    }
  };

  const getNameFromBusinessMap = async (id: number) => {
    if (businessMapSize.value < 1 || !businessMap.value.get(id)) {
      await updateBusinessMap();
    }
    return businessMap.value.get(id) || '';
  };

  return {
    businessMap,
    businessMapSize,
    updateBusinessMap,
    getNameFromBusinessMap,
  };
});
