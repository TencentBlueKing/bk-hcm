import { defineStore } from 'pinia';
import { ref } from 'vue';
import { useAccountStore } from './account';

export const useCloudAreaStore = defineStore('useCloudAreaStore', () => {
  const cloudAreaMap = ref<Map<number, string>>(new Map());
  const cloudAreaList = ref([]);
  const { getAllCloudAreas } = useAccountStore();

  const fetchAllCloudAreas = async () => {
    const res = await getAllCloudAreas();
    cloudAreaList.value = res?.data?.info || [];
    for (const { id, name } of cloudAreaList.value) {
      cloudAreaMap.value.set(id, name);
    }
  };

  return {
    cloudAreaMap,
    cloudAreaList,
    fetchAllCloudAreas,
  };
});
