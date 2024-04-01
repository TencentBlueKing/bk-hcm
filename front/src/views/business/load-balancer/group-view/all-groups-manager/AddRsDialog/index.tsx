import { defineComponent, onMounted, onUnmounted, ref } from 'vue';
// import components
import { Loading, SearchSelect, Table } from 'bkui-vue';
import CommonDialog from '@/components/common-dialog';
import Empty from '@/components/empty';
// import stores
import { useLoadBalancerStore } from '@/store';
// import hooks
import useAddRsTable from './useAddRsTable';
// import utils
import bus from '@/common/bus';
import './index.scss';

export default defineComponent({
  name: 'AddRsDialog',
  setup() {
    // 添加RS
    const loadBalancerStore = useLoadBalancerStore();

    const isShow = ref(false);
    const rsTableList = ref([]);
    const rsSelections = ref([]);

    const handleShow = (data: any) => {
      isShow.value = true;
      rsTableList.value = data;
    };

    // confirm-handler
    const handleAddRs = () => {
      // 根据不同的场景, 判断是否要显示批量添加rs的sideslider
      if (loadBalancerStore.currentScene === 'batchAddRs') {
        // todo
        bus.$emit('showBatchAddRsDialog');
      } else {
        // todo
      }
      // bus.$emit('updateSelectedRsList', rsSelections.value);
    };

    const {
      searchData,
      searchValue,
      isTableLoading,
      tableRef,
      columns,
      pagination,
      selectedCount,
      handleSelect,
      handleSelectAll,
      handleClear,
    } = useAddRsTable(rsSelections);

    onMounted(() => {
      bus.$on('showAddRsDialog', handleShow);
    });

    onUnmounted(() => {
      bus.$off('showAddRsDialog');
    });

    return () => (
      <CommonDialog v-model:isShow={isShow.value} title='添加 RS' width={640} onHandleConfirm={handleAddRs}>
        <div class='add-rs-dialog-content'>
          <SearchSelect class='mb16' v-model={searchValue.value} data={searchData} />
          <Loading loading={isTableLoading.value} class='loading-table-container'>
            <Table
              class='table-container'
              ref={tableRef}
              columns={columns}
              data={rsTableList.value}
              pagination={pagination}
              onSelect={handleSelect}
              onSelectAll={handleSelectAll}>
              {{
                prepend: () =>
                  rsTableList.value.length ? (
                    <div class='table-prepend-wrap'>
                      当前已选择 <span class='number'>{selectedCount.value}</span> 条数据,&nbsp;
                      <span class='clear' onClick={handleClear}>
                        清除选择
                      </span>
                    </div>
                  ) : null,
                empty: () => {
                  if (isTableLoading.value) return null;
                  return <Empty />;
                },
              }}
            </Table>
          </Loading>
        </div>
      </CommonDialog>
    );
  },
});
