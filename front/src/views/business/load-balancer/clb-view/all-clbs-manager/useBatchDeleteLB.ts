import { cloneDeep } from 'lodash';
import { Ref, reactive, ref } from 'vue';
import { Column } from 'bkui-vue/lib/table/props';
import { useResourceStore } from '@/store';
import { Message } from 'bkui-vue';

export default (
  columns: Array<Column>,
  selections: Ref<any[]>,
  resetSelections: (...args: any) => any,
  getListData: (...args: any) => any,
) => {
  const resourceStore = useResourceStore();
  const isBatchDeleteDialogShow = ref(false);
  const isSubmitLoading = ref(false);
  const radioGroupValue = ref(true);

  const tableProps = reactive({
    columns,
    data: [],
  });

  // click-handler
  const handleClickBatchDelete = () => {
    isBatchDeleteDialogShow.value = true;
    tableProps.data = cloneDeep(selections.value);
  };

  // remove-handler - 移除单个监听器
  const handleRemoveSelection = (id: string) => {
    const idx = tableProps.data.findIndex((item) => item.id === id);
    tableProps.data.splice(idx, 1);
  };

  // submit-handler
  const handleBatchDeleteSubmit = async () => {
    try {
      isSubmitLoading.value = true;
      await resourceStore.deleteBatch('load_balancers', {
        ids: tableProps.data.map((item) => item.id),
      });
      Message({ theme: 'success', message: '批量删除成功' });
      isBatchDeleteDialogShow.value = false;
      resetSelections();
      getListData();
    } finally {
      isSubmitLoading.value = false;
    }
  };

  return {
    isBatchDeleteDialogShow,
    isSubmitLoading,
    radioGroupValue,
    tableProps,
    handleClickBatchDelete,
    handleRemoveSelection,
    handleBatchDeleteSubmit,
  };
};
