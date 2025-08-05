import { cloneDeep } from 'lodash';
import { Ref, computed, reactive, ref } from 'vue';
import { Column } from 'bkui-vue/lib/table/props';
import { useResourceStore } from '@/store';
import { Message } from 'bkui-vue';
import bus from '@/common/bus';
import { LB_NETWORK_TYPE_MAP } from '@/constants';

export default (columns: Array<Column>, selections: Ref<any[]>, getListData: (...args: any) => any) => {
  const resourceStore = useResourceStore();
  const isBatchDeleteDialogShow = ref(false);
  const isSubmitLoading = ref(false);
  const radioGroupValue = ref(true);

  const tableProps = reactive({
    columns,
    data: [],
    searchData: [
      { id: 'name', name: '负载均衡名称' },
      { id: 'cloud_id', name: '负载均衡ID' },
      { id: 'domain', name: '负载均衡域名' },
      { id: 'lb_vip', name: '负载均衡VIP' },
      {
        id: 'lb_type',
        name: '网络类型',
        children: Object.keys(LB_NETWORK_TYPE_MAP).map((lbType) => ({
          id: lbType,
          name: LB_NETWORK_TYPE_MAP[lbType as keyof typeof LB_NETWORK_TYPE_MAP],
        })),
      },
    ],
  });

  const computedListenersList = computed(() => {
    if (radioGroupValue.value)
      return tableProps.data.filter(({ listenerNum, delete_protect }: any) => !(listenerNum || delete_protect));
    return tableProps.data.filter(({ listenerNum, delete_protect }: any) => listenerNum > 0 || delete_protect);
  });

  // 如果没有可删除的负载均衡, 则禁用删除按钮
  const isSubmitDisabled = computed(
    () =>
      tableProps.data.filter(({ listenerNum, delete_protect }: any) => !(listenerNum || delete_protect)).length === 0,
  );

  // click-handler
  const handleClickBatchDelete = () => {
    tableProps.data = cloneDeep(selections.value);
    radioGroupValue.value = true;
    if (
      tableProps.data.filter(({ listenerNum, delete_protect }: any) => listenerNum > 0 || delete_protect).length > 0
    ) {
      radioGroupValue.value = false;
    }
    isBatchDeleteDialogShow.value = true;
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
        ids: tableProps.data
          .filter(({ listenerNum, delete_protect }: any) => !(listenerNum || delete_protect))
          .map((item) => item.id),
      });
      Message({ theme: 'success', message: '批量删除成功' });
      isBatchDeleteDialogShow.value = false;
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
