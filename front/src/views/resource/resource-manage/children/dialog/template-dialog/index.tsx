import { Button, Dialog, Form, Input, Select, Table } from 'bkui-vue';
import { PropType, defineComponent, reactive, ref, watch } from 'vue';
import './index.scss';
import { BkButtonGroup } from 'bkui-vue/lib/button';
const { FormItem } = Form;
const { Option } = Select;

export enum TemplateType {
  IP = 'ip',
  IP_GROUP = 'ip_group',
  PORT = 'port',
  PORT_GROUP = 'port_group',
}

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
  },
  setup(props) {
    const formData = reactive({
      name: '',
      type: TemplateType.IP,
    });
    const ipTableData = ref([
      {
        ipAddress: '',
        note: '',
      },
    ]);
    const ipGroupData = ref([]);
    const portTableData = ref([]);
    const portGroupData = ref([]);

    const ipGroupList = ref([]);
    const portGroupList = ref([]);

    const handleSubmit = () => {
      let data = {};
      switch (formData.type) {
        case TemplateType.IP: {
          data = {
            ...formData,
            ip: ipTableData.value,
          };
          break;
        }
        case TemplateType.IP_GROUP: {
          data = {
            ...formData,
            ip: ipGroupData.value,
          };
          break;
        }
        case TemplateType.PORT: {
          data = {
            ...formData,
            ip: portTableData.value,
          };
          break;
        }
        case TemplateType.PORT_GROUP: {
          data = {
            ...formData,
            ip: portGroupData.value,
          };
          break;
        }
      }
      console.log(666666, data);
    };

    watch(
      () => formData.type,
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
        <Form model={formData}>
          <FormItem label='参数模板名称' property='name' required>
            <Input placeholder='输入参数模板名称' v-model={formData.name}/>
          </FormItem>
          <FormItem label='参数模板类型' property='type' required>
            <BkButtonGroup>
              <Button
                selected={formData.type === TemplateType.IP}
                onClick={() => {
                  formData.type = TemplateType.IP;
                }}>
                IP地址
              </Button>
              <Button
                selected={formData.type === TemplateType.IP_GROUP}
                onClick={() => {
                  formData.type = TemplateType.IP_GROUP;
                }}>
                IP地址组
              </Button>
              <Button
                selected={formData.type === TemplateType.PORT}
                onClick={() => {
                  formData.type = TemplateType.PORT;
                }}>
                端口
              </Button>
              <Button
                selected={formData.type === TemplateType.PORT_GROUP}
                onClick={() => {
                  formData.type = TemplateType.PORT_GROUP;
                }}>
                端口组
              </Button>
            </BkButtonGroup>
          </FormItem>
          {[TemplateType.IP_GROUP].includes(formData.type) ? (
            <FormItem label='IP地址'>
              <Select v-model={ipGroupData.value}>
                {
                  ipGroupList.value.map(v => (
                    <Option key={v.key}  id={v.key} name={v.value}></Option>
                  ))
                }
              </Select>
            </FormItem>
          ) : null}
          {[TemplateType.PORT_GROUP].includes(formData.type) ? (
            <FormItem label='IP地址'>
              <Select v-model={portGroupData}>
              {
                  portGroupList.value.map(v => (
                    <Option key={v.key}  id={v.key} name={v.value}></Option>
                  ))
                }
              </Select>
            </FormItem>
          ) : null}
        </Form>
        {[TemplateType.IP, TemplateType.PORT].includes(formData.type) ? (
          <>
            <Table
              maxHeight={500}
              columns={[
                {
                  label: 'IP地址',
                  field: 'ipAddress',
                  render: ({ index }: { index: number }) => (
                    <Input
                      placeholder='输入IP地址'
                      v-model={ipTableData.value[index].ipAddress}
                    />
                  ),
                },
                {
                  label: '备注',
                  field: 'note',
                  render: ({ index }: { index: number }) => (
                    <Input
                      placeholder='备注信息'
                      v-model={ipTableData.value[index].note}
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
                  ipAddress: '',
                  note: '',
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
