import { reactive } from 'vue';

const DEFAULT_LIMIT = 20;

export function usePagination(limit = DEFAULT_LIMIT) {
  const pagination = reactive({
    location: 'left',
    align: 'right',
    current: 1,
    limit,
    count: 100,
  });

  function changePagination(key: keyof typeof pagination) {
    return (val: number) => {
      Object.assign(pagination, {
        [key]: val,
        ...(key === 'limit'
          ? {
              current: 1,
            }
          : {}),
      });
    };
  }

  return {
    pagination,
    handlePageValueChange: changePagination('current'),
    handlePageLimitChange: changePagination('limit'),
    handleTotalChange: changePagination('count'),
  };
}
