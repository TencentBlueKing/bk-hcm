import { reactive } from 'vue';

export default function usePagination(cb: any) {
  const pagination = reactive({ start: 0, limit: 10, count: 0 });

  /**
   * 分页条数改变时调用
   * @param v 分页条数
   */
  const handlePageLimitChange = (v: number) => {
    pagination.limit = v;
    pagination.start = 0;
    cb();
  };

  /**
   * 分页 start offset 改变时调用
   * @param v 当前 start offset
   */
  const handlePageValueChange = (v: number) => {
    pagination.start = (v - 1) * pagination.limit;
    cb();
  };

  return {
    pagination,
    handlePageLimitChange,
    handlePageValueChange,
  };
}
