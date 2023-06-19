import { computed, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

export const DEFAULT_PAGE_SIZE = 20;
export const DEFAULT_PAGE_INDEX = 1;
export const PAGE_SIZE_KEY = 'limit';
export const PAGE_INDEX_KEY = 'current';

export const useMemoPagination = () => {
  const route = useRoute();
  const router = useRouter();
  const pageSize = ref(+route.query[PAGE_SIZE_KEY] || DEFAULT_PAGE_SIZE);
  const pageIndex = ref(+route.query[PAGE_INDEX_KEY] || DEFAULT_PAGE_INDEX);

  const memoPageSize = computed(() => pageSize.value);
  const memoPageIndex = computed(() => pageIndex.value);
  const memoPageStart = computed(() => (memoPageIndex.value - 1) * memoPageSize.value);

  const setMemoPageSize = (val: number) => {
    pageSize.value = val;
    router.push({
      query: {
        [PAGE_INDEX_KEY]: pageIndex.value,
        [PAGE_SIZE_KEY]: val,
      },
    });
  };

  const setMemoPageIndex = (val: number) => {
    pageIndex.value = val;
    router.push({
      query: {
        [PAGE_SIZE_KEY]: pageSize.value,
        [PAGE_INDEX_KEY]: val,
      },
    });
  };

  watch(
    () => route.query,
    (query) => {
      if (query[PAGE_INDEX_KEY] && !isNaN(+query[PAGE_INDEX_KEY])) pageIndex.value = +query[PAGE_INDEX_KEY];
      if (query[PAGE_SIZE_KEY] && !isNaN(+query[PAGE_SIZE_KEY])) pageSize.value = +query[PAGE_SIZE_KEY];
    },
    {
      deep: true,
      immediate: true,
    },
  );

  return {
    setMemoPageIndex,
    setMemoPageSize,
    memoPageIndex,
    memoPageSize,
    memoPageStart,
  };
};
