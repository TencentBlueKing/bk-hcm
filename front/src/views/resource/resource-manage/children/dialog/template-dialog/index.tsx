/* eslint-disable no-nested-ternary */
import { Button, Dialog, Form, Input, Message, Select } from 'bkui-vue';
import { PropType, defineComponent, ref, watch } from 'vue';
import './index.scss';
import { BkButtonGroup } from 'bkui-vue/lib/button';
import { VendorEnum } from '@/common/constant';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { useAccountStore, useResourceStore } from '@/store';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { isPortAvailable, validateIpCidr } from '../security-rule/security-rule-validators';
import { TCLOUD_SECURITY_RULE_PROTOCALS } from '@/constants';
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
        id: number;
      }>,
    },
  },
  setup(props) {
    const resourceAccountStore = useResourceAccountStore();
    const resourceStore = useResourceStore();
    const accountStore = useAccountStore();
    const { whereAmI } = useWhereAmI();
    const isLoading = ref(false);
    const accountList = ref([]);
    const basicForm = ref(null);
    let formInstances = [ref(null)];
    const formData = ref({
      name: props.payload?.name || '',
      type: props.payload?.type || TemplateType.IP,
      vendor: VendorEnum.TCLOUD,
      account_id: resourceAccountStore.resourceAccount?.id || '',
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

    const handleSubmit = async () => {
      await basicForm.value.validate();
      await Promise.all(formInstances.map(formInstance => formInstance.value?.validate()));
      isLoading.value = true;
      let data = {
        ...formData.value,
      };
      switch (formData.value.type) {
        case TemplateType.IP: {
          data = {
            ...data,
            templates: ipTableData.value,
            group_templates: undefined,
          };
          break;
        }
        case TemplateType.IP_GROUP: {
          data = {
            ...data,
            group_templates: ipGroupData.value,
            templates: undefined,
          };
          break;
        }
        case TemplateType.PORT: {
          data = {
            ...data,
            templates: portTableData.value,
            group_templates: undefined,
          };
          break;
        }
        case TemplateType.PORT_GROUP: {
          data = {
            ...data,
            group_templates: portGroupData.value,
            templates: undefined,
          };
          break;
        }
      }
      const submitPromise = props.isEdit
        ? resourceStore.update('argument_templates', data, props.payload.id)
        : resourceStore.add('argument_templates/create', data);
      await submitPromise;
      Message({
        theme: 'success',
        message: props.isEdit ? '编辑成功' : '创建成功',
      });
      props.handleClose();
    };

    watch(
      () => props.isShow,
      (isShow) => {
        if (isShow) getAccountList();
      },
    );

    watch(
      () => formData.value.type,
      async (type) => {
        const params = {
          filter: {
            op: 'and',
            rules: [
              {
                field: 'vendor',
                op: 'eq',
                value: 'tcloud',
              },
              {
                field: 'type',
                op: 'eq',
                value: 'address',
              },
            ],
          },
          page: {
            start: 0,
            limit: 500,
          },
        };

        if (type === TemplateType.IP_GROUP) {
          params.filter.rules[1].value = 'address';
          const res = await resourceStore.getCommonList(
            params,
            'argument_templates/list',
          );
          ipGroupList.value = res.data.details;
        }
        if (type === TemplateType.PORT_GROUP) {
          params.filter.rules[1].value = 'service';
          const res = await resourceStore.getCommonList(
            params,
            'argument_templates/list',
          );
          portGroupList.value = res.data.details;
        }
      },
      {
        immediate: true,
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
            account_id: resourceAccountStore.resourceAccount?.id || '',
            templates: [],
            group_templates: [],
            bk_biz_id: -1,
          };
        } else {
          formData.value = {
            name: props.payload?.name || '',
            type: props.payload?.type || TemplateType.IP,
            vendor: VendorEnum.TCLOUD,
            account_id: resourceAccountStore.resourceAccount?.id || '',
            templates: props.payload?.templates || [],
            group_templates: props.payload?.group_templates || [],
            bk_biz_id: props.payload?.bk_biz_id || -1,
          };
          switch (formData.value.type) {
            case TemplateType.IP: {
              ipTableData.value = props.payload?.templates || [
                {
                  address: '',
                  description: '',
                },
              ];
              formInstances = ipTableData.value.map(_v => ref(null));
              break;
            }
            case TemplateType.IP_GROUP: {
              ipGroupData.value = props.payload?.group_templates || [];
              break;
            }
            case TemplateType.PORT: {
              portTableData.value = props.payload?.templates || [
                {
                  address: '',
                  description: '',
                },
              ];
              formInstances = ipTableData.value.map(_v => ref(null));
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

    const renderTable = (type: TemplateType) => {
      let list = [] as typeof ipTableData.value;
      if (type === TemplateType.IP) list = ipTableData.value;
      else if (type === TemplateType.PORT) list = portTableData.value;
      else return null;
      return (
        <div>
          {list.map((data, idx) => (
            <Form
              class={'template-table-item'}
              formType='vertical'
              ref={formInstances[idx]}
              model={data}
              rules={{
                description: [
                  {
                    trigger: 'blur',
                    message: '备注不能为空',
                    validator: (val: string) => !!val,
                  },
                ],
                address: [
                  {
                    trigger: 'blur',
                    message: formData.value.type === TemplateType.IP ? '请填写正确的IP地址' : '请填写合法的端口',
                    validator: (val: string) => {
                      if (formData.value.type === TemplateType.IP) return validateIpCidr(val) !== 'invalid';
                      if (formData.value.type === TemplateType.PORT) {
                        const arr = val.trim().split(':');
                        if (arr.length !== 2) return false;
                        const [protocal, port] = arr;
                        const protocols = TCLOUD_SECURITY_RULE_PROTOCALS.map(item => item.name);
                        if (!protocols.includes(protocal)) return false;
                        if (!isPortAvailable(port)) return false;
                        return true;
                      }
                    },
                  },
                ],
              }}>
              {
                formData.value.type === TemplateType.IP ? (
                  <FormItem
                    property='address'
                    label={'IP地址'}>
                    <Input
                      placeholder={'输入IP地址'}
                      v-model={list[idx].address}
                    />
                  </FormItem>
                ) : (
                  <FormItem
                    property='address'
                    label={'协议端口'}>
                    <Input
                      placeholder={'输入端口'}
                      v-model={list[idx].address}
                    />
                  </FormItem>
                )
              }
              <FormItem
                label={`${idx > 0 ? '' : '备注'}`}
                property='description'>
                <Input placeholder='备注信息' v-model={list[idx].description} />
              </FormItem>
              <FormItem label={`${idx > 0 ? '' : '操作'}`}>
                <Button
                  text
                  class={'ml6'}
                  theme='primary'
                  onClick={() => {
                    list.splice(idx, 1);
                    formInstances.splice(idx, 1);
                  }}>
                  删除
                </Button>
              </FormItem>
            </Form>
          ))}
          <Button
            text
            theme='primary'
            class={'mt20'}
            onClick={() => {
              list.push({
                address: '',
                description: '',
              });
              formInstances.push(ref(null));
            }}>
            新增一行
          </Button>
        </div>
      );
    };

    const getAccountList = async () => {
      const isResource = whereAmI.value === Senarios.resource;
      const payload = isResource
        ? {
          page: {
            count: false,
            limit: 100,
            start: 0,
          },
          filter: {
            op: 'and',
            rules: [
              {
                field: 'vendor',
                op: 'eq',
                value: VendorEnum.TCLOUD,
              },
            ],
          },
        }
        : {
          params: {
            account_type: 'resource',
          },
        };
      const res = await accountStore.getAccountList(payload, accountStore.bizs);
      if (resourceAccountStore.resourceAccount?.id) {
        accountList.value = res.data?.details.filter(({ id }) => id === resourceAccountStore.resourceAccount?.id);
        return;
      }
      accountList.value = isResource ? res?.data?.details : res?.data;
    };

    return () => (
      <Dialog
        isShow={props.isShow}
        onClosed={() => props.handleClose()}
        onConfirm={() => {
          handleSubmit();
        }}
        title='新建参数模板'
        maxHeight={'720px'}
        width={1000}>
        <Form model={formData.value} ref={basicForm}>
          <FormItem label='云账号' property='account_id' required>
            <Select v-model={formData.value.account_id}>
              {accountList.value.map(item => (
                <Option key={item.id} id={item.id} name={item.name}></Option>
              ))}
            </Select>
          </FormItem>
          <FormItem label='参数模板名称' property='name' required>
            <Input
              placeholder='输入参数模板名称'
              v-model={formData.value.name}
            />
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
              <Select v-model={ipGroupData.value} multiple>
                {ipGroupList.value.map(v => (
                  <Option
                    key={v.cloud_id}
                    id={v.cloud_id}
                    name={v.cloud_id}></Option>
                ))}
              </Select>
            </FormItem>
          ) : null}
          {[TemplateType.PORT_GROUP].includes(formData.value.type) ? (
            <FormItem label='IP地址'>
              <Select v-model={portGroupData} multiple>
                {portGroupList.value.map(v => (
                  <Option
                    key={v.cloud_id}
                    id={v.cloud_id}
                    name={v.cloud_id}></Option>
                ))}
              </Select>
            </FormItem>
          ) : null}
        </Form>
        {[TemplateType.IP, TemplateType.PORT].includes(formData.value.type) ? (
          <>{renderTable(formData.value.type)}</>
        ) : null}
      </Dialog>
    );
  },
});
