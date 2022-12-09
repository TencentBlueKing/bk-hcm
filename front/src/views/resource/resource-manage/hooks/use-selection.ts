/**
 * 选择相关状态和事件
 */
import {
  ref,
} from 'vue';

export default () => {
  const selections = ref([]);

  const handleSelectionChange = ({ row }: { row: any[] }) => {
    selections.value = row;
  };

  const clearSelction = () => {
    selections.value = [];
  };

  return {
    selections,
    handleSelectionChange,
    clearSelction,
  };
};
