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
    let account_id = '';
    let vpc_ids = [] as string[];
    let tgPort = 0;
    let tableRsList = [] as any[];

    const handleShow = ({
      accountId,
      vpcIds,
      port,
      rsList,
      isCorsV2,
    }: {
      accountId: string;
      vpcIds: string[];
      port: number;
      rsList: any[];
      isCorsV2: boolean;
    }) => {
      isShow.value = true;
      nextTick(handleClear);
      account_id = accountId;
      vpc_ids = vpcIds;
      tgPort = port;
      tableRsList = rsList;

      // 根据account_id, vpc_ids查询cvm列表
      getRSTableList(accountId, vpc_ids, isCorsV2);
    };

    // confirm-handler
    const handleAddRs = () => {
      // 初始化选中的rs列表
      const selectedRsList = rsSelections.value.map((item) => ({
        ...item,
        port: tgPort || '',
        weight: 10,
        isNew: true,
      }));

      // 更新目标组 - 场景标识
      if (!loadBalancerStore.currentScene) {
        loadBalancerStore.setUpdateCount(2);
        loadBalancerStore.setCurrentScene('AddRs');
      }

      switch (loadBalancerStore.currentScene) {
        case 'add':
        case 'edit':
        case 'AddRs':
          bus.$emit('updateSelectedRsList', selectedRsList);
          break;
        case 'BatchAddRs':
          // 显示批量添加rs的sideslider
          bus.$emit('showBatchAddRsSideslider', { accountId: account_id, vpcId: vpc_ids, selectedRsList });
        default:
          break;
      }
    };

    const {
      searchData,
      searchVal,
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
    } = useAddRsTable(
      rsSelections,
      () => tableRsList,
      () => ({
        vpc_ids,
        account_id,
      }),
    );

    onMounted(() => {
      bus.$on('showAddRsDialog', handleShow);
    });

    onUnmounted(() => {
      bus.$off('showAddRsDialog');
    });

    return () => (
      <CommonDialog v-model:isShow={isShow.value} title='添加 RS' width={640} onHandleConfirm={handleAddRs}>
        <div class='add-rs-dialog-content'>
          <SearchSelect class='mb16' v-model={searchVal.value} data={searchData} />
          <Loading loading={isTableLoading.value} class='loading-table-container'>
            <Table
              class='table-container'
              ref={tableRef}
              columns={columns}
              data={rsTableList.value}
              pagination={pagination}
              remotePagination
              onSelect={handleSelect}
              onSelectAll={handleSelectAll}
              isRowSelectEnable={({ row }: any) =>
                !tableRsList.some((rs) => rs.id === row.id || rs.inst_id === row.id)
              }>
              {{
                prepend: () =>
                  rsTableList.value.length && selectedCount.value ? (
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
