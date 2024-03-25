import { defineComponent, reactive, ref } from 'vue';
import { Loading, SearchSelect, Table } from 'bkui-vue';
import Empty from '@/components/empty';
import './index.scss';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';

export default defineComponent({
  name: 'AddRsDialogContent',
  props: {
    rsList: {
      type: Array,
      required: true,
    },
    rsTableData: {
      type: Array,
      required: true,
    },
  },
  emits: ['select'],
  setup(props, { emit }) {
    const searchValue = ref('');
    const isTableLoading = ref(false);
    const tableRef = ref(null);
    const { selections, handleSelectionChange } = useSelection();
    const columns = [
      { type: 'selection', width: '100' },
      {
        label: '内网IP',
        render({ data }: any) {
          return [...data.private_ipv4_addresses, ...data.private_ipv6_addresses].join(',');
        },
      },
      {
        label: '公网IP',
        render({ data }: any) {
          return [...data.public_ipv4_addresses, ...data.public_ipv6_addresses].join(',');
        },
      },
      {
        label: '名称',
        field: 'name',
      },
      {
        label: '资源类型',
        field: 'machine_type',
        filter: true,
      },
    ];
    const searchData = [
      {
        name: '内网IP',
        id: 'privateIp',
      },
      {
        name: '公网IP',
        id: 'publicIp',
      },
      {
        name: '名称',
        id: 'name',
      },
      {
        name: '资源类型',
        id: 'resourceType',
      },
    ];
    const selectedCount = ref(0);
    const handleSelect = (selection: any) => {
      handleSelectionChange(selection, () => true, false);
      emit('select', selections.value);
      if (selection.checked) {
        selectedCount.value += 1;
      } else {
        selectedCount.value -= 1;
      }
    };
    const handleSelectAll = (selection: any) => {
      handleSelectionChange(selection, () => true, true);
      emit('select', selections.value);
      if (selection.checked) {
        selectedCount.value =
          selection.data.length > paginationConfig.limit ? paginationConfig.limit : selection.data.length;
      } else {
        selectedCount.value = 0;
      }
    };
    const handleClear = () => {
      tableRef.value.clearSelection();
      selectedCount.value = 0;
    };
    const paginationConfig = reactive({ small: true, align: 'left', limit: 10, limitList: [10, 20, 50, 100] });
    return () => (
      <div class='add-rs-dialog-content'>
        <SearchSelect class='mb16' v-model={searchValue.value} data={searchData} />
        <Loading loading={isTableLoading.value} class='loading-table-container'>
          <Table
            class='table-container'
            ref={tableRef}
            columns={columns}
            data={props.rsList}
            pagination={paginationConfig}
            onSelect={handleSelect}
            onSelectAll={handleSelectAll}>
            {{
              prepend: () =>
                props.rsList.length ? (
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
    );
  },
});
