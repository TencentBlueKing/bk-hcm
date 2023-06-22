import { defineComponent, ref, watch } from 'vue';
import { Input, Select, Button, Form, TagInput } from 'bkui-vue'; // TagInput
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
import './add-rule.scss';
const { Option } = Select;
const { FormItem } = Form;

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
    const { t } = useI18n();

    const resourceStore = useResourceStore();

    const protocolList = ref<any>(GCP_PROTOCOL_LIST);

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
      // {
      //   id: 'source_address_prefixes',
      //   name: t('IP地址组'),
      // },
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
      // {
      //   id: 'destination_address_prefixes',
      //   name: t('IP地址组'),
      // },
      {
        id: 'cloud_destination_security_group_ids',
        name: t('安全组'),
      },
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

    const renderSourceAddressSlot = (data: any, key: string) => {
      [
        'cloud_target_security_group_id',
        'ipv6_cidr',
        'ipv4_cidr',
      ].forEach(dataKey => dataKey !== key && delete data[dataKey]);
      return (
        <Input class=' input-select-warp' placeholder='请输入' v-model={data[key]}>
          {{
            prefix: () => (
              <>
                {props.vendor === 'azure' ? (
                  <Select clearable={false} class='input-prefix-select' v-model={data.sourceAddress}>
                    {azureSecurityGroupSource.value.map(ele => (
                      <Option value={ele.id} label={ele.name} key={ele.id} />
                    ))}
                  </Select>
                ) : (
                  <Select clearable={false} class='input-prefix-select' v-model={data.sourceAddress}>
                    {securityGroupSource.value.map(ele => (
                      <Option value={ele.id} label={ele.name} key={ele.id} />
                    ))}
                  </Select>
                )}
              </>
            ),
          }}
        </Input>
      );
    };

    const renderTargetAddressSlot = (data: any, key: string) => {
      if (data[key]) {
        return data.targetAddress === 'destination_address_prefix' ? (
          <Input class=' input-select-warp' v-model={data[key]}>
            {{
              prefix: () => (
                <>
                  <Select class='input-prefix-select' v-model={data.targetAddress}>
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
            <div class='flex-row align-items-center mt15'>
              <Select class='tag-input-prefix-select' v-model={data.targetAddress}>
                {azureSecurityGroupTarget.value.map(ele => (
                  <Option value={ele.id} label={ele.name} key={ele.id} />
                ))}
              </Select>
              <TagInput class='tag-input-select-warp' allow-create list={[]} v-model={data[key]}></TagInput>
            </div>
          </>
        );
      }
      return (
        <Input
          class=' input-select-warp'
          v-model={data.destination_address_prefix}
          placeholder='10.0.0.1/24、 10.0.0.1'>
          {{
            prefix: () => (
              <>
                <Select class='input-prefix-select' v-model={data.targetAddress}>
                  {azureSecurityGroupTarget.value.map(ele => (
                    <Option value={ele.id} label={ele.name} key={ele.id} />
                  ))}
                </Select>
              </>
            ),
          }}
        </Input>
      );
    };

    const formInstances = [ref(null)];
    const tableData = ref<any>([{}]);
    const steps = [
      {
        component: () => (
          <>
            <div>
              {tableData.value.map((
                data: {
                  name: any;
                  priority: any;
                  ethertype: any;
                  sourceAddress: string;
                  source_port_range: any;
                  targetAddress: string;
                  protocol: string;
                  destination_port_range: any;
                  port: any;
                  access: any;
                  action: any;
                  memo: any;
                },
                index: number,
              ) => (
                  <Form
                    ref={formInstances[index]}
                    formType='vertical'
                    model={data}
                    style={{
                      display: 'flex',
                      justifyContent: 'space-around',
                    }}
                    rules={{
                      protocalAndPort: [
                        {
                          trigger: 'change',
                          message: '协议和端口均不能为空',
                          validator: () => {
                            return !!data.port && !!data.protocol;
                          },
                        },
                      ],
                      sourceAddress: [
                        {
                          trigger: 'blur',
                          message: '源地址类型与内容均不能为空',
                          validator: (val: string) => {
                            return !!val && !!data[val];
                          },
                        },
                      ],
                      targetAddress: [
                        {
                          trigger: 'blur',
                          message: '目标类型与内容均不能为空',
                          validator: (val: string) => {
                            return !!val && !!data[val];
                          },
                        },
                      ],
                    }}>
                    {props.vendor === 'azure' ? (
                      <FormItem label={index === 0 ? t('名称') : ''} required property='name'>
                        <Input v-model={data.name}></Input>
                      </FormItem>
                    ) : (
                      ''
                    )}
                    {props.vendor !== 'tcloud' && props.vendor !== 'aws' ? (
                      <>
                        <FormItem
                          label={index === 0 ? t('优先级') : ''}
                          required
                          property='priority'
                          description={
                            props.vendor === 'azure'
                              ? '跟据优先级顺序处理规则；数字越小，优先级越高。我们建议在规则之间留出间隙 「 100、200、300 」 等\n\r 这样一来便可在无需编辑现有规则的情况下添加新规，同时注意不能和当前已有规则的优先级重复. 取值范围为100-4096'
                              : '必须是 1-100的整数'
                          }>
                          <Input type='number' v-model={data.priority}></Input>
                        </FormItem>
                      </>
                    ) : (
                      ''
                    )}
                    {props.vendor === 'huawei' ? (
                      <FormItem label={index === 0 ? t('类型') : ''} property='ethertype' required>
                        <Select v-model={data.ethertype}>
                          {HUAWEI_TYPE_LIST.map(ele => (
                            <Option value={ele.id} label={ele.name} key={ele.id} />
                          ))}
                        </Select>
                      </FormItem>
                    ) : (
                      ''
                    )}
                    {props.vendor === 'azure' ? (
                      <>
                        <FormItem
                          label={index === 0 ? t('源') : ''}
                          property='sourceAddress'
                          required
                          description='源过滤器可为“任意”、一个 IP 地址范围、一个应用程序安全组或一个默认标记。它指定此规则将允许或拒绝的特定源 IP 地址范围的传入流量'>
                          {renderSourceAddressSlot(data, data.sourceAddress)}
                        </FormItem>
                        <FormItem
                          label={index === 0 ? t('源端口') : ''}
                          property='source_port_range'
                          required
                          description='提供单个端口(如 80)、端口范围(如 1024-65535)，或单个端口和/或端口范围的以逗号分隔的列表(如 80,1024-65535)。这指定了根据此规则将允许或拒绝哪些端口的流量。提供星号(*)可允许任何端口的流量'>
                          <Input placeholder='单个(80)、范围(1024-65535)' v-model={data.source_port_range}></Input>
                        </FormItem>
                        <FormItem
                          label={index === 0 ? t('目标') : ''}
                          property='targetAddress'
                          required
                          description='提供采用 CIDR 表示法的地址范围(例如 192.168.99.0/24 或 2001:1234::/64)或提供 IP 地址(例如 192.168.99.0 或 2001:1234::)。还可提供一个由采用 IPv4 或 IPv6 的 IP 地址或地址范围构成的列表(以逗号分隔)'>
                          {renderTargetAddressSlot(data, data.targetAddress)}
                        </FormItem>
                        <FormItem
                          label={index === 0 ? t('目标协议端口') : ''}
                          property='destination_port_range'
                          required>
                          <Input
                            disabled={data?.protocol === '*'}
                            class=' input-select-warp'
                            v-model={data.destination_port_range}>
                            {{
                              prefix: () => (
                                <Select class='input-prefix-select' v-model={data.protocol}>
                                  {AZURE_PROTOCOL_LIST.map(ele => (
                                    <Option value={ele.id} label={ele.name} key={ele.id} />
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
                            <Input
                              disabled={
                                data?.protocol === 'ALL' || data?.protocol === 'huaweiAll' || data?.protocol === '-1'
                              }
                              placeholder='请输入0-65535之间数字、ALL'
                              clearable
                              class='input-select-warp'
                              v-model={data.port}>
                              {{
                                prefix: () => (
                                  <Select
                                    v-model={data.protocol}
                                    clearable={false}
                                    class='input-prefix-select'
                                    onChange={handleChange}>
                                    {protocolList.value.map((ele: any) => (
                                      <Option value={ele.id} label={ele.name} key={ele.id} />
                                    ))}
                                  </Select>
                                ),
                              }}
                            </Input>
                          }
                        </FormItem>
                        <FormItem
                          label={index === 0 ? t('源地址') : ''}
                          property='sourceAddress'
                          required
                          description='必须指定 CIDR 数据块 或者 安全组 ID'>
                          {renderSourceAddressSlot(data, data.sourceAddress)}
                        </FormItem>
                      </>
                    ) : (
                      ''
                    )}
                    {props.vendor !== 'aws' ? ( // aws没有策略
                      <FormItem
                        label={index === 0 ? t('策略') : ''}
                        property={props.vendor === 'azure' ? 'access' : 'action'}
                        required>
                        {props.vendor === 'azure' ? (
                          <Select v-model={data.access}>
                            {HUAWEI_ACTION_STATUS.map((ele: any) => (
                              <Option value={ele.id} label={ele.name} key={ele.id} />
                            ))}
                          </Select>
                        ) : (
                          <Select v-model={data.action}>
                            {(props.vendor === 'huawei' ? HUAWEI_ACTION_STATUS : ACTION_STATUS).map((ele: any) => (
                              <Option value={ele.id} label={ele.name} key={ele.id} />
                            ))}
                          </Select>
                        )}
                      </FormItem>
                    ) : (
                      ''
                    )}
                    <FormItem
                      label={index === 0 ? t('描述') : ''}
                      property='memo'
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
              <Button text theme='primary' class='ml20 mt20' onClick={handlerAdd}>
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
        if (e.protocol === 'ALL' || e.protocol === '-1' || e.protocol === '*' || e.protocol === 'huaweiAll') {
          // 依次为tcloud AWS AZURE HUAWEI
          e.port = 'ALL';
        } else if (e.protocol === '-1') {
          e.port = -1;
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
