import { computed, defineComponent, ref, watch } from 'vue';
// import components
import { SearchSelect, Loading, Table, Input, Button, Form } from 'bkui-vue';
import Empty from '@/components/empty';
import BatchUpdatePopConfirm from '@/components/batch-update-popconfirm';
// import hooks
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
// import stores
import { useLoadBalancerStore } from '@/store';
import { useRegionsStore } from '@/store/useRegionsStore';
// import utils
import bus from '@/common/bus';
import { getLocalFilterConditions } from '@/utils';
import './index.scss';

const { FormItem } = Form;

export default defineComponent({
  name: 'RsConfigTable',
  props: {
    rsList: Array<any>,
    deletedRsList: Array<any>,
    accountId: String,
    vpcId: String,
    port: Number,
    noSearch: Boolean,
    noDisabled: Boolean, // 禁用所有disabled
    onlyShow: Boolean, // 只用于显示(基本信息页面使用)
    lbDetail: Object,
  },
  emits: ['update:rsList', 'update:deletedRsList'],
  setup(props, { emit }) {
    // use stores
    const loadBalancerStore = useLoadBalancerStore();
    const regionsStore = useRegionsStore();
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
    const handleDeleteRs = (data: any) => {
      const { id, isNew } = data;
      // 如果待移除的rs不是新增的, 而是目标组已经绑定的, 则记录操作场景, 并记录待删除的rs
      if (!isNew) {
        loadBalancerStore.setCurrentScene('BatchDeleteRs');
        loadBalancerStore.setUpdateCount(2);
        emit('update:deletedRsList', [...props.deletedRsList, data]);
      }
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
                disabledTip='目标组基本信息修改，添加，RS权重批量修改，RS端口批量修改，RS批量移除等操作暂不支持同时执行'
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
        render: ({ cell, data, index }: { cell: number; data: any; index: number }) => {
          const port = loadBalancerStore.listenerDetailWithTargetGroup?.end_port
            ? `${cell}-${
                cell +
                loadBalancerStore.listenerDetailWithTargetGroup?.end_port -
                loadBalancerStore.listenerDetailWithTargetGroup?.port
              }`
            : cell;

          if (props.onlyShow) return port;
          return (
            <FormItem
              property={`rs_list.${index}.port`}
              errorDisplayType='tooltips'
              required
              rules={[
                { validator: (v: number) => v >= 1 && v <= 65535, message: '端口范围为1-65535', trigger: 'change' },
              ]}>
              <Input
                modelValue={port}
                onChange={handleUpdate(data.id, 'port')}
                disabled={!props.noDisabled && !(isAdd.value || (isAddRs.value && data.isNew))}
              />
            </FormItem>
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
                disabledTip='目标组基本信息修改，添加，RS权重批量修改，RS端口批量修改，RS批量移除等操作暂不支持同时执行'
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
        render: ({ cell, data, index }: { cell: number; data: any; index: number }) => {
          if (props.onlyShow) return cell;
          return (
            <FormItem
              property={`rs_list.${index}.weight`}
              errorDisplayType='tooltips'
              required
              rules={[{ validator: (v: number) => v >= 0 && v <= 100, message: '权重范围为0-100', trigger: 'change' }]}>
              <Input
                modelValue={cell}
                onChange={handleUpdate(data.id, 'weight')}
                disabled={!props.noDisabled && !(isAdd.value || (isAddRs.value && data.isNew))}
                type='number'
                class='no-number-control'
              />
            </FormItem>
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
          <Button text onClick={() => handleDeleteRs(data)}>
            <i class='hcm-icon bkhcm-icon-minus-circle-shape'></i>
          </Button>
        ),
      });
    // 补充 port 和 weight 的 settings 配置
    settings.value.checked.push('port', 'weight');
    settings.value.fields.push({ label: '端口', field: 'port' }, { label: '权重', field: 'weight' });

    // click-handler - 添加rs
    const handleAddRs = () => {
      bus.$emit('showAddRsDialog', {
        accountId: props.accountId,
        vpcIds: [vpc_id.value],
        port: props.port,
        rsList: props.rsList,
        isCorsV2: props.lbDetail?.extension?.snat_pro,
      });
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

    const searchData = computed(() => {
      const tmpArr = [
        {
          id: 'private_ip_address',
          name: '内网IP',
        },
        {
          id: 'public_ip_address',
          name: '公网IP',
        },
        {
          id: 'inst_name',
          name: '名称',
        },
        {
          id: 'region',
          name: '地域',
        },
        {
          id: 'inst_type',
          name: '资源类型',
        },
        {
          id: 'cloud_vpc_ids',
          name: '所属VPC',
        },
        {
          id: 'port',
          name: '端口',
        },
        {
          id: 'weight',
          name: '权重',
        },
      ];
      if (!props.onlyShow)
        tmpArr.splice(
          0,
          2,
          { id: 'private_ipv4_addresses', name: '内网IP' },
          { id: 'public_ipv4_addresses', name: '公网IP' },
        );
      return tmpArr;
    });
    const searchValue = ref();

    // 监听 searchValue 的变化，根据过滤条件过滤得到 实际用于渲染的数据
    const renderTableData = computed(() => {
      const filterConditions = getLocalFilterConditions(searchValue.value, (rule) => {
        switch (rule.id) {
          case 'region':
            return regionsStore.getRegionNameEN(rule.values[0].id);
          default:
            return rule.values[0].id;
        }
      });

      return props.rsList?.filter((item) =>
        Object.keys(filterConditions).every((key) => {
          switch (key) {
            case 'private_ip_address':
            case 'private_ipv4_addresses':
            case 'public_ip_address':
            case 'public_ipv4_addresses':
            case 'cloud_vpc_ids':
              return filterConditions[key].includes(item[key][0]);
            case 'port':
            case 'weight':
              return filterConditions[key].includes(`${item[key]}`);
            default:
              return filterConditions[key].includes(item[key]);
          }
        }),
      );
    });

    return () => (
      <div class='rs-config-table'>
        <div class={`rs-config-operation-wrap${props.onlyShow ? ' jc-right' : ''}`}>
          {props.onlyShow ? null : (
            <Button
              class='left-wrap'
              text
              theme='primary'
              onClick={handleAddRs}
              v-bk-tooltips={{
                content: '目标组基本信息，RS变更，RS权重修改，RS端口修改不支持同时变更',
                disabled: isInitialState.value || isAddRs.value,
              }}
              disabled={!isInitialState.value && !isAddRs.value}>
              <i class='hcm-icon bkhcm-icon-plus-circle-shape'></i>
              <span>添加 RS</span>
            </Button>
          )}
          {props.noSearch ? null : (
            <div class='search-wrap'>
              <SearchSelect class='table-search-select' v-model={searchValue.value} data={searchData.value} />
            </div>
          )}
        </div>
        <Loading loading={isTableLoading.value}>
          <Table data={renderTableData.value} columns={rsTableColumns} settings={settings.value} showOverflowTooltip>
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
