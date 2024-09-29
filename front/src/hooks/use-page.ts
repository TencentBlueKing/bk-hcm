import { reactive } from 'vue';
import routeQuery from '@/router/utils/query';
import { IPageQuery, PaginationType, SortType } from '@/typings';

type GetDefaultPagination = (custom?: Partial<PaginationType>) => PaginationType;

export default function usePage(useQuery = true) {
  const defaultPagination =
    window.innerHeight > 750 ? { limit: 20, 'limit-list': [20, 50, 100] } : { limit: 10, 'limit-list': [10, 50, 100] };

  const getDefaultPagination: GetDefaultPagination = (custom = {}) => {
    const config = {
      count: 0,
      current: useQuery ? parseInt(routeQuery.get('page', 1), 10) : 1,
      limit: useQuery ? parseInt(routeQuery.get('limit', defaultPagination.limit), 10) : defaultPagination.limit,
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

  const pagination = reactive(getDefaultPagination());

  const handlePageChange = (page: number) => {
    if (!useQuery) {
      pagination.current = page;
      return;
    }
    routeQuery.set({
      page,
      _t: Date.now(),
    });
  };

  const handlePageSizeChange = (limit: number) => {
    if (!useQuery) {
      pagination.limit = limit;
      return;
    }
    routeQuery.set({
      limit,
      page: 1,
      _t: Date.now(),
    });
  };

  const handleSort = ({ column, type }: SortType) => {
    if (!useQuery) {
      return;
    }
    routeQuery.set({
      sort: column.field,
      order: type === 'desc' ? 'DESC' : 'ASC',
      _t: Date.now(),
    });
  };

  return {
    pagination,
    getDefaultPagination,
    getPageParams,
    handlePageChange,
    handlePageSizeChange,
    handleSort,
  };
}
