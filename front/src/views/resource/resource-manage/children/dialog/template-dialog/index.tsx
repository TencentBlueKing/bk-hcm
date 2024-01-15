import { Button, Dialog, Form, Input, Select, Table } from 'bkui-vue';
import { PropType, defineComponent, ref, watch } from 'vue';
import './index.scss';
import { BkButtonGroup } from 'bkui-vue/lib/button';
import { VendorEnum } from '@/common/constant';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
const { FormItem } = Form;
const { Option } = Select;

export enum TemplateType {
  IP = 'address',
  IP_GROUP = 'address_group',
  PORT = 'service',
  PORT_GROUP = 'service_group',
}

export const TemplateTypeMap = {
  [TemplateType.IP]: 'IP地址',
  [TemplateType.IP_GROUP]: 'IP地址组',
  [TemplateType.PORT]: '协议端口',
  [TemplateType.PORT_GROUP]: '协议端口组',
};

export default defineComponent({
  props: {
    isShow: {
      required: true,
      type: Boolean,
    },
    handleClose: {
      required: true,
      type: Function as PropType<() => void>,
    },
    isEdit: {
      required: true,
      type: Boolean,
    },
    payload: {
      required: false,
      type: Object as PropType<{
        name: string;
        type: TemplateType;
        templates?: Array<{
          address: string;
          description: string;
        }>;
        group_templates?: Array<string>;
        bk_biz_id: number;
      }>,
    },
  },
  setup(props) {
    const resourceAccountStore = useResourceAccountStore();
    const formData = ref({
      name: props.payload?.name || '',
      type: props.payload?.type || TemplateType.IP,
      vendor: VendorEnum.TCLOUD,
      account_id: resourceAccountStore.resourceAccount.id || '',
      templates: props.payload?.templates || [],
      group_templates: props.payload?.group_templates || [],
      bk_biz_id: props.payload?.bk_biz_id || -1,
    });
    const ipTableData = ref([
      {
        address: '',
        description: '',
      },
    ]);
    const ipGroupData = ref([]);
    const portTableData = ref([
      {
        address: '',
        description: '',
      },
    ]);
    const portGroupData = ref([]);

    const ipGroupList = ref([]);
    const portGroupList = ref([]);

    const handleSubmit = () => {
      let data = {};
      switch (formData.value.type) {
        case TemplateType.IP: {
          data = {
            ...formData.value,
            templates: ipTableData.value,
            group_templates: undefined,
          };
          break;
        }
        case TemplateType.IP_GROUP: {
          data = {
            ...formData.value,
            group_templates: ipGroupData.value,
            templates: undefined,
          };
          break;
        }
        case TemplateType.PORT: {
          data = {
            ...formData.value,
            templates: portTableData.value,
            group_templates: undefined,
          };
          break;
        }
        case TemplateType.PORT_GROUP: {
          data = {
            ...formData.value,
            group_templates: portGroupData.value,
            templates: undefined,
          };
          break;
        }
      }
      console.log(666666, data);
    };

    watch(
      () => formData.value.type,
      (type) => {
        if (type === TemplateType.IP_GROUP) {
          ipGroupList.value = [
            {
              key: 'qweqwe',
              value: '123',
            },
          ];
        }
        if (type === TemplateType.PORT_GROUP) {
          portGroupList.value = [
            {
              key: 'qweqwe',
              value: '456',
            },
          ];
        }
      },
    );

    watch(
      () => props.isEdit,
      (isEdit) => {
        if (!isEdit) {
          formData.value = {
            name: '',
            type: TemplateType.IP,
            vendor: VendorEnum.TCLOUD,
            account_id: resourceAccountStore.resourceAccount.id || '',
            templates: [],
            group_templates: [],
            bk_biz_id: -1,
          };
        } else {
          formData.value = {
            name: props.payload?.name || '',
            type: props.payload?.type || TemplateType.IP,
            vendor: VendorEnum.TCLOUD,
            account_id: resourceAccountStore.resourceAccount.id || '',
            templates: props.payload?.templates || [],
            group_templates: props.payload?.group_templates || [],
            bk_biz_id: props.payload?.bk_biz_id || -1,
          };
          switch (formData.value.type) {
            case TemplateType.IP: {
              ipTableData.value = props.payload?.templates || [{
                address: '',
                description: '',
              }];
              break;
            }
            case TemplateType.IP_GROUP: {
              ipGroupData.value = props.payload?.group_templates || [];
              break;
            }
            case TemplateType.PORT: {
              portTableData.value = props.payload?.templates || [{
                address: '',
                description: '',
              }];
              break;
            }
            case TemplateType.PORT_GROUP: {
              portGroupData.value = props.payload?.group_templates || [];
              break;
            }
          }
        }
      },
    );

    return () => (
      <Dialog
        isShow={props.isShow}
        onClosed={() => props.handleClose()}
        onConfirm={() => {
          handleSubmit();
          props.handleClose();
        }}
        title='新建参数模板'
        maxHeight={'720px'}
        width={1000}>
        <Form model={formData.value}>
          <FormItem label='参数模板名称' property='name' required>
            <Input placeholder='输入参数模板名称' v-model={formData.value.name} />
          </FormItem>
          <FormItem label='参数模板类型' property='type' required>
            <BkButtonGroup>
              <Button
                selected={formData.value.type === TemplateType.IP}
                onClick={() => {
                  formData.value.type = TemplateType.IP;
                }}>
                IP地址
              </Button>
              <Button
                selected={formData.value.type === TemplateType.IP_GROUP}
                onClick={() => {
                  formData.value.type = TemplateType.IP_GROUP;
                }}>
                IP地址组
              </Button>
              <Button
                selected={formData.value.type === TemplateType.PORT}
                onClick={() => {
                  formData.value.type = TemplateType.PORT;
                }}>
                端口
              </Button>
              <Button
                selected={formData.value.type === TemplateType.PORT_GROUP}
                onClick={() => {
                  formData.value.type = TemplateType.PORT_GROUP;
                }}>
                端口组
              </Button>
            </BkButtonGroup>
          </FormItem>
          {[TemplateType.IP_GROUP].includes(formData.value.type) ? (
            <FormItem label='IP地址'>
              <Select v-model={ipGroupData.value}>
                {ipGroupList.value.map(v => (
                  <Option key={v.key} id={v.key} name={v.value}></Option>
                ))}
              </Select>
            </FormItem>
          ) : null}
          {[TemplateType.PORT_GROUP].includes(formData.value.type) ? (
            <FormItem label='IP地址'>
              <Select v-model={portGroupData}>
                {portGroupList.value.map(v => (
                  <Option key={v.key} id={v.key} name={v.value}></Option>
                ))}
              </Select>
            </FormItem>
          ) : null}
        </Form>
        {[TemplateType.IP, TemplateType.PORT].includes(formData.value.type) ? (
          <>
            <Table
              maxHeight={500}
              columns={[
                {
                  label: formData.value.type === TemplateType.IP ? 'IP地址' : '协议端口',
                  field: 'address',
                  render: ({ index }: { index: number }) => (
                    <div>
                      {formData.value.type === TemplateType.IP ? (
                        <Input
                          placeholder='输入IP地址'
                          v-model={ipTableData.value[index].address}
                        />
                      ) : (
                        <Input
                          placeholder='输入协议端口'
                          v-model={portTableData.value[index].address}
                        />
                      )}
                    </div>
                  ),
                },
                {
                  label: '备注',
                  field: 'description',
                  render: ({ index }: { index: number }) => (
                    <Input
                      placeholder='备注信息'
                      v-model={ipTableData.value[index].description}
                    />
                  ),
                },
                {
                  label: '操作',
                  field: 'actions',
                  render: ({ index }: { index: number }) => (
                    <div>
                      <Button
                        text
                        class={'ml6'}
                        theme='primary'
                        onClick={() => {
                          ipTableData.value.splice(index, 1);
                          console.log(index, ipTableData.value);
                        }}>
                        删除
                      </Button>
                    </div>
                  ),
                },
              ]}
              data={ipTableData.value}
            />
            <Button
              text
              theme='primary'
              class={'mt20'}
              onClick={() => {
                ipTableData.value.push({
                  address: '',
                  description: '',
                });
              }}>
              新增一行
            </Button>
          </>
        ) : null}
      </Dialog>
    );
  },
});
