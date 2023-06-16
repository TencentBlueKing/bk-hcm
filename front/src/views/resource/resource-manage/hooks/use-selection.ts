/**
 * 选择相关状态和事件
 */
import {
  ref,
} from 'vue';

type SelectionType = {
  checked: boolean;
  data: any[];
  isAll: boolean;
  row: any
};

export default () => {
  const selections = ref([]);

  const handleSelectionChange = (selection: SelectionType) => {
    // 全选
    if (selection.isAll && selection.checked) {
      selections.value = JSON.parse(JSON.stringify(selection.data));
    }
    // 取消全选
    if (selection.isAll && !selection.checked) {
      selections.value = [];
    }
    // 选择某一个
    if (!selection.isAll && selection.checked) {
      selections.value.push(selection.row);
    }
    // 取消选择某一个
    if (!selection.isAll && !selection.checked) {
      const index = selections.value.findIndex(item => item === selection.row);
      selections.value.splice(index, 1);
    }
  };

  const resetSelections = () => {
    selections.value = [];
  };

  return {
    selections,
    handleSelectionChange,
    resetSelections,
  };
};
