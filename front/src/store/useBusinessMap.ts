import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { useAccountStore } from './account';

export const useBusinessMapStore = defineStore('businessMapStore', () => {
  const businessMap = ref<Map<number, string>>(new Map());
  const businessMapSize = computed(() => businessMap.value.size);
  const businessList = ref([]);

  const accountStore = useAccountStore();
  const fetchBusinessMap = async () => {
    const { data } = await accountStore.getBizList();
    if (data && data.length > 0) {
      businessList.value = data;
      businessMap.value = new Map();
      for (const { id, name } of data) {
        businessMap.value.set(id, name);
      }
    }
  };

  const getNameFromBusinessMap = (id: number) => {
    return businessMap.value.get(id) || '';
  };

  return {
    businessMap,
    businessList,
    businessMapSize,
    fetchBusinessMap,
    getNameFromBusinessMap,
  };
});
