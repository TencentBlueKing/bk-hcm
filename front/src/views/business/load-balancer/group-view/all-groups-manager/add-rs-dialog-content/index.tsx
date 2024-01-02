import { defineComponent, reactive, ref } from 'vue';
import { Loading, SearchSelect, Table } from 'bkui-vue';
import Empty from '@/components/empty';
import './index.scss';

export default defineComponent({
  name: 'AddRsDialogContent',
  setup() {
    const searchValue = ref('');
    const isTableLoading = ref(false);
    const tableRef = ref(null);
    const columns = [
      { type: 'selection', width: '100' },
      {
        label: '内网IP',
        field: 'privateIp',
      },
      {
        label: '公网IP',
        field: 'publicIp',
      },
      {
        label: '名称',
        field: 'name',
      },
      {
        label: '资源类型',
        field: 'resourceType',
        filter: true,
      },
    ];
    const tableData = [
      {
        privateIp: '192.168.10.1',
        publicIp: '52.14.72.95',
        name: '应用服务器1',
        resourceType: 'EC2',
      },
      {
        privateIp: '192.168.10.2',
        publicIp: '52.14.72.96',
        name: '数据库服务器1',
        resourceType: 'RDS',
      },
      {
        privateIp: '192.168.10.3',
        publicIp: '52.14.72.97',
        name: '缓存服务器1',
        resourceType: 'ElastiCache',
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
    const handleSelect = ({ checked }: { row: any; index: Number; checked: boolean; data: Array<any> }) => {
      if (checked) {
        selectedCount.value += 1;
      } else {
        selectedCount.value -= 1;
      }
    };
    const handleSelectAll = ({ checked, data }: { checked: boolean; data: Array<any> }) => {
      if (checked) {
        selectedCount.value = data.length > paginationConfig.limit ? paginationConfig.limit : data.length;
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
            data={tableData}
            pagination={paginationConfig}
            onSelect={handleSelect}
            onSelectAll={handleSelectAll}>
            {{
              prepend: () => (
                <div class='table-prepend-wrap'>
                  当前已选择 <span class='number'>{selectedCount.value}</span> 条数据,&nbsp;
                  <span class='clear' onClick={handleClear}>
                    清除选择
                  </span>
                </div>
              ),
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
