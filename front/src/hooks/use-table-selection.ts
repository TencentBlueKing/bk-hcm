import { ref, toRaw } from 'vue';

type SelectionType = {
  checked: boolean;
  data: any[];
  row?: any;
};

type UseTableSelectionParams = {
  isRowSelectable: (args: { row: SelectionType['row'] }) => boolean;
};

export default function useTableSelection({ isRowSelectable }: UseTableSelectionParams) {
  const selections = ref([]);

  const handleSelectChange = (selection: SelectionType, isAll = false) => {
    // 全选
    if (isAll && selection.checked) {
      const selectionClone = structuredClone(toRaw(selection.data));
      selections.value = selectionClone.filter((row) => isRowSelectable({ row }));
    }
    // 取消全选
    if (isAll && !selection.checked) {
      selections.value = [];
    }
    // 选择某一个
    if (!isAll && selection.checked) {
      selections.value.push(structuredClone(toRaw(selection.row)));
    }
    // 取消选择某一个
    if (!isAll && !selection.checked) {
      const index = selections.value.findIndex((item) => item === selection.row);
      selections.value.splice(index, 1);
    }
  };

  const handleSelectAll = (selection: SelectionType) => {
    handleSelectChange(selection, true);
  };

  const resetSelections = () => {
    selections.value = [];
  };

  return {
    selections,
    resetSelections,
    handleSelectAll,
    handleSelectChange,
  };
}
