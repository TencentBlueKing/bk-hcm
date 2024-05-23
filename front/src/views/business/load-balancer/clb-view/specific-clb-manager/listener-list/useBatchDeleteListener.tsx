import { Ref, computed, reactive, ref } from 'vue';
import { Button, Message, Tag } from 'bkui-vue';
import { Column } from 'bkui-vue/lib/table/props';
import { useResourceStore } from '@/store';
import { cloneDeep } from 'lodash';

export default (
  columns: Array<Column>,
  selections: Ref<any[]>,
  resetSelections: (...args: any) => any,
  getListData: (...args: any) => any,
  isDomain = false,
) => {
  // use stores
  const resourceStore = useResourceStore();

  const isSubmitLoading = ref(false);
  const isBatchDeleteDialogShow = ref(false);
  const radioGroupValue = ref(false);
  const renderColumn = isDomain
    ? columns.slice(1, 5)
    : [
        ...columns.slice(0, 5),
        {
          label: '是否绑定目标组',
          field: 'target_group_id',
          render: ({ cell }: { cell: string }) => {
            if (cell)
              return (
                <Tag theme='success' v-bk-tooltips={{ content: cell }}>
                  已绑定
                </Tag>
              );
            return <Tag>未绑定</Tag>;
          },
        },
        {
          label: 'RS权重为0',
          field: 'rs_weight',
          render: ({ data }: any) => {
            const { rs_weight_zero_num, rs_weight_non_zero_num } = data;
            return (
              <div class='rs-weight-col'>
                <span class={rs_weight_zero_num ? 'exception' : 'normal'}>{rs_weight_zero_num}</span>/
                <span>{rs_weight_zero_num + rs_weight_non_zero_num}</span>
              </div>
            );
          },
        },
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
      ];
  const tableProps = reactive({
    columns: renderColumn,
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
        id: 'binding_status',
      },
    ],
  });

  const computedListenersList = computed(() => {
    if (radioGroupValue.value)
      return tableProps.data.filter(({ rs_weight_non_zero_num }: any) => rs_weight_non_zero_num === 0);
    return tableProps.data.filter(({ rs_weight_non_zero_num }: any) => rs_weight_non_zero_num > 0);
  });

  // 如果没有可删除的负载均衡, 则禁用删除按钮
  const isSubmitDisabled = computed(
    () => tableProps.data.filter(({ rs_weight_non_zero_num }: any) => rs_weight_non_zero_num === 0).length === 0,
  );

  // click-handler - 批量删除监听器
  const handleBatchDeleteListener = () => {
    isBatchDeleteDialogShow.value = true;
    tableProps.data = cloneDeep(selections.value);
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
        ids: tableProps.data
          // 只删除rs权重全部为零的监听器
          .filter(({ rs_weight_non_zero_num }: any) => rs_weight_non_zero_num === 0)
          .map((item) => item.id),
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
    isSubmitLoading,
    isSubmitDisabled,
    isBatchDeleteDialogShow,
    radioGroupValue,
    tableProps,
    handleBatchDeleteListener,
    handleBatchDeleteSubmit,
    computedListenersList,
  };
};
