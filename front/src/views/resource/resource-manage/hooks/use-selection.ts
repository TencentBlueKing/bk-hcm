/**
 * 选择相关状态和事件
 */
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { useAccountStore } from '@/store';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { ref, watch } from 'vue';

type SelectionType = {
  checked: boolean;
  data: any[];
  isAll: boolean;
  row: any;
};

export default () => {
  const selections = ref([]);
  const { whereAmI } = useWhereAmI();
  const resourceAccountStore = useResourceAccountStore();
  const accountStore = useAccountStore();

  watch(
    () => resourceAccountStore.resourceAccount,
    () => {
      if (whereAmI.value !== Senarios.resource) return;
      resetSelections();
    },
    {
      deep: true,
    },
  );

  watch(
    () => accountStore.bizs,
    () => {
      if (whereAmI.value !== Senarios.business) return;
      resetSelections();
    },
  );

  const handleSelectionChange = (selection: SelectionType, isCurRowSelectEnable: (row: any) => void, isAll = false) => {
    console.log(666, selection);
    // 全选
    if (isAll && selection.checked) {
      selections.value = JSON.parse(JSON.stringify(selection.data));
      selections.value = selections.value.filter((row) => isCurRowSelectEnable(row));
      console.log(66666, selections.value);
    }
    // 取消全选
    if (isAll && !selection.checked) {
      selections.value = [];
    }
    // 选择某一个
    if (!isAll && selection.checked) {
      selections.value.push(selection.row);
    }
    // 取消选择某一个
    if (!isAll && !selection.checked) {
      const index = selections.value.findIndex((item) => item === selection.row);
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
