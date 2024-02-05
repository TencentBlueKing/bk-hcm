import { PropType, defineComponent, ref, watch } from 'vue';
import { Input, Select, Button, Form } from 'bkui-vue'; // TagInput
import {
  ACTION_STATUS,
  IP_TYPE_LIST,
  HUAWEI_ACTION_STATUS,
  AZURE_ACTION_STATUS,
  HUAWEI_TYPE_LIST,
  AZURE_PROTOCOL_LIST,
  SECURITY_RULES_MAP,
  TCLOUD_SOURCE_IP_TYPE_LIST,
} from '@/constants';
import Confirm from '@/components/confirm';
import { useI18n } from 'vue-i18n';
import StepDialog from '@/components/step-dialog/step-dialog';
import { useResourceStore } from '@/store/resource';
import './add-rule.scss';
import { securityRuleValidators } from './security-rule-validators';
import { VendorEnum } from '@/common/constant';
const { Option } = Select;
const { FormItem } = Form;

export type SecurityRule = {
  name: string;
  priority: number;
  ethertype: string;
  sourceAddress: string;
  source_port_range: string;
  targetAddress: string;
  protocol: string;
  destination_port_range: string;
  port: number | string;
  access: string;
  action: string;
  memo: string;
  cloud_service_id: string;
  cloud_service_group_id: string;
};

export enum IP_CIDR {
  IPV4_ALL = '0.0.0.0/0',
  IPV6_ALL = '::/0',
}

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
    relatedSecurityGroups: {
      type: Array as PropType<any>,
    },
    isEdit: {
      type: Boolean as PropType<boolean>,
    },
    templateData: {
      type: Object as PropType<Record<string, Array<any>>>,
    },
  },

  emits: ['update:isShow', 'submit'],

  setup(props, { emit }) {
    const { t } = useI18n();

    const resourceStore = useResourceStore();

    const protocolList = ref<any>(SECURITY_RULES_MAP[props.vendor]);

    const securityGroupSource = ref([
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
      // {
      //   id: 'source_address_prefixes',
      //   name: t('IP地址组'),
      // },
      // 微软云暂时禁用安全组
      // {
      //   id: 'cloud_source_security_group_ids',
      //   name: t('安全组'),
      // },
    ]);

    const azureSecurityGroupTarget = ref([
      // 微软云目标
      {
        id: 'destination_address_prefix',
        name: t('IP地址'),
      },
      // {
      //   id: 'destination_address_prefixes',
      //   name: t('IP地址组'),
      // },
      // 微软云暂时禁用安全组
      // {
      //   id: 'cloud_destination_security_group_ids',
      //   name: t('安全组'),
      // },
    ]);

    const securityRuleId = ref('');

    protocolList.value = protocolList.value.filter((e: any) => e.name !== 'ALL');
    if (props.vendor === 'aws') {
      protocolList.value.unshift({
        // @ts-ignore
        id: '-1',
        name: 'ALL',
      });
    } else if (props.vendor === 'tcloud') {
      protocolList.value.unshift({
        id: 'ALL',
        name: 'ALL',
      });
    } else if (props.vendor === 'huawei') {
      protocolList.value.unshift({
        id: 'huaweiAll',
        name: 'ALL',
      });
    }

    const translateAll = (ipType: string) => {
      return ['ipv4_cidr'].includes(ipType)
        ? IP_CIDR.IPV4_ALL
        : IP_CIDR.IPV6_ALL;
    };

    const renderSourceAddressSlot = (
      data: SecurityRule,
      key:
      | 'cloud_target_security_group_id'
      | 'ipv6_cidr'
      | 'ipv4_cidr'
      | 'source_address_prefix' // AZURE 源 IP地址
      | 'cloud_source_security_group_ids' // AZURE 源 安全组
      | 'remote_ip_prefix' // HUAWEI IP地址
      | 'cloud_remote_group_id' // HUAWEI 安全组
      | 'cloud_address_id' // 腾讯云 IP参数模板
      | 'cloud_address_group_id', // 腾讯云 IP参数模板组
    ) => {
      [
        'cloud_target_security_group_id',
        'ipv6_cidr',
        'ipv4_cidr',
        'source_address_prefix',
        'cloud_source_security_group_ids',
        'remote_ip_prefix',
        'cloud_remote_group_id',
        'cloud_address_id',
        'cloud_address_group_id',
      ].forEach(dataKey => dataKey !== key && delete data[dataKey]);

      const prefix = () => (
        <>
          {props.vendor === 'azure' ? (
            <Select
              clearable={false}
              class='input-prefix-select w120'
              v-model={data.sourceAddress}>
              {azureSecurityGroupSource.value.map(ele => (
                <Option value={ele.id} label={ele.name} key={ele.id} />
              ))}
            </Select>
          ) : (
            <Select
              clearable={false}
              class='input-prefix-large-select'
              v-model={data.sourceAddress}
              disabled={props.isEdit}>
              {securityGroupSource.value.map(ele => (
                <Option value={ele.id} label={ele.name} key={ele.id} />
              ))}
            </Select>
          )}
        </>
      );

      let list = [];
      switch (key) {
        case 'cloud_target_security_group_id':
        case 'cloud_source_security_group_ids':
        case 'cloud_remote_group_id': {
          list = props.relatedSecurityGroups;
          break;
        }
        case 'cloud_address_id': {
          list = props.templateData.ipList;
          break;
        }
        case 'cloud_address_group_id': {
          list = props.templateData.ipGroupList;
        }
      }

      return [
        'cloud_target_security_group_id',
        'cloud_source_security_group_ids',
        'cloud_remote_group_id',
        'cloud_address_id',
        'cloud_address_group_id',
      ].includes(key) ? (
        <div class={'security-group-select w120'}>
          {prefix()}
          <Select v-model={data[key]} class={'input-prefix-large-select'}>
            {list.map((securityGroup: {
              cloud_id: string | number | symbol;
              name: string;
            }) => (
                <Option
                  value={securityGroup.cloud_id}
                  label={
                    [
                      'cloud_address_id',
                      'cloud_address_group_id',
                    ].includes(key) ? `${String(securityGroup.cloud_id)} (${securityGroup.name})`
                      : securityGroup.name
                  }
                  key={securityGroup.cloud_id}
                />
            ))}
          </Select>
        </div>
        ) : (
        <Input
          class=' input-select-warp'
          placeholder='请输入'
          v-model={data[key]}
          onChange={(val: string) => {
            if (['all', 'ALL'].includes(val.trim())) {
              data[key] = translateAll(data.sourceAddress);
            }
          }}
          disabled={
            data.protocol === 'icmpv6' && data.sourceAddress === 'ipv4_cidr'
          }>
          {{
            prefix,
          }}
        </Input>
        );
    };

    const renderTargetAddressSlot = (
      data: SecurityRule,
      key:
      | 'destination_address_prefix'
      | 'cloud_destination_security_group_ids',
    ) => {
      [
        'destination_address_prefix', // AZURE 目标 IP地址
        'cloud_destination_security_group_ids', // AZURE 目标 安全组
      ].forEach(dataKey => dataKey !== key && delete data[dataKey]);
      console.log(key);
      return key !== 'cloud_destination_security_group_ids' ? (
        <Input
          class=' input-select-warp w120'
          v-model={data[key]}
          placeholder='10.0.0.1/24、10.0.0.1'
          onChange={(val: string) => {
            if (['all', 'ALL'].includes(val.trim())) {
              data[key] = translateAll(data.targetAddress);
            }
          }}>
          {{
            prefix: () => (
              <>
                <Select
                  class='input-prefix-select w100'
                  v-model={data.targetAddress}>
                  {azureSecurityGroupTarget.value.map(ele => (
                    <Option value={ele.id} label={ele.name} key={ele.id} />
                  ))}
                </Select>
              </>
            ),
          }}
        </Input>
      ) : (
        <>
          <div class='flex-row align-items-center'>
            <Select
              class='tag-input-prefix-select w100'
              v-model={data.targetAddress}>
              {azureSecurityGroupTarget.value.map(ele => (
                <Option value={ele.id} label={ele.name} key={ele.id} />
              ))}
            </Select>
            <Select v-model={data[key]} class='tag-input-select-warp w100'>
              {props.relatedSecurityGroups.map((securityGroup: {
                cloud_id: string | number | symbol;
                name: string;
              }) => (
                  <Option
                    value={securityGroup.cloud_id}
                    label={securityGroup.name}
                    key={securityGroup.cloud_id}
                  />
              ))}
            </Select>
          </div>
        </>
      );
    };

    const formInstances = [ref(null)];
    const tableData = ref<any>([{}]);
    const steps = [
      {
        component: () => (
          <>
            <div>
              {tableData.value.map((data: SecurityRule, index: number) => (
                <Form
                  ref={formInstances[index]}
                  formType='vertical'
                  model={data}
                  style={{
                    display: 'flex',
                    justifyContent: 'space-around',
                  }}
                  rules={securityRuleValidators(
                    data,
                    props.vendor as VendorEnum,
                  )}>
                  {props.vendor === 'azure' ? (
                    <FormItem
                      class='w150'
                      label={index === 0 ? t('名称') : ''}
                      required
                      property='name'>
                      <Input v-model={data.name}></Input>
                    </FormItem>
                  ) : (
                    ''
                  )}
                  {props.vendor !== 'tcloud' && props.vendor !== 'aws' ? (
                    <>
                      <FormItem
                        class='w150'
                        label={index === 0 ? t('优先级') : ''}
                        required
                        property='priority'
                        description={
                          props.vendor === 'azure'
                            ? '根据优先级顺序处理规则；数字越小，优先级越高。我们建议在规则之间留出间隙「100、200、300」等。\n这样一来便可在无需编辑现有规则的情况下添加新规，同时注意不能和当前已有规则的优先级重复。取值范围为100-4096'
                            : '优先级可选范围为1-100，默认值为1，即最高优先级。优先级数字越小，规则优先级级别越高。'
                        }>
                        <Input type='number' v-model={data.priority}></Input>
                      </FormItem>
                    </>
                  ) : (
                    ''
                  )}
                  {props.vendor === 'huawei' ? (
                    <FormItem
                      class='w150'
                      label={index === 0 ? t('类型') : ''}
                      property='ethertype'
                      required>
                      <Select v-model={data.ethertype}>
                        {HUAWEI_TYPE_LIST.map(ele => (
                          <Option
                            value={ele.id}
                            label={ele.name}
                            key={ele.id}
                          />
                        ))}
                      </Select>
                    </FormItem>
                  ) : (
                    ''
                  )}
                  {props.vendor === 'azure' ? (
                    <>
                      <FormItem
                        class='w200'
                        label={index === 0 ? t('源') : ''}
                        property='sourceAddress'
                        required
                        description='源过滤器可为“任意”、一个 IP 地址范围、一个应用程序安全组或一个默认标记。它指定此规则将允许或拒绝的特定源 IP 地址范围的传入流量'>
                        {renderSourceAddressSlot(
                          data,
                          data.sourceAddress as
                            | 'source_address_prefix'
                            | 'cloud_source_security_group_ids',
                        )}
                      </FormItem>
                      <FormItem
                        class='w200'
                        label={index === 0 ? t('源端口') : ''}
                        property='source_port_range'
                        required
                        description='提供单个端口(如 80)、端口范围(如 1024-65535)，或单个端口和/或端口范围的以逗号分隔的列表(如 80,1024-65535)。\n\r 这指定了根据此规则将允许或拒绝哪些端口的流量。提供星号(*)可允许任何端口的流量'>
                        <Input
                          placeholder='单个(80)、范围(1024-65535)'
                          v-model={data.source_port_range}></Input>
                      </FormItem>
                      <FormItem
                        class='w249'
                        label={index === 0 ? t('目标') : ''}
                        property='targetAddress'
                        required
                        description='提供采用 CIDR 表示法的地址范围(例如 192.168.99.0/24 或 2001:1234::/64)或提供 IP 地址(例如 192.168.99.0 或 2001:1234::)。\n\r 还可提供一个由采用 IPv4 或 IPv6 的 IP 地址或地址范围构成的列表(以逗号分隔)'>
                        {renderTargetAddressSlot(
                          data,
                          data.targetAddress as
                            | 'destination_address_prefix'
                            | 'cloud_destination_security_group_ids',
                        )}
                      </FormItem>
                      <FormItem
                        class='w200'
                        label={index === 0 ? t('目标协议端口') : ''}
                        property='destination_port_range'>
                        <Input
                          disabled={data?.protocol === '*'}
                          class=' input-select-warp'
                          v-model={data.destination_port_range}>
                          {{
                            prefix: () => (
                              <Select
                                class='input-prefix-select w120'
                                v-model={data.protocol}
                                onChange={(val) => {
                                  delete data.destination_port_range;
                                  if (val === '*') data.destination_port_range = '*';
                                }}>
                                {AZURE_PROTOCOL_LIST.map(ele => (
                                  <Option
                                    value={ele.id}
                                    label={ele.name}
                                    key={ele.id}
                                  />
                                ))}
                              </Select>
                            ),
                          }}
                        </Input>
                      </FormItem>
                    </>
                  ) : (
                    ''
                  )}
                  {props.vendor !== 'azure' ? (
                    <>
                      <FormItem
                        label={index === 0 ? t('协议端口') : ''}
                        property='protocalAndPort'
                        description={
                          props.vendor === 'aws'
                            ? '对于 TCP、UDP 协议，允许的端口范围。您可以指定单个端口号（例如 22）或端口号范围（例如7000-8000）'
                            : '请输入0-65535之间数字或者ALL'
                        }>
                        {
                          (() => {
                            const prefix = () => (
                              <Select
                                v-model={data.protocol}
                                clearable={false}
                                class='input-prefix-large-select'
                                onChange={handleChange}>
                                {protocolList.value.map((ele: any) => (
                                  <Option
                                    value={ele.id}
                                    label={ele.name}
                                    key={ele.id}
                                  />
                                ))}
                              </Select>
                            );

                            return ['cloud_service_id', 'cloud_service_group_id'].includes(data.protocol) ? (
                              <div class={'flex-row'}>
                                {
                                  prefix()
                                }
                                {
                                  data.protocol === 'cloud_service_id' ? (
                                    <Select v-model={data.cloud_service_id}>
                                      {
                                         props.templateData.portList.map(item => (
                                          <Option
                                            name={`${item.cloud_id} (${item.name})`}
                                            id={item.cloud_id}
                                            key={item.cloud_id}
                                          />
                                         ))
                                      }
                                    </Select>
                                  ) : (
                                    <Select v-model={data.cloud_service_group_id}>
                                      {
                                        props.templateData.portGroupList.map(item => (
                                          <Option
                                            name={`${item.cloud_id} (${item.name})`}
                                            id={item.cloud_id}
                                            key={item.cloud_id}
                                          />
                                        ))
                                      }
                                    </Select>
                                  )
                                }
                              </div>
                            ) : (<Input
                              disabled={
                                data?.protocol === 'ALL'
                                || data?.protocol === 'huaweiAll'
                                || data?.protocol === '-1'
                                || ['icmpv6', 'gre', 'icmp'].includes(data?.protocol)
                              }
                              placeholder='请输入0-65535之间数字、ALL'
                              class='input-select-warp'
                              v-model={data.port}>
                              {{
                                prefix,
                              }}
                            </Input>);
                          })()
                        }
                      </FormItem>
                      <FormItem
                        label={index === 0 ? t('源地址') : ''}
                        property='sourceAddress'
                        required
                        description='必须指定 CIDR 数据块 或者 安全组 ID'>
                        {renderSourceAddressSlot(
                          data,
                          data.sourceAddress as
                            | 'cloud_target_security_group_id'
                            | 'ipv6_cidr'
                            | 'ipv4_cidr'
                            | 'cloud_address_id'
                            | 'cloud_address_group_id',
                        )}
                      </FormItem>
                    </>
                  ) : (
                    ''
                  )}
                  {props.vendor !== 'aws' ? ( // aws没有策略
                    <FormItem
                      class='w100'
                      label={index === 0 ? t('策略') : ''}
                      property={props.vendor === 'azure' ? 'access' : 'action'}
                      required>
                      {props.vendor === 'azure' ? (
                        <Select v-model={data.access}>
                          {AZURE_ACTION_STATUS.map((ele: any) => (
                            <Option
                              value={ele.id}
                              label={ele.name}
                              key={ele.id}
                            />
                          ))}
                        </Select>
                      ) : (
                        <Select v-model={data.action}>
                          {(props.vendor === 'huawei'
                            ? HUAWEI_ACTION_STATUS
                            : ACTION_STATUS
                          ).map((ele: any) => (
                            <Option
                              value={ele.id}
                              label={ele.name}
                              key={ele.id}
                            />
                          ))}
                        </Select>
                      )}
                    </FormItem>
                  ) : (
                    ''
                  )}
                  <FormItem
                    label={index === 0 ? t('描述') : ''}
                    description='请输入英文描述, 最大不超过256个字符'>
                    <Input placeholder='请输入描述' v-model={data.memo}></Input>
                  </FormItem>
                  {!securityRuleId.value ? (
                    <FormItem label={index === 0 ? t('操作') : ''}>
                      <div>
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
                          class={'ml10'}
                          onClick={() => {
                            handlerDelete(data, index);
                          }}>
                          {t('删除')}
                        </Button>
                      </div>
                    </FormItem>
                  ) : (
                    ''
                  )}
                </Form>
              ))}
            </div>
            {securityRuleId.value ? (
              ''
            ) : (
              <Button
                text
                theme='primary'
                class='ml20 mt20'
                onClick={handlerAdd}>
                {t('新增一条规则')}
              </Button>
            )}
          </>
        ),
      },
    ];

    watch(
      () => props.isShow,
      (v) => {
        if (!v) {
          tableData.value = [{}];
          return;
        }
        let sourceAddressData: any[] = [];
        let targetAddressData: any[] = [];
        if (props.vendor === 'tcloud' || props.vendor === 'aws') {
          // 腾讯云、aws不需要优先级和类型
          securityGroupSource.value = [
            ...IP_TYPE_LIST,
            ...[
              {
                // 腾讯云、aws源地址特殊处理
                id: 'cloud_target_security_group_id',
                name: t('安全组'),
              },
            ],
            ...(props.vendor === 'tcloud' ? TCLOUD_SOURCE_IP_TYPE_LIST : []),
          ];
          sourceAddressData = securityGroupSource.value.filter((e: any) => resourceStore.securityRuleDetail[e.id]);
        } else if (props.vendor === 'azure') {
          sourceAddressData = azureSecurityGroupSource.value.filter((e: any) => resourceStore.securityRuleDetail[e.id]);
          targetAddressData = azureSecurityGroupTarget.value.filter((e: any) => resourceStore.securityRuleDetail[e.id]);
        }

        // @ts-ignore
        securityRuleId.value = resourceStore.securityRuleDetail?.id;
        if (securityRuleId.value) {
          // 如果是编辑 则需要将详细数据展示成列表数据
          tableData.value = [
            {
              ...resourceStore.securityRuleDetail,
              ...{ sourceAddress: sourceAddressData[0]?.id },
              ...{ targetAddress: targetAddressData[0]?.id },
            },
          ];
          if (props.vendor === 'aws') {
            // aws处理
            tableData.value.forEach((e: any) => {
              if (e.from_port && e.to_port && e.from_port === e.to_port) {
                e.port = e.from_port;
              }
              if (e.protocol === '-1') {
                e.port = 'ALL';
              }
            });
          } else if (props.vendor === 'azure') {
            tableData.value.forEach((e: any) => {
              if (e?.destination_port_ranges?.length) {
                e.destination_port_range = e.destination_port_ranges.join(',');
              }
            });
          }
        }
      },
      {
        immediate: true,
      },
    );

    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleConfirm = async () => {
      try {
        await Promise.all(formInstances.map(formInstance => formInstance.value.validate()));
      } catch (err) {
        console.log(err);
        return;
      }

      // eslint-disable-next-line @typescript-eslint/prefer-for-of
      for (let index = 0; index < tableData.value.length; index++) {
        const e = tableData.value[index];
        if (e.source_port_range?.includes(',')) {
          e.source_port_ranges = e.source_port_range.split(',');
          delete e.source_port_range;
        }
        if (e.destination_port_range?.includes(',')) {
          e.destination_port_ranges = e.destination_port_range.split(',');
          delete e.destination_port_range;
        }
        if (['cloud_service_id', 'cloud_service_group_id'].includes(e.protocol)) delete e.protocol;
      }

      emit('submit', tableData.value);
    };

    // 新增
    const handlerAdd = () => {
      formInstances.push(ref(null));
      tableData.value.push({});
    };

    // 删除
    const handlerDelete = (data: any, index: any) => {
      Confirm('确定删除', '删除之后不可恢复', () => {
        formInstances.splice(index, 1);
        tableData.value.splice(index, 1);
      });
    };

    // 复制
    const hanlerCopy = (data: any) => {
      formInstances.push(ref(null));
      const copyData = JSON.parse(JSON.stringify(data));
      tableData.value.push(copyData);
    };

    // 处理selectChange
    const handleChange = () => {
      tableData.value.forEach((e: any) => {
        if (
          e.protocol === 'ALL'
          || e.protocol === '-1'
          || e.protocol === '*'
          || ['icmpv6', 'gre', 'icmp'].includes(e.protocol)
        ) {
          // 依次为tcloud AWS AZURE HUAWEI
          e.port = 'ALL';
        } else if (e.protocol === '-1') {
          e.port = '-1';
        }
        if (
          e.protocol === 'huaweiAll'
          || (e.protocol === 'icmp' && props.vendor === VendorEnum.HUAWEI)
        ) {
          e.port = undefined;
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
