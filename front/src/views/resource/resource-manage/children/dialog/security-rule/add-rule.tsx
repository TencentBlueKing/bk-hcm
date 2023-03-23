import {
  defineComponent,
  ref,
  watch,
} from 'vue';
import { Table, Input, Select, Button } from 'bkui-vue'; // TagInput
import { Info } from 'bkui-vue/lib/icon';
import { ACTION_STATUS, GCP_PROTOCOL_LIST, IP_TYPE_LIST, HUAWEI_ACTION_STATUS, HUAWEI_TYPE_LIST, AZURE_PROTOCOL_LIST } from '@/constants';
import Confirm from '@/components/confirm';
import {
  useI18n,
} from 'vue-i18n';
import StepDialog from '@/components/step-dialog/step-dialog';
import {
  useResourceStore,
} from '@/store/resource';
import './add-rule.scss';
const { Option } = Select;


export default defineComponent({
  components: {
    StepDialog,
  },

  props: {
    title: {
      type: String,
    },
    isShow: {
      type: Boolean,
    },
    vendor: {
      type: String,
    },
    loading: {
      type: Boolean,
    },
    dialogWidth: {
      type: String,
    },
    activeType: {
      type: String,
    },
  },

  emits: ['update:isShow', 'submit'],

  setup(props, { emit }) {
    const {
      t,
    } = useI18n();

    const resourceStore = useResourceStore();

    // const cloudTargetSecurityGroup = ;

    const securityGroupSource = ref([   // 华为源
      {
        id: 'remote_ip_prefix',
        name: t('IP地址'),
      },
      {
        id: 'cloud_remote_group_id',
        name: t('安全组'),
      },
    ]);

    const azureSecurityGroupSource = ref([    // 微软云源
      {
        id: 'source_address_prefix',
        name: t('IP地址'),
      },
      {
        id: 'source_address_prefixs',
        name: t('IP地址组'),
      },
      {
        id: 'cloud_source_security_group_ids',
        name: t('安全组'),
      },
    ]);

    const azureSecurityGroupTarget = ref([    // 微软云目标
      {
        id: 'destination_address_prefix',
        name: t('IP地址'),
      },
      {
        id: 'destination_address_prefixes',
        name: t('IP地址组'),
      },
      {
        id: 'cloud_destination_security_group_ids',
        name: t('安全组'),
      },
    ]);

    const securityRuleId = ref('');

    const renderSourceAddressSlot = (data: any, key: string) => {
      if (data[key]) {
        return <Input class="mt20 mb10 input-select-warp"
        placeholder="请输入"
         v-model={ data[key] }>
          {{
            prefix: () => (
              <>
                {props.vendor === 'azure' ? <Select v-else class="input-prefix-select" v-model={data.sourceAddress}>
                {azureSecurityGroupSource.value.map(ele => (
                <Option value={ele.id} label={ele.name} key={ele.id} />
                ))}
                </Select> : <Select class="input-prefix-select" v-model={data.sourceAddress}>
                {securityGroupSource.value.map(ele => (
                <Option value={ele.id} label={ele.name} key={ele.id} />
                ))}
                </Select>}
              </>
            ),
          }}
        </Input>;
      }
      return <Input class="mt20 mb10 input-select-warp"
      placeholder="10.0.0.1/24、 10.0.0.1"
      v-model={ data.ipv4_cidr }>
        {{
          prefix: () => (
            <>
                {props.vendor === 'azure' ? <Select v-else class="input-prefix-select" v-model={data.sourceAddress}>
                {azureSecurityGroupSource.value.map(ele => (
                <Option value={ele.id} label={ele.name} key={ele.id} />
                ))}
                </Select> : <Select class="input-prefix-select" v-model={data.sourceAddress}>
                {securityGroupSource.value.map(ele => (
                <Option value={ele.id} label={ele.name} key={ele.id} />
                ))}
                </Select>}
              </>
          ),
        }}
      </Input>;
    };

    const renderTargetAddressSlot = (data: any, key: string) => {
      if (data[key]) {
        return <Input class="mt20 mb10 input-select-warp" v-model={ data[key] }>
          {{
            prefix: () => (
              <>
                <Select class="input-prefix-select" v-model={data.targetAddress}>
                {azureSecurityGroupTarget.value.map(ele => (
                <Option value={ele.id} label={ele.name} key={ele.id} />
                ))}
                </Select>
              </>
            ),
          }}
        </Input>;
      }
      return <Input class="mt20 mb10 input-select-warp" v-model={ data.destination_address_prefix }>
        {{
          prefix: () => (
              <>
                <Select class="input-prefix-select" v-model={data.targetAddress}>
                {azureSecurityGroupTarget.value.map(ele => (
                <Option value={ele.id} label={ele.name} key={ele.id} />
                ))}
                </Select>
              </>
          ),
        }}
      </Input>;
    };
    const columnsData = [
      { label: () => {
        return (
          <>
          <span >{t('优先级')}</span>
          <Info v-BkTooltips={{ content: '必须是 1-100的整数' }}></Info>
          </>
        );
      },
      field: 'priority',
      render: ({ data }: any) => <Input class="mt20" type='number' v-model={ data.priority }></Input>,
      },
      { label: t('策略'),
        field: 'action',
        render: ({ data }: any) => {
          return (
            <Select class="mt15 mb15" v-model={data.action}>
                {(props.vendor === 'huawei' ? HUAWEI_ACTION_STATUS : ACTION_STATUS).map((ele: any) => (
                <Option value={ele.id} label={ele.name} key={ele.id} />
                ))}
          </Select>
          );
        },
      },
      {
        label: () => {
          return (
          <>
          <span >{t('协议端口')}</span>
          <Info v-BkTooltips={{ content: '请输入0-65535之间数字或者ALL' }}></Info>
          </>
          );
        },
        field: 'port',
        render: ({ data }: any) => {
          return (
              <>
                <Input disabled={data.protocol === 'ALL'}
                placeholder="请输入0-65535之间数字、ALL"
                class="mt20 mb10 input-select-warp" v-model={ data.port }>
                {{
                  prefix: () => (
                    <Select v-model={data.protocol} clearable={false} class="input-prefix-select" onChange={handleChange}>
                    {GCP_PROTOCOL_LIST.map(ele => (
                    <Option value={ele.id} label={ele.name} key={ele.id} />
                    ))}
                    </Select>
                  ),
                }}
                </Input>
                </>
          );
        },
      },
      { label: t('类型'),
        field: 'ethertype',
        render: ({ data }: any) => {
          return (
                <>
                <Select v-model={data.ethertype} class="mt15">
                    {HUAWEI_TYPE_LIST.map(ele => (
                    <Option value={ele.id} label={ele.name} key={ele.id} />
                    ))}
                </Select>
                </>
          );
        },
      },
      {
        label: () => {
          return (
              <>
              <span >{props.activeType === 'egress' ? t('源地址') : t('目标地址')}</span>
              <Info v-BkTooltips={{ content: '必须指定 CIDR 数据块 或者 安全组 ID' }}></Info>
              </>
          );
        },
        field: 'address',
        render: ({ data }: any) => {
          return (renderSourceAddressSlot(data, data.sourceAddress));
        },
      },
      {
        label: () => {
          return (
            <>
            <span >{t('描述')}</span>
            <Info v-BkTooltips={{ content: '请输入英文描述, 最大不超过256字节' }}></Info>
            </>
          );
        },
        field: 'memo',
        render: ({ data }: any) => <Input placeholder="请输入描述" class="mt20 mb10" v-model={ data.memo }></Input>,
      },
      { label: t('操作'),
        field: 'operate',
        width: 100,
        render: ({ data, index }: any) => {
          return (
                <div class="mt15">
                <Button text theme="primary" onClick={() => {
                  hanlerCopy(data);
                }}>{t('复制')}</Button>
                <Button text theme="primary" class="ml20" onClick={() => {
                  handlerDelete(data, index);
                }}>{t('删除')}</Button>
                </div>
          );
        },
      },
    ];

    const azureColumnsData = [
      { label: t('名称'),
        field: 'name',
        render: ({ data }: any) => <Input class="mt20" v-model={ data.name }></Input>,
      },
      { label: () => {
        return (
          <>
          <span >{t('优先级')}</span>
          <Info v-BkTooltips={{ content: '跟据优先级顺序处理规则；数字越小，优先级越高。我们建议在规则之间留出间隙 - 100、200、300 等 - 这样一来便可在无需编辑现有规则的情况下添加新规，同时注意不能和当前已有规则的优先级重复' }}></Info>
          </>
        );
      },
      field: 'priority',
      render: ({ data }: any) => <Input class="mt20" type='number' placeholder="优先级" v-model={ data.priority }></Input>,
      },
      { label: t('策略'),
        field: 'access',
        render: ({ data }: any) => {
          return (
            <Select class="mt15 mb15" v-model={data.access}>
                {HUAWEI_ACTION_STATUS.map((ele: any) => (
                <Option value={ele.id} label={ele.name} key={ele.id} />
                ))}
          </Select>
          );
        },
      },
      { label: () => {
        return (
          <>
          <span>{t('源')}</span>
          <Info v-BkTooltips={{ content: '源过滤器可为“任意”、一个 IP 地址范围、一个应用程序安全组或一个默认标记。它指定此规则将允许或拒绝的特定源 IP 地址范围的传入流量' }}></Info>
          </>
        );
      },
      field: 'source',
      width: 240,
      render: ({ data }: any) => {
        return (renderSourceAddressSlot(data, data.sourceAddress));
      },
      },
      { label: () => {
        return (
          <>
          <span>{t('源端口')}</span>
          <Info v-BkTooltips={{ content: '提供单个端口(如 80)、端口范围(如 1024-65535)，或单个端口和/或端口范围的以逗号分隔的列表(如 80,1024-65535)。这指定了根据此规则将允许或拒绝哪些端口的流量。提供星号(*)可允许任何端口的流量' }}></Info>
          </>
        );
      },
      field: 'source_port_range',
      width: 180,
      render: ({ data }: any) => <Input class="mt20" placeholder="单个(80)、范围(1024-65535)" v-model={ data.source_port_range }></Input>,
      },
      { label: () => {
        return (
          <>
          <span>{t('目标')}</span>
          <Info v-BkTooltips={{ content: '提供采用 CIDR 表示法的地址范围(例如 192.168.99.0/24 或 2001:1234::/64)或提供 IP 地址(例如 192.168.99.0 或 2001:1234::)。还可提供一个由采用 IPv4 或 IPv6 的 IP 地址或地址范围构成的列表(以逗号分隔)' }}></Info>
          </>
        );
      },
      field: 'target',
      width: 240,
      render: ({ data }: any) => {
        return (renderTargetAddressSlot(data, data.targetAddress));
      },
      },
      { label: t('目标协议端口'),
        field: 'destination_port_range',
        width: 240,
        render: ({ data }: any) => {
          return (
                <>
                <Input class="mt20 mb10 input-select-warp" v-model={ data.destination_port_range }>
                  {{
                    prefix: () => (
                      <Select class="input-prefix-select" v-model={data.protocol}>
                          {AZURE_PROTOCOL_LIST.map(ele => (
                          <Option value={ele.id} label={ele.name} key={ele.id} />
                          ))}
                      </Select>
                    ),
                  }}
                </Input>
                </>
          );
        },
      },
      { label: () => {
        return (
          <>
          <span >{t('描述')}</span>
          <Info v-BkTooltips={{ content: '请输入英文描述, 最大不超过256字节' }}></Info>
          </>
        );
      },
      field: 'memo',
      render: ({ data }: any) => <Input placeholder="请输入描述" class="mt20" v-model={ data.memo }></Input>,
      },
      { label: t('操作'),
        field: 'operate',
        render: ({ data, index }: any) => {
          return (
                <div class="mt20">
                <Button text theme="primary" onClick={() => {
                  hanlerCopy(data);
                }}>{t('复制')}</Button>
                <Button text theme="primary" class="ml20" onClick={() => {
                  handlerDelete(data, index);
                }}>{t('删除')}</Button>
                </div>
          );
        },
      },
    ];
    const tableData = ref<any>([{}]);
    const columns = ref<any>(columnsData);
    const steps = [
      {
        component: () => <>
            <Table
              class="mt20"
              row-hover="auto"
              columns={columns.value}
              data={tableData.value}
            />
            {securityRuleId.value ? '' : <Button text theme="primary" class="ml20 mt20" onClick={handlerAdd}>{t('新增一条规则')}</Button>}
          </>,
      },
    ];

    watch(
      () => props.isShow,
      (v) => {
        if (!v) {
          tableData.value = [{}];
          return;
        }
        columns.value = columnsData;  // 初始化表表格列
        console.log('props.activeType', props.activeType);
        let sourceAddressData: any[] = [];
        let targetAddressData: any[] = [];
        if (props.vendor === 'tcloud' || props.vendor === 'aws') {    // 腾讯云、aws不需要优先级和类型
          columns.value = columns.value.filter((e: any) => {
            return e.field !== 'priority' && e.field !== 'ethertype';
          });
          securityGroupSource.value = [...IP_TYPE_LIST, ...[{ // 腾讯云、aws源地址特殊处理
            id: 'cloud_target_security_group_id',
            name: t('安全组'),
          }]];
          sourceAddressData = securityGroupSource.value
            .filter((e: any) => resourceStore.securityRuleDetail[e.id]);
        } else if (props.vendor === 'azure') {
          columns.value = azureColumnsData;
          sourceAddressData = azureSecurityGroupSource.value
            .filter((e: any) => resourceStore.securityRuleDetail[e.id]);
          targetAddressData = azureSecurityGroupTarget.value
            .filter((e: any) => resourceStore.securityRuleDetail[e.id]);
        }

        // @ts-ignore
        securityRuleId.value = resourceStore.securityRuleDetail?.id;
        if (securityRuleId.value) { // 如果是编辑 则需要将详细数据展示成列表数据
          tableData.value = [{ ...resourceStore.securityRuleDetail, ...{ sourceAddress: sourceAddressData[0].id },
            ...{ targetAddress: targetAddressData[0].id } }];
          columns.value = columns.value.filter((e: any) => {    // 编辑不能进行复制和删除操作
            return e.field !== 'operate';
          });
        }
      },
      {
        immediate: true,
      },
    );

    // 方法
    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleConfirm = () => {
      console.log('tableData.value', tableData.value);
      tableData.value.forEach((e: any) => {
        e[e.sourceAddress] = e.ipv4_cidr || e.ipv6_cidr || e.cloud_target_security_group_id;
        if (e.sourceAddress !== 'ipv4_cidr') {
          delete e.ipv4_cidr;
        }
        delete e.sourceAddress;
        delete e.targetAddress;
      });
      // let params: any = {
      //   id: tableData.value[0].id,
      //   // protocol: tableData.value[0].protocol,
      //   // port: tableData.value[0].port,
      //   // ipv4_cidr: tableData.value[0].ipv4_cidr,
      //   // action: tableData.value[0].action,
      //   // memo: tableData.value[0].memo,
      // };
      // if (props.vendor === 'huawei') {
      //   // params = {

      //   // };
      // }
      // if (props.vendor === 'azure') {
      //   params = { ...params, ...tableData.value[0] };
      // }
      // @ts-ignore
      if (securityRuleId.value) {  // 更新
        emit('submit', tableData.value[0]);
      } else {
        emit('submit', tableData.value);  // 新增
      }
    };

    // 新增
    const handlerAdd = () => {
      tableData.value.push({});
    };

    // 删除
    const handlerDelete = (data: any, index: any) => {
      Confirm('确定删除', '删除之后不可恢复', () => {
        tableData.value.splice(index, 1);
      });
    };

    // 复制
    const hanlerCopy = (data: any) => {
      const copyData = JSON.parse(JSON.stringify(data));
      tableData.value.push(copyData);
    };

    // 处理selectChange
    const handleChange = () => {
      tableData.value.forEach((e: any) => {
        if (e.protocol === 'ALL') {
          e.port = 'ALL';
        }
      });
    };

    return {
      steps,
      handleClose,
      handleConfirm,
    };
  },

  render() {
    return <>
        <step-dialog
        dialogWidth={this.dialogWidth}
          title={this.title}
          loading={this.loading}
          isShow={this.isShow}
          steps={this.steps}
          onConfirm={this.handleConfirm}
          onCancel={this.handleClose}
        >
        </step-dialog>
      </>;
  },
});

