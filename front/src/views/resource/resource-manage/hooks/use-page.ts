/**
 * 分页相关状态和事件
 */
import {
  ref,
  onBeforeMount,
} from 'vue';

type CallBackType = ({ count, start, limit, sort, order }: { [propName: string]: number | boolean | string }) => any;
type SortType = {
  column: {
    prop: string
  };
  type: string
};

export default (callBack: CallBackType) => {
  const pagination = ref({
    current: 1,
    limit: 10,
    count: 0,
  });
  const sort = ref();
  const order = ref();

  // 把最终数据交给接口
  const triggerCb = () => {
    callBack({
      count: true,
      start: pagination.value.current * pagination.value.limit,
      limit: pagination.value.limit,
      sort: sort.value,
      order: order.value,
    });
  };

  // 页码变化发生的事件
  const handlePageChange = (current: number) => {
    pagination.value.current = current;
    triggerCb();
  };

  // 条数变化发生的事件
  const handlePageSizeChange = (limit: number) => {
    pagination.value.limit = limit;
    triggerCb();
  };

  // 排序变化发生的事件
  const handleSort = ({ column, type }: SortType) => {
    sort.value = column.prop;
    order.value = type;
  };

  onBeforeMount(triggerCb);

  return {
    pagination,
    handlePageChange,
    handlePageSizeChange,
    handleSort,
  };
};
