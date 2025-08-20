import { computed, ref, toRaw } from 'vue';

type SelectionType = {
  checked: boolean;
  data: any[];
  row?: any;
};

type UseTableSelectionParams = {
  rowKey?: string;
  isRowSelectable: (args: { row: SelectionType['row'] }) => boolean;
};

export default function useTableSelection({ rowKey = 'id', isRowSelectable }: UseTableSelectionParams) {
  const selections = ref([]);
  // 用于设置表格刷新后的默认勾选项
  const checked = computed(() => selections.value.map((item) => item[rowKey]));

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
      const index = selections.value.findIndex((item) => item[rowKey] === selection.row[rowKey]);
      if (index !== -1) {
        selections.value.splice(index, 1);
      }
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
    checked,
    resetSelections,
    handleSelectAll,
    handleSelectChange,
  };
}
