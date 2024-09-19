import { defineComponent, PropType, ref, watch } from 'vue';
import './index.scss';
import { SelectColumn, InputColumn, OperationColumn } from '@blueking/ediatable';
import { SecurityVendorType, useProtocols } from '../useProtocolList';
import useFormModel from '@/hooks/useFormModel';
import SourceAddress from '../tcloud/SourceAddress';
import { cleanObject, isPortAvailable, random } from '../util';
import { HUAWEI_TYPE_LIST } from '@/constants/resource';
import { Ext, IHead, SecurityRuleType } from '../useVendorHanlder';

export interface HuaweiSecurityGroupRule {
  protocol: string; // 协议类型, 取值范围: icmp、tcp、udp、icmpv6或IP协议号约束
  ethertype: string; // IP地址协议类型范围。(枚举值: IPv4、IPv6)
  cloud_remote_group_id?: string; // 远端安全组ID
  remote_ip_prefix?: string; // 远端IP地址
  port: string; // 端口取值范围
  priority: number; // 优先级取值范围: 1~100
  action: string; // 安全组规则生效策略
  memo?: string; // 备注
  key: string;
}

export const HuaweiRecord = (): Ext<HuaweiSecurityGroupRule> => ({
  protocol: '',
  ethertype: '',
  cloud_remote_group_id: '',
  remote_ip_prefix: '',
  port: '',
  priority: 0,
  action: '',
  memo: '',
  key: random(),
  sourceAddress: HuaweiSourceAddressType.IPV4,
});

export const huaweiTitles: (type: SecurityRuleType) => IHead[] = (type) => [
  {
    width: 450,
    minWidth: 120,
    title: '优先级',
    memo: '优先级可选范围为1-100，默认值为1，即最高优先级。优先级数字越小，规则优先级级别越高。',
  },
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
    title: '类型',
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
    memo: '请输入英文描述, 最大不超过256个字符',
    required: false,
  },
  {
    width: 450,
    minWidth: 120,
    title: '操作',
    required: false,
  },
];

export const huaweiSourceAddressTypes = [
  { value: 'ipv4', label: 'IPv4' },
  { value: 'ipv6', label: 'IPv6' },
  { value: 'remote_ip_prefix', label: 'IP地址' },
  { value: 'cloud_remote_group_id', label: '安全组' },
];

export enum HuaweiSourceAddressType {
  IPV4 = 'ipv4',
  IPV6 = 'ipv6',
  IP_ADDRESS = 'remote_ip_prefix',
  SECURITY_GROUP = 'cloud_remote_group_id',
}

export const TcloudSourceTypeArr = [
  HuaweiSourceAddressType.IPV4,
  HuaweiSourceAddressType.IPV6,
  HuaweiSourceAddressType.SECURITY_GROUP,
  HuaweiSourceAddressType.IP_ADDRESS,
];

export const huaweiStrategys = [
  { value: 'allow', label: '允许' },
  { value: 'deny', label: '拒绝' },
];

export const HuaweiRenderRow = defineComponent({
  props: {
    vendor: String as PropType<SecurityVendorType>,
    templateData: Object as PropType<{
      ipList: Array<string>;
      ipGroupList: Array<string>;
    }>,
    relatedSecurityGroups: Array as PropType<Array<Object>>,
    removeable: Boolean as PropType<boolean>,
    value: Object as PropType<HuaweiSecurityGroupRule>,
    isEdit: Boolean as PropType<boolean>,
  },
  emits: ['add', 'remove', 'copy', 'change'],
  setup(props, { expose, emit }) {
    const { protocols } = useProtocols(props.vendor);
    const { formModel } = useFormModel(props.value);

    const priorityRef = ref();
    const ethertypeRef = ref();
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

    const handleChange = (val: HuaweiSecurityGroupRule) => {
      emit('change', val);
    };

    watch(
      () => formModel.protocol,
      () => {
        if (['icmp', 'huaweiAll'].includes(formModel.protocol)) {
          formModel.port = 'ALL';
        } else formModel.port = '';
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
          priorityRef.value.getValue(),
          ethertypeRef.value.getValue(),
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
            <InputColumn
              ref={priorityRef}
              v-model={formModel.priority}
              type='number'
              min={1}
              max={100}
              rules={[
                {
                  validator: (value: string) => Boolean(value),
                  message: '优先级不能为空',
                },
              ]}
            />
          </td>
          <td>
            <SelectColumn
              list={huaweiSourceAddressTypes}
              v-model={formModel.sourceAddress}
              ref={sourceAddressTypeRef}
              rules={[
                {
                  validator: (value: string) => Boolean(value),
                  message: '源地址类型不能为空',
                },
              ]}
            />
          </td>
          <td>
            <SourceAddress
              v-model={formModel[formModel.sourceAddress]}
              {...props}
              sourceAddressType={formModel.sourceAddress as HuaweiSourceAddressType}
              ref={sourceAddressValRef}
            />
          </td>
          <td>
            <SelectColumn
              list={HUAWEI_TYPE_LIST.map(({ id, name }) => ({ label: name, value: id }))}
              v-model={formModel.ethertype}
              ref={ethertypeRef}
              rules={[
                {
                  validator: (value: string) => Boolean(value),
                  message: '类型不能为空',
                },
              ]}
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
            <InputColumn
              disabled={['icmp', 'huaweiAll'].includes(formModel.protocol)}
              v-model={formModel.port}
              ref={portRef}
              rules={[
                {
                  validator: (value: string) => Boolean(value),
                  message: '端口不能为空',
                },
                {
                  validator: (value: string) => {
                    return isPortAvailable(value);
                  },
                  message: '请填写合法的端口号, 注意需要在 1-65535 之间, 若需使用逗号时请注意使用英文逗号,',
                },
              ]}
            />
          </td>
          <td>
            <SelectColumn
              list={huaweiStrategys}
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
