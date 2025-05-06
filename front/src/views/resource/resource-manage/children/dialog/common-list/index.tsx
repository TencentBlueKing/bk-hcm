import { defineComponent, ref, watch } from 'vue';
import { Table, Input, Select, Button } from 'bkui-vue'; // TagInput
import {
  ACTION_STATUS,
  GCP_PROTOCOL_LIST,
  IP_TYPE_LIST,
  HUAWEI_ACTION_STATUS,
  HUAWEI_TYPE_LIST,
  AZURE_PROTOCOL_LIST,
} from '@/constants';
import Confirm from '@/components/confirm';
import { useI18n } from 'vue-i18n';
import StepDialog from '@/components/step-dialog/step-dialog';
import { useResourceStore } from '@/store/resource';
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
  },

  emits: ['update:isShow', 'submit'],

  setup(props, { emit }) {
    const { t } = useI18n();

    const resourceStore = useResourceStore();

    // const cloudTargetSecurityGroup = ;

    const securityGroupSource = ref([
      // 华为源
      {
        id: 'remote_ip_prefix',
        name: t('IP地址'),
      },
      {
        id: 'cloud_remote_group_id',
        name: t('安全组'),
      },
    ]);

    const azureSecurityGroupSource = ref([
      // 微软云源
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

    const azureSecurityGroupTarget = ref([
      // 微软云目标
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
      // console.log('key', key);
      // if (data.ipv4_cidr) {
      //   return <Input v-model={ data.ipv4_cidr }></Input>;
      // } if (data.ipv6_cidr) {
      //   return <Input v-model={ data.ipv6_cidr }></Input>;
      // } if (data.cloud_target_security_group_id) {
      //   return <Input v-model={ data.cloud_target_security_group_id }></Input>;
      // } if (data.remote_ip_prefix) {
      //   return <Input v-model={ data.remote_ip_prefix }></Input>;
      // } if (data.cloud_remote_group_id) {
      //   return <Input v-model={ data.cloud_remote_group_id }></Input>;
      // } if (data.source_address_prefix) {
      //   return <Input v-model={ data.source_address_prefix }></Input>;
      // } if (data.source_address_prefixs) {
      //   return <TagInput v-model={ data.source_address_prefixs }></TagInput>;
      // } if (data.cloud_source_security_group_ids) {
      //   return <Input v-model={ data.cloud_source_security_group_ids }></Input>;
      // } if (data.destination_address_prefix) {
      //   return <Input v-model={ data.destination_address_prefix }></Input>;
      // } if (data.destination_address_prefixes) {
      //   return <TagInput v-model={ data.destination_address_prefixes }></TagInput>;
      // } if (data.cloud_destination_security_group_ids) {
      //   return <Input v-model={ data.cloud_destination_security_group_ids }></Input>;
      // }
      // return <Input v-model={ data.ipv4_cidr }></Input>;
      if (data[key]) {
        return <Input v-model={data[key]}></Input>;
      }
      return <Input v-model={data.ipv4_cidr}></Input>;
    };

    const renderTargetAddressSlot = (data: any, key: string) => {
      if (data[key]) {
        return <Input v-model={data[key]}></Input>;
      }
      return <Input v-model={data.destination_address_prefix}></Input>;
    };
    const columnsData = [
      {
        label: t('优先级'),
        field: 'priority',
        render: ({ data }: any) => <Input class='mt25' type='number' v-model_number={data.priority}></Input>,
      },
      {
        label: t('策略'),
        field: 'action',
        render: ({ data }: any) => {
          return (
            <Select class='mt25' v-model={data.action}>
              {(props.vendor === 'huawei' ? HUAWEI_ACTION_STATUS : ACTION_STATUS).map((ele: any) => (
                <Option value={ele.id} label={ele.name} key={ele.id} />
              ))}
            </Select>
          );
        },
      },
      {
        label: t('协议端口'),
        field: 'port',
        render: ({ data }: any) => {
          return (
            <>
              <Select v-model={data.protocol}>
                {GCP_PROTOCOL_LIST.map((ele) => (
                  <Option value={ele.id} label={ele.name} key={ele.id} />
                ))}
              </Select>
              <Input v-model={data.port}></Input>
            </>
          );
        },
      },
      {
        label: t('类型'),
        field: 'ethertype',
        render: ({ data }: any) => {
          return (
            <>
              <Select v-model={data.ethertype}>
                {HUAWEI_TYPE_LIST.map((ele) => (
                  <Option value={ele.id} label={ele.name} key={ele.id} />
                ))}
              </Select>
            </>
          );
        },
      },
      {
        label: t('源地址'),
        field: 'id',
        render: ({ data }: any) => {
          return (
            <>
              <Select v-model={data.sourceAddress}>
                {securityGroupSource.value.map((ele) => (
                  <Option value={ele.id} label={ele.name} key={ele.id} />
                ))}
              </Select>
              {renderSourceAddressSlot(data, data.sourceAddress)}
            </>
          );
        },
      },
      { label: t('描述'), field: 'memo', render: ({ data }: any) => <Input class='mt25' v-model={data.memo}></Input> },
      {
        label: t('操作'),
        field: 'operate',
        render: ({ data, row }: any) => {
          return (
            <div class='mt20'>
              <Button
                text
                theme='primary'
                onClick={() => {
                  hanlerCopy(data);
                }}>
                {t('复制')}
              </Button>
              <Button
                text
                theme='primary'
                class='ml20'
                onClick={() => {
                  handlerDelete(data, row);
                }}>
                {t('删除')}
              </Button>
            </div>
          );
        },
      },
    ];

    const azureColumnsData = [
      { label: t('名称'), field: 'name', render: ({ data }: any) => <Input class='mt25' v-model={data.name}></Input> },
      {
        label: t('优先级'),
        field: 'priority',
        render: ({ data }: any) => <Input class='mt25' type='number' v-model_number={data.priority}></Input>,
      },
      {
        label: t('策略'),
        field: 'access',
        render: ({ data }: any) => {
          return (
            <Select class='mt25' v-model={data.access}>
              {HUAWEI_ACTION_STATUS.map((ele: any) => (
                <Option value={ele.id} label={ele.name} key={ele.id} />
              ))}
            </Select>
          );
        },
      },
      {
        label: t('源'),
        field: 'source',
        render: ({ data }: any) => {
          return (
            <>
              <Select v-model={data.sourceAddress}>
                {azureSecurityGroupSource.value.map((ele) => (
                  <Option value={ele.id} label={ele.name} key={ele.id} />
                ))}
              </Select>
              {renderSourceAddressSlot(data, data.sourceAddress)}
            </>
          );
        },
      },
      {
        label: t('源端口范围'),
        field: 'source_port_range',
        render: ({ data }: any) => <Input class='mt25' v-model={data.source_port_range}></Input>,
      },
      {
        label: t('目标'),
        field: 'target',
        render: ({ data }: any) => {
          return (
            <>
              <Select v-model={data.targetAddress}>
                {azureSecurityGroupTarget.value.map((ele) => (
                  <Option value={ele.id} label={ele.name} key={ele.id} />
                ))}
              </Select>
              {renderTargetAddressSlot(data, data.targetAddress)}
            </>
          );
        },
      },
      {
        label: t('目标协议端口'),
        field: 'destination_port_range',
        render: ({ data }: any) => {
          return (
            <>
              <Select v-model={data.protocol}>
                {AZURE_PROTOCOL_LIST.map((ele) => (
                  <Option value={ele.id} label={ele.name} key={ele.id} />
                ))}
              </Select>
              <Input v-model={data.destination_port_range}></Input>
            </>
          );
        },
      },
      { label: t('描述'), field: 'memo', render: ({ data }: any) => <Input class='mt25' v-model={data.memo}></Input> },
      {
        label: t('操作'),
        field: 'operate',
        render: ({ data, row }: any) => {
          return (
            <div class='mt20'>
              <Button
                text
                theme='primary'
                onClick={() => {
                  hanlerCopy(data);
                }}>
                {t('复制')}
              </Button>
              <Button
                text
                theme='primary'
                class='ml20'
                onClick={() => {
                  handlerDelete(data, row);
                }}>
                {t('删除')}
              </Button>
            </div>
          );
        },
      },
    ];
    const tableData = ref<any>([{}]);
    const columns = ref<any>(columnsData);
    const steps = [
      {
        component: () => (
          <>
            <Table class='mt20' row-hover='auto' columns={columns.value} data={tableData.value} />
            {securityRuleId.value ? (
              ''
            ) : (
              <Button text theme='primary' class='ml20 mt20' onClick={handlerAdd}>
                {t('新增一条规则')}
              </Button>
            )}
          </>
        ),
      },
    ];

    // watch(
    //   () => props.vendor,
    //   (v) => {
    //     if (v === 'tcloud' || v === 'aws') {
    //       nextTick(() => {
    //         columns.value = columns.value.filter((e: any) => {
    //           return e.field !== 'priority' && e.field !== 'type';
    //         });
    //       });
    //     }
    //   },
    //   {
    //     immediate: true,
    //   },
    // );

    watch(
      () => props.isShow,
      (v) => {
        if (!v) {
          tableData.value = [{}];
          return;
        }
        columns.value = columnsData; // 初始化表表格列
        if (props.vendor === 'tcloud' || props.vendor === 'aws') {
          // 腾讯云、aws不需要优先级和类型
          columns.value = columns.value.filter((e: any) => {
            return e.field !== 'priority' && e.field !== 'ethertype';
          });
          securityGroupSource.value = [
            ...IP_TYPE_LIST,
            ...[
              {
                // 腾讯云、aws源地址特殊处理
                id: 'cloud_target_security_group_id',
                name: t('安全组'),
              },
            ],
          ];
        } else if (props.vendor === 'azure') {
          columns.value = azureColumnsData;
        }

        // @ts-ignore
        securityRuleId.value = resourceStore.securityRuleDetail?.id;
        if (securityRuleId.value) {
          // 如果是编辑 则需要将详细数据展示成列表数据
          const sourceAddressData = securityGroupSource.value.filter(
            (e: any) => resourceStore.securityRuleDetail[e.id],
          );
          tableData.value = [{ ...resourceStore.securityRuleDetail, ...{ sourceAddress: sourceAddressData[0].id } }];
          columns.value = columns.value.filter((e: any) => {
            // 编辑不能进行复制和删除操作
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
      tableData.value.forEach((e: any) => {
        e[e.sourceAddress] = e.ipv4_cidr || e.ipv6_cidr || e.cloud_target_security_group_id;
        if (e.sourceAddress !== 'ipv4_cidr') {
          delete e.ipv4_cidr;
        }
        delete e.sourceAddress;
        delete e.targetAddress;
      });
      const params = {
        id: tableData.value[0].id,
        protocol: tableData.value[0].protocol,
        port: tableData.value[0].port,
        ipv4_cidr: tableData.value[0].ipv4_cidr,
        action: tableData.value[0].action,
        memo: tableData.value[0].memo,
      };
      if (props.vendor === 'huawei') {
        // params = {
        // };
      }
      // @ts-ignore
      if (securityRuleId.value) {
        // 更新
        emit('submit', params);
      } else {
        emit('submit', tableData.value); // 新增
      }
    };

    // 新增
    const handlerAdd = () => {
      tableData.value.push({});
    };

    // 删除
    const handlerDelete = (data: any, row: any) => {
      const index = row.__$table_row_index;
      Confirm('确定删除', '删除之后不可恢复', () => {
        tableData.value.splice(index, 1);
      });
    };

    // 复制
    const hanlerCopy = (data: any) => {
      const copyData = JSON.parse(JSON.stringify(data));
      tableData.value.push(copyData);
    };

    return {
      steps,
      handleClose,
      handleConfirm,
    };
  },

  render() {
    return (
      <>
        <step-dialog
          dialogWidth={this.dialogWidth}
          title={this.title}
          loading={this.loading}
          isShow={this.isShow}
          steps={this.steps}
          onConfirm={this.handleConfirm}
          onCancel={this.handleClose}></step-dialog>
      </>
    );
  },
});
