import { defineComponent, PropType, ref, watch } from 'vue';
import './index.scss';
import { SelectColumn, InputColumn, OperationColumn } from '@blueking/ediatable';
import { SecurityVendorType, useProtocols } from '../useProtocolList';
import useFormModel from '@/hooks/useFormModel';
import SourceAddress, { TcloudSourceAddressType } from './SourceAddress';
import { Ext, IHead, SecurityRuleType } from '../useVendorHanlder';
import { cleanObject, isPortAvailable, random } from '../util';
export interface TcloudSecurityGroupRule {
  protocol: string; // 协议, 取值: TCP, UDP, ICMP, ICMPv6, ALL
  port: string; // 端口(all, 离散port, range)
  cloud_service_id?: string; // 协议端口云ID
  cloud_service_group_id?: string; // 协议端口组云ID
  ipv4_cidr?: string; // IPv4网段
  ipv6_cidr?: string; // IPv6网段
  cloud_address_id?: string; // IP参数模版云ID
  cloud_address_group_id?: string; // IP参数模版集合云ID
  cloud_target_security_group_id?: string; // 下一跳安全组实例云ID
  action: string; // ACCEPT 或 DROP
  memo?: string; // 备注
  key: string;
}

export const TcloudRecord = (): Ext<TcloudSecurityGroupRule> => ({
  protocol: '',
  port: '',
  cloud_service_id: '',
  cloud_service_group_id: '',
  ipv4_cidr: '',
  ipv6_cidr: '',
  cloud_address_id: '',
  cloud_address_group_id: '',
  cloud_target_security_group_id: '',
  action: '',
  memo: '',
  key: random(),
  sourceAddress: TcloudSourceAddressType.IPV4,
});

export const tcloudTitles: (type: SecurityRuleType) => IHead[] = (type) => [
  {
    width: 450,
    minWidth: 120,
    title: type === 'ingress' ? '源地址类型' : '目标地址类型',
  },
  {
    width: 450,
    minWidth: 120,
    title: type === 'ingress' ? '源地址' : '目标地址',
  },
  {
    width: 450,
    minWidth: 120,
    title: '协议',
  },
  {
    width: 450,
    minWidth: 120,
    title: '端口',
    memo: '请输入1-65535之间数字或者ALL',
  },
  {
    width: 450,
    minWidth: 120,
    title: '策略',
  },
  {
    width: 450,
    minWidth: 120,
    title: '备注',
    memo: '请输入英文描述, 最大不超过100个字符',
    required: false,
  },
  {
    width: 450,
    minWidth: 120,
    title: '操作',
    required: false,
  },
];

export const tcloudSourceAddressTypes = [
  { value: 'ipv4_cidr', label: 'IPv4' },
  { value: 'ipv6_cidr', label: 'IPv6' },
  { value: 'cloud_target_security_group_id', label: '安全组' },
  { value: 'cloud_address_id', label: '参数模板-IP地址' },
  { value: 'cloud_address_group_id', label: '参数模板-IP地址组' },
];

export const tcloudStrategys = [
  { value: 'ACCEPT', label: '允许' },
  { value: 'DROP', label: '拒绝' },
];

export const TcloudSourceTypeArr = [
  TcloudSourceAddressType.IPV4,
  TcloudSourceAddressType.IPV6,
  TcloudSourceAddressType.SECURITY_GROUP,
  TcloudSourceAddressType.TEMPLATE_IP,
  TcloudSourceAddressType.TEMPLATE_IP_GROUP,
];

export enum TcloudTemplatePort {
  Port = 'cloud_service_id', // 参数模板-端口
  Port_Group = 'cloud_service_group_id', // 参数模板-端口组
}

export const TcloudTemplatePortArr = [TcloudTemplatePort.Port, TcloudTemplatePort.Port_Group];

export const TcloudRenderRow = defineComponent({
  props: {
    vendor: String as PropType<SecurityVendorType>,
    templateData: Object as PropType<{
      ipList: Array<string>;
      ipGroupList: Array<string>;
      portList: Array<{
        cloud_id: string;
        name: string;
      }>;
      portGroupList: Array<{
        cloud_id: string;
        name: string;
      }>;
    }>,
    relatedSecurityGroups: Array as PropType<Array<Object>>,
    removeable: Boolean as PropType<boolean>,
    value: Object as PropType<TcloudSecurityGroupRule>,
    isEdit: Boolean as PropType<boolean>,
  },
  emits: ['add', 'remove', 'copy', 'change'],
  setup(props, { expose, emit }) {
    const { protocols } = useProtocols(props.vendor);
    const { formModel } = useFormModel(props.value);

    const protocolRef = ref();
    const portRef = ref();
    const sourceAddressTypeRef = ref();
    const sourceAddressValRef = ref();
    const actionRef = ref();

    const handleAdd = () => {
      emit('add');
    };

    const handleRemove = () => {
      emit('remove');
    };

    const handleCopy = () => {
      emit('copy', formModel);
    };

    const handleChange = (val: TcloudSecurityGroupRule) => {
      emit('change', val);
    };

    watch(
      () => formModel.protocol,
      () => {
        if (['ALL', 'icmp', 'gre', 'icmpv6'].includes(formModel.protocol)) {
          formModel.port = 'ALL';
        } else {
          formModel.port = '';
          formModel.cloud_service_id = '';
          formModel.cloud_service_group_id = '';
        }
      },
    );

    watch(
      () => formModel,
      (val) => {
        handleChange(val);
      },
      {
        deep: true,
      },
    );

    expose({
      getValue: async () => {
        await Promise.all([
          protocolRef.value.getValue(),
          portRef.value.getValue(),
          sourceAddressTypeRef.value.getValue(),
          sourceAddressValRef.value.getValue(),
          actionRef.value.getValue(),
        ]);
        return cleanObject(formModel);
      },
    });

    return () => (
      <>
        <tr>
          <td>
            <SelectColumn
              list={tcloudSourceAddressTypes}
              v-model={formModel.sourceAddress}
              ref={sourceAddressTypeRef}
              rules={[
                {
                  validator: (value: string) => Boolean(value),
                  message: '源地址类型不能为空',
                },
                {
                  validator: (value: string) =>
                    (formModel.protocol === 'icmpv6' && value !== TcloudSourceAddressType.IPV4) ||
                    formModel.protocol !== 'icmpv6',
                  message: 'ICMPV6 不支持 IPV4',
                },
              ]}
            />
          </td>
          <td>
            <SourceAddress
              v-model={formModel[formModel.sourceAddress]}
              {...props}
              sourceAddressType={formModel.sourceAddress as TcloudSourceAddressType}
              ref={sourceAddressValRef}
            />
          </td>
          <td>
            <SelectColumn
              list={protocols.value}
              v-model={formModel.protocol}
              ref={protocolRef}
              rules={[
                {
                  validator: (value: string) => Boolean(value),
                  message: '协议不能为空',
                },
              ]}
            />
          </td>
          <td>
            {
              // 参数模板-端口
              formModel.protocol === TcloudTemplatePort.Port && (
                <SelectColumn
                  ref={portRef}
                  v-model={formModel.cloud_service_id}
                  list={props.templateData.portList.map(({ name, cloud_id }) => ({
                    label: name,
                    value: cloud_id,
                    key: cloud_id,
                  }))}
                  rules={[
                    {
                      validator: (value: string) => Boolean(value),
                      message: '端口不能为空',
                    },
                  ]}
                />
              )
            }
            {
              // 参数模板-端口组
              formModel.protocol === TcloudTemplatePort.Port_Group && (
                <SelectColumn
                  ref={portRef}
                  v-model={formModel.cloud_service_group_id}
                  list={props.templateData.portGroupList.map(({ name, cloud_id }) => ({
                    label: name,
                    value: cloud_id,
                    key: cloud_id,
                  }))}
                  rules={[
                    {
                      validator: (value: string) => Boolean(value),
                      message: '端口不能为空',
                    },
                  ]}
                />
              )
            }
            {
              // 协议
              ![TcloudTemplatePort.Port, TcloudTemplatePort.Port_Group].includes(formModel.protocol) && (
                <InputColumn
                  v-model={formModel.port}
                  ref={portRef}
                  disabled={['ALL', 'icmp', 'gre', 'icmpv6'].includes(formModel.protocol)}
                  rules={[
                    {
                      validator: (value: string) => Boolean(value),
                      message: '端口不能为空',
                    },
                    {
                      validator: (value: string) => isPortAvailable(value),
                      message: '请填写合法的端口号, 注意需要在 1-65535 之间, 若需使用逗号时请注意使用英文逗号,',
                    },
                  ]}
                />
              )
            }
          </td>
          <td>
            <SelectColumn
              list={tcloudStrategys}
              v-model={formModel.action}
              ref={actionRef}
              rules={[
                {
                  validator: (value: string) => Boolean(value),
                  message: '策略不能为空',
                },
              ]}
            />
          </td>
          <td>
            <InputColumn
              v-model={formModel.memo}
              rules={[
                {
                  validator: (value: string) => value.length <= 100,
                  message: '备注长度最大不超过100个字符',
                },
              ]}
            />
          </td>
          {!props.isEdit && (
            <td>
              <OperationColumn
                showCopy
                onAdd={handleAdd}
                onRemove={handleRemove}
                onCopy={handleCopy}
                removeable={props.removeable}
              />
            </td>
          )}
        </tr>
      </>
    );
  },
});
