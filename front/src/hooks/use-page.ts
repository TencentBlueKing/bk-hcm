import { computed, type Reactive, reactive } from 'vue';
import routeQuery from '@/router/utils/query';
import { IPageQuery, PaginationType, SortType } from '@/typings';

type GetDefaultPagination = (custom?: Partial<PaginationType>) => PaginationType;

export default function usePage(enableQuery = true, pageRef?: Reactive<PaginationType>) {
  const defaultPagination =
    window.innerHeight > 750
      ? { limit: 20, 'limit-list': [10, 20, 50, 100] }
      : { limit: 10, 'limit-list': [10, 20, 50, 100] };

  const getDefaultPagination: GetDefaultPagination = (custom = {}) => {
    const config = {
      count: 0,
      current: enableQuery ? parseInt(routeQuery.get('page', 1), 10) : 1,
      limit: enableQuery ? parseInt(routeQuery.get('limit', defaultPagination.limit), 10) : defaultPagination.limit,
      'limit-list': custom['limit-list'] || defaultPagination['limit-list'],
    };
    return config;
  };

  const getPageParams = (pagination: PaginationType, extra?: Partial<IPageQuery>) => {
    return {
      start: (pagination.current - 1) * pagination.limit,
      limit: pagination.limit,
      ...extra,
    };
  };

  const sorting = reactive<{ sort?: string; order?: 'DESC' | 'ASC' }>({});
  const pageParams = computed(() => {
    return {
      start: (pagination.current - 1) * pagination.limit,
      limit: pagination.limit,
      ...sorting,
    };
  });

  // 传递了pageRef则直接使用，当enableQuery为false时可用于在多个组件之间共享pagination
  const pagination = reactive(pageRef ? pageRef : getDefaultPagination());

  const handlePageChange = (page: number) => {
    if (!enableQuery) {
      pagination.current = page;
      return;
    }
    routeQuery.set({
      page,
      _t: Date.now(),
    });
  };

  const handlePageSizeChange = (limit: number) => {
    if (!enableQuery) {
      pagination.limit = limit;
      pagination.current = 1;
      return;
    }
    routeQuery.set({
      limit,
      page: 1,
      _t: Date.now(),
    });
  };

  const handleSort = ({ column, type }: SortType) => {
    const sort = type === 'null' ? undefined : column.field;
    const order = type === 'null' ? undefined : (type.toUpperCase() as 'DESC' | 'ASC');
    if (!enableQuery) {
      sorting.sort = sort;
      sorting.order = order;
      return;
    }
    routeQuery.set({
      sort,
      order,
      _t: Date.now(),
    });
  };

  return {
    pagination,
    pageParams,
    getDefaultPagination,
    getPageParams,
    handlePageChange,
    handlePageSizeChange,
    handleSort,
  };
}
