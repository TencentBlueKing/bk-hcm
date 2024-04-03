import { defineComponent, nextTick, onMounted, onUnmounted, ref } from 'vue';
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
    const rsSelections = ref([]);
    const accountId = ref('');

    const handleShow = (account_id: string) => {
      isShow.value = true;
      nextTick(handleClear);
      accountId.value = account_id;
      getRSTableList(accountId.value);
    };

    // confirm-handler
    const handleAddRs = () => {
      // 将选中的rs列表添加到store中
      loadBalancerStore.setSelectedRsList(
        rsSelections.value.map((item) => ({
          ...item,
          port: 0,
          weight: 0,
          isNew: true,
        })),
      );
      // 根据不同的场景, 判断是否要显示批量添加rs的sideslider
      if (loadBalancerStore.currentScene === 'BatchAddRs') {
        bus.$emit('showBatchAddRsSideslider', accountId.value);
      }
      // 更新目标组 - 场景标识
      if (!loadBalancerStore.currentScene) {
        loadBalancerStore.setUpdateCount(2);
        loadBalancerStore.setCurrentScene('AddRs');
      }
    };

    const {
      searchData,
      searchValue,
      isTableLoading,
      tableRef,
      columns,
      rsTableList,
      pagination,
      selectedCount,
      handleSelect,
      handleSelectAll,
      handleClear,
      getRSTableList,
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
              remotePagination
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
