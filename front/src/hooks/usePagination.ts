import { Reactive, reactive } from 'vue';

export interface PaginationType {
  start: number;
  count: number;
  limit: number;
  current?: number;
  'limit-list'?: number[];
}
type GetDefaultPagination = (custom?: Partial<PaginationType>) => PaginationType;

export default function usePagination(cb: any, pageRef?: Reactive<PaginationType>) {
  const defaultPagination =
    window.innerHeight > 750
      ? { limit: 20, 'limit-list': [10, 20, 50, 100] }
      : { limit: 10, 'limit-list': [10, 20, 50, 100] };

  const getDefaultPagination: GetDefaultPagination = () => {
    const config = {
      start: 0,
      count: 0,
      current: 1,
      limit: defaultPagination.limit,
      'limit-list': defaultPagination['limit-list'],
    };
    return config;
  };

  const pagination = reactive(pageRef ? pageRef : getDefaultPagination());

  /**
   * 分页条数改变时调用
   * @param v 分页条数
   */
  const handlePageLimitChange = (v: number) => {
    pagination.limit = v;
    pagination.start = 0;
    pagination.current = 1;
    cb();
  };

  /**
   * 分页 start offset 改变时调用
   * @param v 当前 start offset
   */
  const handlePageValueChange = (v: number) => {
    pagination.start = (v - 1) * pagination.limit;
    pagination.current = v;
    cb();
  };

  return {
    pagination,
    handlePageLimitChange,
    handlePageValueChange,
  };
}
