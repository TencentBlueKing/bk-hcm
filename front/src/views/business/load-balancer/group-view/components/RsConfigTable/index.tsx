import { computed, defineComponent, ref, watch } from 'vue';
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
    vpcId: String,
    port: Number,
    noSearch: Boolean,
    noDisabled: Boolean, // 禁用所有disabled
    onlyShow: Boolean, // 只用于显示(基本信息页面使用)
  },
  emits: ['update:rsList'],
  setup(props, { emit }) {
    // use stores
    const loadBalancerStore = useLoadBalancerStore();
    const vpc_id = ref('');

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
      return (val: number) => {
        emit(
          'update:rsList',
          props.rsList.map((item) => {
            if (item.id === id) {
              item[key] = val;
            }
            return item;
          }),
        );
      };
    };

    // 修改所有row的port/weight
    const handleBatchUpdate = (v: number, key: 'port' | 'weight') => {
      if (loadBalancerStore.updateCount === 1 && !loadBalancerStore.currentScene) {
        loadBalancerStore.setUpdateCount(2);
        loadBalancerStore.setCurrentScene(key);
      }

      switch (loadBalancerStore.currentScene) {
        // 新增rs只支持批量修改新增的rs
        case 'AddRs':
          emit(
            'update:rsList',
            props.rsList.map((item) => {
              if (item.isNew) {
                item[key] = v;
              }
              return item;
            }),
          );
          break;
        case 'add':
        case 'port':
        case 'weight':
        case 'BatchAddRs':
          emit(
            'update:rsList',
            props.rsList.map((item) => {
              item[key] = v;
              return item;
            }),
          );
          break;
        default:
          break;
      }
    };

    // delete-handler
    const handleDeleteRs = (id: string) => {
      emit(
        'update:rsList',
        // 本期暂时以id来区分rs, 后续可能会变更为ip+port
        // 如果变更为ip+port, 则后端在cvm/list以及target_groups/detail接口中提供rs的「ip类型」字段或「统一的ip地址」字段, 前端处理起来会方便一点
        props.rsList.filter((item) => item.id !== id),
      );
    };

    const rsTableColumns = [
      ...columns,
      {
        label: () => {
          if (props.onlyShow) return '端口';
          return (
            <>
              <span>端口</span>
              <BatchUpdatePopConfirm
                title='端口'
                min={1}
                max={65535}
                onUpdateValue={(v) => handleBatchUpdate(v, 'port')}
                disabled={!props.noDisabled && !isInitialState.value && !isBatchUpdatePort.value && !isAddRs.value}
              />
            </>
          );
        },
        field: 'port',
        isDefaultShow: true,
        render: ({ cell, data }: { cell: number; data: any }) => {
          if (props.onlyShow) return cell;
          return (
            <Input
              modelValue={cell}
              onChange={handleUpdate(data.id, 'port')}
              disabled={!props.noDisabled && !(isAdd.value || (isAddRs.value && data.isNew))}
              type='number'
              min={1}
              max={65535}
              class='no-number-control'
              placeholder='1-65535'
            />
          );
        },
      },
      {
        label: () => {
          if (props.onlyShow) return '权重';
          return (
            <>
              <span>权重</span>
              <BatchUpdatePopConfirm
                title='权重'
                min={0}
                max={100}
                onUpdateValue={(v) => handleBatchUpdate(v, 'weight')}
                disabled={!props.noDisabled && !isInitialState.value && !isBatchUpdateWeight.value && !isAddRs.value}
              />
            </>
          );
        },
        field: 'weight',
        isDefaultShow: true,
        render: ({ cell, data }: { cell: number; data: any }) => {
          if (props.onlyShow) return cell;
          return (
            <Input
              modelValue={cell}
              onChange={handleUpdate(data.id, 'weight')}
              disabled={!props.noDisabled && !(isAdd.value || (isAddRs.value && data.isNew))}
              type='number'
              min={0}
              max={100}
              class='no-number-control'
              placeholder='0-100'
            />
          );
        },
      },
    ];
    // 如果组件仅用于显示, 则不需要操作列
    if (!props.onlyShow)
      rsTableColumns.push({
        label: '',
        width: 80,
        render: ({ data }: any) => (
          <Button text onClick={() => handleDeleteRs(data.id)} disabled={!props.noDisabled && !data.isNew}>
            <i class='hcm-icon bkhcm-icon-minus-circle-shape'></i>
          </Button>
        ),
      });
    // 补充 port 和 weight 的 settings 配置
    settings.value.checked.push('port', 'weight');
    settings.value.fields.push({ label: '端口', field: 'port' }, { label: '权重', field: 'weight' });

    // click-handler - 添加rs
    const handleAddRs = () => {
      bus.$emit('showAddRsDialog', { accountId: props.accountId, vpcId: vpc_id.value, port: props.port });
    };

    watch(
      () => props.vpcId,
      (val) => {
        vpc_id.value = val;
      },
      {
        immediate: true,
      },
    );

    return () => (
      <div class='rs-config-table'>
        <div class={`rs-config-operation-wrap${props.onlyShow ? ' jc-right' : ''}`}>
          {props.onlyShow ? null : (
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
