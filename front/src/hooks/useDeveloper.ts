import { useDeveloperStore } from '@/stores';
import { ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { usePagination } from './usePagination';

export function useMappingDeveloper(isMap = false) {
  const developerStore = useDeveloperStore();
  const route = useRoute();
  const list = ref([
    {
      email: 'lockiechen@tencent.com',
      username: 'roy',
      wework: 'lockiechen',
    },
    {
      email: 'roy@tencent.com',
      username: 'roy',
      wework: 'lockiechen',
    },
    {
      email: 'demo@tencent.com',
      username: 'roy',
      wework: 'demo',
    },
  ]);
  const { pagination, handleTotalChange, handlePageValueChange, handlePageLimitChange } = usePagination();

  watch(
    () => ({ ...pagination }),
    (newPagination) => {
      getDeveloperMap(newPagination);
    },
    { immediate: true },
  );

  async function getDeveloperMap(paginationConf: typeof pagination) {
    const data = await developerStore.fetchDeveloperMap({
      is_mapped: isMap,
      projectId: route.params.projectId,
      page: paginationConf.current,
      page_size: paginationConf.limit,
    });
    console.log(data);
    // handleTotalChange(data.count);
    // list.value = data.result;
  }

  return {
    list,
    pagination,
    handleTotalChange,
    handlePageValueChange,
    handlePageLimitChange,
  };
}
