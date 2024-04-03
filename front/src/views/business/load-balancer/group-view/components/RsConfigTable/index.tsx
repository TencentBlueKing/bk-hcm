import { computed, defineComponent, ref } from 'vue';
// import components
import { SearchSelect, Loading, Table, Input, Button } from 'bkui-vue';
import Empty from '@/components/empty';
import BatchUpdatePopConfirm from '@/components/batch-update-popconfirm';
// import hooks
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
// import stores
import { useLoadBalancerStore } from '@/store';
// import utils
import bus from '@/common/bus';
import './index.scss';

export default defineComponent({
  name: 'RsConfigTable',
  props: {
    rsList: Array<any>,
    accountId: String,
    noOperation: Boolean,
    noSearch: Boolean,
    noDisabled: Boolean, // 禁用所有disabled
  },
  emits: ['update:rsList'],
  setup(props, { emit }) {
    // use stores
    const loadBalancerStore = useLoadBalancerStore();

    // rs配置表单项
    const isTableLoading = ref(false);
    const { columns, settings } = useColumns('rsConfig');

    const isInitialState = computed(() => loadBalancerStore.updateCount !== 2);
    const isAdd = computed(() => loadBalancerStore.currentScene === 'add');
    const isAddRs = computed(() => loadBalancerStore.currentScene === 'AddRs');
    const isBatchUpdatePort = computed(() => loadBalancerStore.currentScene === 'port');
    const isBatchUpdateWeight = computed(() => loadBalancerStore.currentScene === 'weight');

    // 修改单条row的port/weight
    const handleUpdate = (id: string, key: string) => {
      return (val: string) => {
        emit(
          'update:rsList',
          props.rsList.map((item) => {
            if (item.id === id) {
              item[key] = +val;
            }
            return item;
          }),
        );
      };
    };

    // 修改所有row的port/weight
    const handleBatchUpdate = (v: number, key: 'port' | 'weight') => {
      // 批量修改操作
      if (loadBalancerStore.updateCount === 1 && !loadBalancerStore.currentScene) {
        loadBalancerStore.setUpdateCount(2);
        loadBalancerStore.setCurrentScene(key);
        emit(
          'update:rsList',
          props.rsList.map((item) => {
            item[key] = v;
            return item;
          }),
        );
      }
      // 新增rs可支持批量修改新增的rs
      else {
        emit(
          'update:rsList',
          props.rsList.map((item) => {
            if (item.isNew) {
              item[key] = v;
            }
            return item;
          }),
        );
      }
    };

    const handleDeleteRs = () => {};

    const rsTableColumns = [
      ...columns,
      {
        label: () => (
          <>
            <span>端口</span>
            <BatchUpdatePopConfirm
              title='端口'
              onUpdateValue={(v) => handleBatchUpdate(v, 'port')}
              disabled={!props.noDisabled && !isInitialState.value && !isBatchUpdatePort.value && !isAddRs.value}
            />
          </>
        ),
        field: 'port',
        isDefaultShow: true,
        render: ({ cell, data }: { cell: number; data: any }) => (
          <Input
            modelValue={cell}
            onUpdate:modelValue={handleUpdate(data.id, 'port')}
            disabled={!props.noDisabled && !(isAdd.value || isBatchUpdatePort.value || (isAddRs.value && data.isNew))}
          />
        ),
      },
      {
        label: () => (
          <>
            <span>权重</span>
            <BatchUpdatePopConfirm
              title='权重'
              onUpdateValue={(v) => handleBatchUpdate(v, 'weight')}
              disabled={!props.noDisabled && !isInitialState.value && !isBatchUpdateWeight.value && !isAddRs.value}
            />
          </>
        ),
        field: 'weight',
        isDefaultShow: true,
        render: ({ cell, data }: { cell: number; data: any }) => (
          <Input
            modelValue={cell}
            onUpdate:modelValue={handleUpdate(data.id, 'weight')}
            disabled={!props.noDisabled && !(isAdd.value || isBatchUpdateWeight.value || (isAddRs.value && data.isNew))}
          />
        ),
      },
      {
        label: '',
        width: 80,
        render: ({ data }: any) => (
          <Button text onClick={handleDeleteRs} disabled={!props.noDisabled && !data.isNew}>
            <i class='hcm-icon bkhcm-icon-minus-circle-shape'></i>
          </Button>
        ),
      },
    ];
    // 补充 port 和 weight 的 settings 配置
    settings.value.checked.push('port', 'weight');
    settings.value.fields.push({ label: '端口', field: 'port' }, { label: '权重', field: 'weight' });

    // click-handler - 添加rs
    const handleAddRs = () => {
      bus.$emit('showAddRsDialog', props.accountId);
    };

    return () => (
      <div class='rs-config-table'>
        <div class={`rs-config-operation-wrap${props.noOperation ? ' jc-right' : ''}`}>
          {props.noOperation ? null : (
            <Button
              class='left-wrap'
              text
              theme='primary'
              onClick={handleAddRs}
              disabled={!isInitialState.value && !isAddRs.value}>
              <i class='hcm-icon bkhcm-icon-plus-circle-shape'></i>
              <span>添加 RS</span>
            </Button>
          )}
          {props.noSearch ? null : (
            <div class='search-wrap'>
              <SearchSelect />
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
