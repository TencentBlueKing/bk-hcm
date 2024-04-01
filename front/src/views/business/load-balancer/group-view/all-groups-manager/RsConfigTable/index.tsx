import { defineComponent, ref } from 'vue';
// import components
import { SearchSelect, Loading, Table, Input } from 'bkui-vue';
import Empty from '@/components/empty';
import BatchUpdatePopConfirm from '@/components/batch-update-popconfirm';
// import hooks
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import './index.scss';

export default defineComponent({
  name: 'RsConfigTable',
  props: {
    rsList: Array<any>,
    noOperation: Boolean,
    noSearch: Boolean,
  },
  emits: ['update:modelValue', 'showAddRsDialog'],
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
            <BatchUpdatePopConfirm title='端口' onUpdateValue={handleBatchUpdatePort} />
          </>
        ),
        field: 'port',
        isDefaultShow: true,
        render: () => <Input />,
      },
      {
        label: () => (
          <>
            <span>权重</span>
            <BatchUpdatePopConfirm title='权重' onUpdateValue={handleBatchUpdateWeight} />
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
          <Table data={props.rsList} columns={rsTableColumns} settings={settings.value} showOverflowTooltip>
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
