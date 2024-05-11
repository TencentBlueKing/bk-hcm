import { cloneDeep } from 'lodash';
import { Ref, computed, reactive, ref } from 'vue';
import { Column } from 'bkui-vue/lib/table/props';
import { useResourceStore } from '@/store';
import { Message } from 'bkui-vue';
import bus from '@/common/bus';

export default (
  columns: Array<Column>,
  selections: Ref<any[]>,
  resetSelections: (...args: any) => any,
  getListData: (...args: any) => any,
) => {
  const resourceStore = useResourceStore();
  const isBatchDeleteDialogShow = ref(false);
  const isSubmitLoading = ref(false);
  const radioGroupValue = ref(false);

  const tableProps = reactive({
    columns,
    data: [],
    searchData: [],
  });

  const computedListenersList = computed(() => {
    if (radioGroupValue.value) return tableProps.data.filter(({ listenerNum }: any) => listenerNum === undefined);
    return tableProps.data.filter(({ listenerNum }: any) => listenerNum > 0);
  });

  // 如果没有可删除的负载均衡, 则禁用删除按钮
  const isSubmitDisabled = computed(
    () => tableProps.data.filter(({ listenerNum }: any) => listenerNum === undefined).length === 0,
  );

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
        // 只删除没有监听器的负载均衡
        ids: tableProps.data.filter(({ listenerNum }: any) => listenerNum === undefined).map((item) => item.id),
      });
      Message({ theme: 'success', message: '批量删除成功' });
      isBatchDeleteDialogShow.value = false;
      resetSelections();
      getListData();
      // 重新拉取lb-tree数据
      bus.$emit('resetLbTree');
    } finally {
      isSubmitLoading.value = false;
    }
  };

  return {
    isBatchDeleteDialogShow,
    isSubmitLoading,
    isSubmitDisabled,
    radioGroupValue,
    tableProps,
    handleClickBatchDelete,
    handleRemoveSelection,
    handleBatchDeleteSubmit,
    computedListenersList,
  };
};