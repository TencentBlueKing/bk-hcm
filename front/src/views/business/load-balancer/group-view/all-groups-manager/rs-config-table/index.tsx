import { defineComponent, ref } from 'vue';
import { SearchSelect, Loading, Table } from 'bkui-vue';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import Empty from '@/components/empty';
import BatchUpdatePopconfirm from '@/components/batch-update-popconfirm';
import './index.scss';

export default defineComponent({
  name: 'RsConfigTable',
  props: {
    noOperation: Boolean,
    noSearch: Boolean,
  },
  emits: ['showAddRsDialog'],
  setup(props, { emit }) {
    // rs配置表单项
    const isTableLoading = ref(false);
    const { columns, settings } = useColumns('rsConfig');
    const handleBatchUpdatePort = (_port: number) => {};
    const handleBatchUpdateWeight = (_weight: number) => {};
    const handleDeleteRs = () => {};
    const rsTableColumns = [
      ...columns,
      {
        label: () => (
          <>
            <span>端口</span>
            <BatchUpdatePopconfirm title='端口' onUpdateValue={handleBatchUpdatePort}></BatchUpdatePopconfirm>
          </>
        ),
        field: 'port',
        isDefaultShow: true,
      },
      {
        label: () => (
          <>
            <span>权重</span>
            <BatchUpdatePopconfirm title='权重' onUpdateValue={handleBatchUpdateWeight}></BatchUpdatePopconfirm>
          </>
        ),
        field: 'weight',
        isDefaultShow: true,
      },
      {
        label: '',
        width: 80,
        render: () => <i class='hcm-icon bkhcm-icon-minus-circle-shape' onClick={handleDeleteRs}></i>,
      },
    ];
    // 补充 port 和 weight 的 settings 配置
    settings.value.checked.push('port', 'weight');
    settings.value.fields.push({ label: '端口', field: 'port' }, { label: '权重', field: 'weight' });
    const rsConfigData = [
      {
        privateIp: '10.0.0.1',
        publicIp: '203.0.113.10',
        name: '服务器A',
        region: '华北1区',
        resourceType: 'VM',
        network: 'VPC-XYZ',
        port: 8080,
        weight: 20,
      },
      {
        privateIp: '10.0.1.2',
        publicIp: '203.0.113.20',
        name: '数据库B',
        region: '华东2区',
        resourceType: 'RDS',
        network: 'VPC-ABC',
        port: 3306,
        weight: 10,
      },
      {
        privateIp: '10.0.2.3',
        publicIp: '203.0.113.30',
        name: '负载均衡C',
        region: '华南3区',
        resourceType: 'LoadBalancer',
        network: 'VPC-DEF',
        port: 80,
        weight: 30,
      },
    ];
    return () => (
      <div class='rs-config-table'>
        <div class={`rs-config-operation-wrap${props.noOperation ? ' jc-right' : ''}`}>
          {props.noOperation ? null : (
            <div class='left-wrap' onClick={() => emit('showAddRsDialog')}>
              <i class='hcm-icon bkhcm-icon-plus-circle-shape'></i>
              <span>添加 RS</span>
            </div>
          )}
          {props.noSearch ? null : (
            <div class='search-wrap'>
              <SearchSelect></SearchSelect>
            </div>
          )}
        </div>
        <Loading loading={isTableLoading.value}>
          <Table data={rsConfigData} columns={rsTableColumns} settings={settings.value} showOverflowTooltip>
            {{
              empty: () => {
                if (isTableLoading.value) return null;
                return <Empty text='暂未添加实例' />;
              },
            }}
          </Table>
        </Loading>
      </div>
    );
  },
});
