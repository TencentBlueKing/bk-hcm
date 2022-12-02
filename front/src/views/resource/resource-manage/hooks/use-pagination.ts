/**
 * 分页相关状态和事件
 */
import {
  ref,
  onBeforeMount,
} from 'vue';

type CallBackType = ({ current, limit, count }: { [propName: string]: number }) => any;

export default (callBack: CallBackType) => {
  const pagination = ref({
    current: 1,
    limit: 10,
    count: 0,
  });

  const triggerCb = () => {
    callBack(pagination.value);
  };

  const handlePageChange = (current: number) => {
    pagination.value.current = current;
    triggerCb();
  };

  const handlePageSizeChange = (limit: number) => {
    pagination.value.limit = limit;
    triggerCb();
  };

  onBeforeMount(triggerCb);

  return {
    pagination,
    handlePageChange,
    handlePageSizeChange,
  };
};
