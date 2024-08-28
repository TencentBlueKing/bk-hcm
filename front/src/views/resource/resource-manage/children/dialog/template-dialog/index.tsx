/* eslint-disable no-nested-ternary */
import { Button, Dialog, Form, Input, Message, Select } from 'bkui-vue';
import { PropType, defineComponent, ref, watch, nextTick } from 'vue';
import { analysisIP, analysisPort, isIpsValid, isPortValid } from '@/utils';
import { BkButtonGroup } from 'bkui-vue/lib/button';
import { VendorEnum } from '@/common/constant';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { useAccountStore, useResourceStore } from '@/store';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { pluginHandlerDialog } from '@pluginHandler/resource-template-dialog';
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
    handleSuccess: {
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
        id: string;
        account_id: string;
      }>,
    },
  },
  setup(props) {
    const resourceAccountStore = useResourceAccountStore();
    const resourceStore = useResourceStore();
    const accountStore = useAccountStore();
    const { whereAmI } = useWhereAmI();
    const { vendorArr } = pluginHandlerDialog;
    const isLoading = ref(false);
    const accountList = ref([]);
    const basicForm = ref(null);
    const isGroupLoading = ref(false);
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
    const clearValidate = () => {
      basicForm.value.clearValidate();
      formInstance.value?.clearValidate();
    };
    const handleSubmit = async () => {
      await basicForm.value.validate();
      await formInstance.value?.validate();
      isLoading.value = true;
      let data = {
        ...formData.value,
      };
      switch (formData.value.type) {
        case TemplateType.IP: {
          ipTableData.value = analysisIP(formList.value.ipsList);
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
          portTableData.value = analysisPort(formList.value.portList);
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
      try {
        const submitPromise = props.isEdit
          ? resourceStore.update('argument_templates', data, props.payload.id)
          : resourceStore.add('argument_templates/create', data);
        await submitPromise;
      } finally {
        isLoading.value = false;
      }

      Message({
        theme: 'success',
        message: props.isEdit ? '编辑成功' : '创建成功',
      });
      props.handleSuccess();
      clearFormData();
    };

    watch(
      () => props.isShow,
      async (isShow) => {
        if (isShow) getAccountList();
        await nextTick();
        clearValidate();
      },
    );
    const editAssignment = () => {
      switch (formData.value.type) {
        case TemplateType.IP: {
          let result = '';
          ipTableData.value = props.payload?.templates || [
            {
              address: '',
              description: '',
            },
          ];
          ipTableData.value.forEach((item) => {
            result += `${item.address} ${item.description}\n`;
          });

          // 删除最后一个多余的换行符（可选）
          result = result.trim();
          formList.value.ipsList = result;
          break;
        }
        case TemplateType.IP_GROUP: {
          ipGroupData.value = props.payload?.group_templates || [];
          break;
        }
        case TemplateType.PORT: {
          let result = '';
          portTableData.value = props.payload?.templates || [
            {
              address: '',
              description: '',
            },
          ];
          portTableData.value.forEach((item) => {
            result += `${item.address} ${item.description}\n`;
          });
          // 删除最后一个多余的换行符（可选）
          result = result.trim();
          formList.value.portList = result;
          break;
        }
        case TemplateType.PORT_GROUP: {
          portGroupData.value = props.payload?.group_templates || [];
          break;
        }
      }
    };
    watch(
      () => [formData.value.type, formData.value.account_id],
      async ([type, accountID]) => {
        if (!accountID) return;
        isGroupLoading.value = true;
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
              {
                field: 'account_id',
                op: 'eq',
                value: formData.value.account_id,
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
          const res = await resourceStore.getCommonList(params, 'argument_templates/list');
          ipGroupData.value = [];
          ipGroupList.value = res.data.details;
        }
        if (type === TemplateType.PORT_GROUP) {
          params.filter.rules[1].value = 'service';
          const res = await resourceStore.getCommonList(params, 'argument_templates/list');

          portGroupData.value = [];
          portGroupList.value = res.data.details;
        }
        if (props.isEdit) {
          editAssignment();
        }
        isGroupLoading.value = false;
      },
      {
        immediate: true,
      },
    );

    watch(
      () => props.payload,
      async () => {
        formData.value = {
          name: props.payload?.name || '',
          type: props.payload?.type || TemplateType.IP,
          vendor: VendorEnum.TCLOUD,
          account_id: props.payload?.account_id || resourceAccountStore.resourceAccount?.id || '',
          templates: props.payload?.templates || [],
          group_templates: props.payload?.group_templates || [],
          bk_biz_id: props.payload?.bk_biz_id || -1,
        };
        editAssignment();
      },
      {
        deep: true,
      },
    );
    const modifyVendor = (account_id: string) => {
      const idx = accountList.value.findIndex(({ id }) => id === account_id);
      formData.value.vendor = accountList.value[idx]?.vendor;
    };
    watch(
      () => accountList.value,
      () => {
        modifyVendor(formData.value.account_id);
      },
      {
        deep: true,
      },
    );
    const formInstance = ref();
    const clearFormData = () => {
      formData.value = {
        name: '',
        type: TemplateType.IP,
        vendor: VendorEnum.TCLOUD,
        account_id: resourceAccountStore.resourceAccount?.id || '',
        templates: [],
        group_templates: [],
        bk_biz_id: -1,
      };
      ipGroupData.value = [];
      ipTableData.value = [
        {
          address: '',
          description: '',
        },
      ];
      portGroupData.value = [];
      portTableData.value = [
        {
          address: '',
          description: '',
        },
      ];
      formList.value = {
        ipsList: '',
        portList: '',
      };
      clearValidate();
    };
    const autoSizeConf = {
      minRows: 9,
      maxRows: 9,
    };
    const formList = ref({
      ipsList: '',
      portList: '',
    });
    const ipsMessage = ref('IP地址不能为空');
    const portMessage = ref('协议端口不能为空');
    watch(
      () => formList.value.ipsList,
      (val) => {
        ipsMessage.value = val === '' ? 'IP地址不能为空' : 'IP地址不合法';
      },
      {
        deep: true,
      },
    );
    watch(
      () => formList.value.portList,
      (val) => {
        portMessage.value = val === '' ? '协议端口不能为空' : '协议端口不合法';
      },
      {
        deep: true,
      },
    );
    const renderTable = (type: TemplateType) => {
      if (![TemplateType.IP, TemplateType.PORT].includes(type)) return null;
      return (
        <div>
          <Form
            formType='vertical'
            ref={formInstance}
            model={formList.value}
            rules={{
              ipsList: [
                {
                  required: true,
                  trigger: 'blur',
                  message: ipsMessage.value,
                  validator: (val: string) => {
                    if (val === '') {
                      return false;
                    }
                    const isValid = isIpsValid(val);
                    return isValid;
                  },
                },
              ],
              portList: [
                {
                  required: true,
                  trigger: 'blur',
                  message: portMessage.value,
                  validator: (val: string) => {
                    if (val === '') {
                      return false;
                    }
                    const isValid = isPortValid(val);
                    return isValid;
                  },
                },
              ],
            }}>
            {formData.value.type === TemplateType.IP ? (
              <FormItem property='ipsList' label={'IP地址'} required>
                <Input
                  placeholder={'每行一个IP,使用空格区隔IP与备注信息,换行后可输入多个IP'}
                  autosize={autoSizeConf}
                  type='textarea'
                  v-model={formList.value.ipsList}
                />
              </FormItem>
            ) : (
              <FormItem property='portList' label={'协议端口'} required>
                <Input
                  placeholder={`协议端口可添加多个协议端口,换行分隔,案例如下:\n【单个端口】TCP:80 备注说明\n【多个离散端口】TCP:80,433 备注说明\n【连续端口】TCP:3306-20000 备注说明\n【所有端口】TCP:ALL 备注说明\n【ICMP协议】 ICMP 备注说明\n【GRE协议】 GRE 备注说明 `}
                  autosize={autoSizeConf}
                  type='textarea'
                  v-model={formList.value.portList}
                />
              </FormItem>
            )}
          </Form>
        </div>
      );
    };

    const getAccountList = async () => {
      const isResource = whereAmI.value === Senarios.resource;
      const isBusiness = whereAmI.value === Senarios.business;
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
                  op: 'in',
                  value: vendorArr,
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
      accountList.value = isResource ? res?.data?.details : res?.data;
      if (resourceAccountStore.resourceAccount?.id) {
        accountList.value = res.data?.details.filter(({ id }) => id === resourceAccountStore.resourceAccount?.id);
        return;
      }
      if (isBusiness) {
        accountList.value = accountList.value.filter(({ vendor }) => vendorArr.includes(vendor));
      }
    };
    return () => (
      <Dialog
        isShow={props.isShow}
        title={props.isEdit ? '编辑参数模板' : '新建参数模板'}
        width={640}
        onClosed={() => {
          props.handleClose();
          clearFormData();
        }}>
        {{
          default: () => (
            <>
              <Form model={formData.value} ref={basicForm} formType='vertical'>
                <FormItem label='云账号' property='account_id' required>
                  <Select v-model={formData.value.account_id} disabled={props.isEdit} onChange={modifyVendor}>
                    {accountList.value.map((item) => (
                      <Option key={item.id} id={item.id} name={item.name}></Option>
                    ))}
                  </Select>
                </FormItem>
                <FormItem label='参数模板名称' property='name' required>
                  <Input placeholder='输入参数模板名称' v-model={formData.value.name} />
                </FormItem>
                <FormItem label='参数模板类型' property='type' required>
                  <BkButtonGroup style={'width:100%'}>
                    <Button
                      style={'width:25%'}
                      selected={formData.value.type === TemplateType.IP}
                      disabled={props.isEdit && !(props.payload.type === TemplateType.IP)}
                      onClick={() => {
                        formData.value.type = TemplateType.IP;
                      }}>
                      IP地址
                    </Button>
                    <Button
                      style={'width:25%'}
                      selected={formData.value.type === TemplateType.IP_GROUP}
                      disabled={props.isEdit && !(props.payload.type === TemplateType.IP_GROUP)}
                      onClick={() => {
                        formData.value.type = TemplateType.IP_GROUP;
                      }}>
                      IP地址组
                    </Button>
                    <Button
                      style={'width:25%'}
                      selected={formData.value.type === TemplateType.PORT}
                      disabled={props.isEdit && !(props.payload.type === TemplateType.PORT)}
                      onClick={() => {
                        formData.value.type = TemplateType.PORT;
                      }}>
                      协议端口
                    </Button>
                    <Button
                      style={'width:25%'}
                      selected={formData.value.type === TemplateType.PORT_GROUP}
                      disabled={props.isEdit && !(props.payload.type === TemplateType.PORT_GROUP)}
                      onClick={() => {
                        formData.value.type = TemplateType.PORT_GROUP;
                      }}>
                      协议端口组
                    </Button>
                  </BkButtonGroup>
                </FormItem>
                {[TemplateType.IP_GROUP].includes(formData.value.type) ? (
                  <FormItem label='IP地址'>
                    <Select v-model={ipGroupData.value} multiple multipleMode='tag'>
                      {ipGroupList.value.map((v) => (
                        <Option key={v.cloud_id} id={v.cloud_id} name={`${v.cloud_id} (${v.name})`}></Option>
                      ))}
                    </Select>
                  </FormItem>
                ) : null}
                {[TemplateType.PORT_GROUP].includes(formData.value.type) ? (
                  <FormItem label='协议端口'>
                    <Select v-model={portGroupData.value} multiple multipleMode='tag'>
                      {portGroupList.value.map((v) => (
                        <Option key={v.cloud_id} id={v.cloud_id} name={`${v.cloud_id} (${v.name})`}></Option>
                      ))}
                    </Select>
                  </FormItem>
                ) : null}
              </Form>
              {[TemplateType.IP, TemplateType.PORT].includes(formData.value.type) ? (
                <>{renderTable(formData.value.type)}</>
              ) : null}
            </>
          ),
          footer: () => (
            <>
              <Button theme='primary' loading={isLoading.value} onClick={handleSubmit}>
                确定
              </Button>
              <Button
                class='ml8'
                onClick={() => {
                  props.handleClose();
                  clearFormData();
                }}>
                取消
              </Button>
            </>
          ),
        }}
      </Dialog>
    );
  },
});
