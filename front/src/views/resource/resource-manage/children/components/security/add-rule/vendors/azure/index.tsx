import { defineComponent, PropType, ref, watch } from 'vue';
import './index.scss';
import { SelectColumn, InputColumn, OperationColumn } from '@blueking/ediatable';
import { SecurityVendorType } from '../useProtocolList';
import useFormModel from '@/hooks/useFormModel';
import SourceAddress from '../tcloud/SourceAddress';
import { cleanObject, isPortAvailable, random } from '../util';
import { Ext, IHead } from '../useVendorHanlder';
import { AZURE_PROTOCOL_LIST } from '@/constants';

export interface AzureSecurityGroupRule {
  name: string; // 资源组唯一的资源名称
  memo?: string; // 备注
  destination_address_prefix?: string; // 目的地址前缀
  destination_address_prefixes?: string; // 目的地址带有前缀
  destination_port_range?: string; // 目标端口或范围
  destination_port_ranges?: string; // 目的端口范围
  protocol: string; // 网络协议
  source_address_prefix?: string; // 源地址前缀
  source_address_prefixes?: string; // 源地址带有前缀
  source_port_range?: string; // 源端口或范围
  source_port_ranges?: string; // 源端口范围
  priority: number; // 规则的优先级
  access: string; // 允许或拒绝网络流量
  key: string;
}

export const AzureRecord = (): Ext<AzureSecurityGroupRule> => ({
  name: '',
  memo: '',
  destination_address_prefix: '',
  destination_address_prefixes: '',
  destination_port_range: '',
  destination_port_ranges: '',
  protocol: '',
  source_address_prefix: '',
  source_address_prefixes: '',
  source_port_range: '',
  source_port_ranges: '',
  priority: 0,
  access: '',
  key: random(),
  sourceAddress: AzureSourceAddressType.IP,
  targetAddress: AzureTargetAddressType.IP,
});

export const azureTitles: IHead[] = [
  {
    width: 450,
    minWidth: 120,
    title: '名称',
  },
  {
    width: 450,
    minWidth: 120,
    title: '优先级',
    memo: '根据优先级顺序处理规则；数字越小，优先级越高。我们建议在规则之间留出间隙「100、200、300」等。这样一来便可在无需编辑现有规则的情况下添加新规，同时注意不能和当前已有规则的优先级重复。取值范围为100-4096',
  },
  {
    width: 450,
    minWidth: 120,
    title: '源地址类型',
  },
  {
    width: 450,
    minWidth: 120,
    title: '源地址',
    memo: '源过滤器可为“任意”、一个 IP 地址范围、一个应用程序安全组或一个默认标记。它指定此规则将允许或拒绝的特定源 IP 地址范围的传入流量',
  },
  {
    width: 450,
    minWidth: 120,
    title: '源端口',
    memo: '提供单个端口(如 80)、端口范围(如 1024-65535)，或单个端口和/或端口范围的以逗号分隔的列表(如 80,1024-65535)。这指定了根据此规则将允许或拒绝哪些端口的流量。提供星号(*)可允许任何端口的流量',
  },
  {
    width: 450,
    minWidth: 120,
    title: '目标地址类型',
  },
  {
    width: 450,
    minWidth: 120,
    title: '目标地址',
    memo: '提供采用 CIDR 表示法的地址范围(例如 192.168.99.0/24 或 2001:1234::/64)或提供 IP 地址(例如 192.168.99.0 或 2001:1234::)。\n\r 还可提供一个由采用 IPv4 或 IPv6 的 IP 地址或地址范围构成的列表(以逗号分隔)',
  },
  {
    width: 450,
    minWidth: 120,
    title: '目标协议端口类型',
  },
  {
    width: 450,
    minWidth: 120,
    title: '目标协议端口',
  },
  {
    width: 450,
    minWidth: 120,
    title: '策略',
    memo: '请输入英文描述, 最大不超过256个字符',
  },
  {
    width: 450,
    minWidth: 120,
    title: '备注',
    required: false,
  },
  {
    width: 450,
    minWidth: 120,
    title: '操作',
    required: false,
  },
];

export const azureSourceAddressTypes = [
  { value: 'source_address_prefix', label: 'IP地址' },
  {
    value: 'cloud_source_security_group_ids',
    label: '安全组',
  },
];

export enum AzureSourceAddressType {
  IP = 'source_address_prefix',
  Security_Group = 'cloud_source_security_group_ids',
}

export const AzureSourceTypeArr = [AzureSourceAddressType.IP, AzureSourceAddressType.Security_Group];

export const azureTargetAddressTypes = [
  {
    value: 'destination_address_prefix',
    label: 'IP地址',
  },
  {
    value: 'cloud_destination_security_group_ids',
    label: '安全组',
  },
];

export enum AzureTargetAddressType {
  IP = 'destination_address_prefix',
  Security_Group = 'cloud_destination_security_group_ids',
}

export const AzureTargetTypeArr = [AzureTargetAddressType.IP, AzureTargetAddressType.Security_Group];

export const azureStrategys = [
  { value: 'Allow', label: '允许' },
  { value: 'Deny', label: '拒绝' },
];

export const AzureRenderRow = defineComponent({
  props: {
    vendor: String as PropType<SecurityVendorType>,
    templateData: Object as PropType<{
      ipList: Array<string>;
      ipGroupList: Array<string>;
    }>,
    relatedSecurityGroups: Array as PropType<Array<Object>>,
    removeable: Boolean as PropType<boolean>,
    value: Object as PropType<AzureSecurityGroupRule>,
    isEdit: Boolean as PropType<boolean>,
  },
  emits: ['add', 'remove', 'copy', 'change'],
  setup(props, { expose, emit }) {
    const { formModel } = useFormModel(props.value);

    const nameRef = ref();
    const priorityRef = ref();
    const sourceAddressTypeRef = ref();
    const sourceAddressValRef = ref();
    const sourceAddressValPort = ref();
    const targetAddressTypeRef = ref();
    const targetAddressValRef = ref();
    const targetProtocolTypeRef = ref();
    const targetProtocolValRef = ref();
    const accessRef = ref();

    const handleAdd = () => {
      emit('add');
    };

    const handleRemove = () => {
      emit('remove');
    };

    const handleCopy = () => {
      emit('copy', formModel);
    };

    const handleChange = (val: AzureSecurityGroupRule) => {
      emit('change', val);
    };

    watch(
      () => formModel,
      (val) => {
        handleChange(val);
      },
      {
        deep: true,
      },
    );

    watch(
      () => formModel.protocol,
      (val) => {
        if (['*', 'Icmp'].includes(val)) {
          formModel.destination_port_range = 'ALL';
        } else formModel.destination_port_range = '';
      },
    );

    expose({
      getValue: async () => {
        await Promise.all([
          nameRef.value.getValue(),
          priorityRef.value.getValue(),
          sourceAddressTypeRef.value.getValue(),
          sourceAddressValRef.value.getValue(),
          sourceAddressValPort.value.getValue(),
          targetAddressTypeRef.value.getValue(),
          targetAddressValRef.value.getValue(),
          targetProtocolTypeRef.value.getValue(),
          targetProtocolValRef.value.getValue(),
          accessRef.value.getValue(),
        ]);
        return cleanObject(formModel);
      },
    });

    return () => (
      <>
        <tr>
          <td>
            <InputColumn
              ref={nameRef}
              v-model={formModel.name}
              rules={[
                {
                  validator: (value: string) => Boolean(value),
                  message: '名称不能为空',
                },
              ]}
            />
          </td>
          <td>
            <InputColumn
              ref={priorityRef}
              v-model={formModel.priority}
              type='number'
              min={100}
              max={4096}
              rules={[
                {
                  validator: (value: string) => Boolean(value),
                  message: '优先级不能为空',
                },
              ]}
            />
          </td>
          <td>
            <SelectColumn list={azureSourceAddressTypes} v-model={formModel.sourceAddress} ref={sourceAddressTypeRef} />
          </td>
          <td>
            <SourceAddress
              ref={sourceAddressValRef}
              {...props}
              v-model={formModel[formModel.sourceAddress]}
              sourceAddressType={formModel.sourceAddress}
            />
          </td>
          <td>
            <InputColumn
              ref={sourceAddressValPort}
              v-model={formModel.source_port_range}
              rules={[
                {
                  validator: (value: string) => {
                    return Boolean(value);
                  },
                  message: '端口不能为空',
                },
                {
                  validator: (value: string) => {
                    return isPortAvailable(value);
                  },
                  message: '请填写合法的端口号, 注意需要在 1-65535 之间',
                },
                {
                  validator: (value: string) => {
                    return !/,/.test(value);
                  },
                  message: '请填写合法的端口号,不支持逗号分隔',
                },
              ]}
            />
          </td>
          <td>
            <SelectColumn list={azureTargetAddressTypes} v-model={formModel.targetAddress} ref={targetAddressTypeRef} />
          </td>
          <td>
            <SourceAddress
              {...props}
              v-model={formModel[formModel.targetAddress]}
              sourceAddressType={formModel.targetAddress}
              ref={targetAddressValRef}
            />
          </td>
          <td>
            <SelectColumn
              list={AZURE_PROTOCOL_LIST.map(({ id, name }) => ({ label: name, value: id }))}
              v-model={formModel.protocol}
              ref={targetProtocolTypeRef}
              rules={[
                {
                  validator: (value: string) => Boolean(value),
                  message: '目标协议端口类型不能为空',
                },
              ]}
            />
          </td>
          <td>
            <InputColumn
              v-model={formModel.destination_port_range}
              ref={targetProtocolValRef}
              disabled={['*', 'Icmp'].includes(formModel.protocol)}
              rules={[
                {
                  validator: (value: string) => Boolean(value),
                  message: '端口不能为空',
                },
                {
                  validator: (value: string) => isPortAvailable(value),
                  message: '请填写合法的端口号, 注意需要在 1-65535 之间, 若需使用逗号时请注意使用英文逗号,',
                },
                {
                  validator: (value: string) => {
                    return !/,/.test(value);
                  },
                  message: '请填写合法的端口号,不支持逗号分隔',
                },
              ]}
            />
          </td>
          <td>
            <SelectColumn
              list={azureStrategys}
              v-model={formModel.access}
              ref={accessRef}
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
                  validator: (value: string) => value.length <= 256,
                  message: '备注长度不能超过256个字符',
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
