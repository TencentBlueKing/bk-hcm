import { Ref, reactive, ref, watch } from 'vue';
import { Button, Message } from 'bkui-vue';
import { Column } from 'bkui-vue/lib/table/props';
import { useResourceStore } from '@/store';
import { cloneDeep } from 'lodash';

export default (
  columns: Array<Column>,
  selections: Ref<any[]>,
  resetSelections: (...args: any) => any,
  getListData: (...args: any) => any,
) => {
  // use stores
  const resourceStore = useResourceStore();

  const isSubmitLoading = ref(false);
  const isBatchDeleteDialogShow = ref(false);
  const radioGroupValue = ref(false);
  const tableProps = reactive({
    columns: [
      ...columns.slice(0, 5),
      { label: '是否绑定目标组', field: '' },
      { label: 'RS权重为O', field: '' },
      {
        label: '',
        width: 50,
        minWidth: 50,
        render: ({ data }: any) => (
          <Button text onClick={() => handleRemoveSelection(data.id)}>
            <i class='hcm-icon bkhcm-icon-minus-circle-shape'></i>
          </Button>
        ),
      },
    ],
    data: [],
    searchData: [
      {
        name: '监听器名称',
        id: 'name',
      },
      {
        name: '协议',
        id: 'protocol',
      },
      {
        name: '端口',
        id: 'port',
      },
      {
        name: '均衡方式',
        id: 'scheduler',
      },
      {
        name: '域名数量',
        id: 'domain_num',
      },
      {
        name: 'URL数量',
        id: 'url_num',
      },
      {
        name: '同步状态',
        id: 'syncStatus',
      },
    ],
  });

  // click-handler - 批量删除监听器
  const handleBatchDeleteListener = () => {
    isBatchDeleteDialogShow.value = true;
  };

  // remove-handler - 移除单个监听器
  const handleRemoveSelection = (id: string) => {
    const idx = tableProps.data.findIndex((item) => item.id === id);
    tableProps.data.splice(idx, 1);
  };

  // submit
  const handleBatchDeleteSubmit = async () => {
    try {
      isSubmitLoading.value = true;
      await resourceStore.deleteBatch('listeners', {
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

  watch(
    selections,
    (val) => {
      tableProps.data = cloneDeep(val);
    },
    { deep: true },
  );

  return {
    isSubmitLoading,
    isBatchDeleteDialogShow,
    radioGroupValue,
    tableProps,
    handleBatchDeleteListener,
    handleBatchDeleteSubmit,
  };
};
