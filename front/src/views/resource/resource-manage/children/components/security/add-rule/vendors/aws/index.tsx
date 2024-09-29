import { defineComponent, PropType, ref, watch } from 'vue';
import { SelectColumn, InputColumn, OperationColumn } from '@blueking/ediatable';
import './index.scss';
import SourceAddress from '../tcloud/SourceAddress';
import useFormModel from '@/hooks/useFormModel';
import { SecurityVendorType, useProtocols } from '../useProtocolList';
import { cleanObject, isPortAvailable, random } from '../util';
import { Ext, IHead, SecurityRuleType } from '../useVendorHanlder';
import { AWS_PORT_ALL, AWS_PROTOCOL } from './DataHandler';

export interface AwsSecurityGroupRule {
  protocol: string; // 协议, 取值: tcp, udp, icmp, icmpv6,用数字 -1 代表所有协议
  from_port: number; // 起始端口，与 to_port 配合使用。-1代表所有端口。
  to_port: number; // 结束端口，与 from_port 配合使用。-1代表所有端口。
  ipv4_cidr?: string; // IPv4网段 (可选)
  ipv6_cidr?: string; // IPv6网段 (可选)
  cloud_target_security_group_id?: string; // 下一跳安全组实例云ID (可选)
  memo?: string; // 备注 (可选)
  key: string;
}

export const AwsRecord = (): Ext<AwsSecurityGroupRule> => ({
  protocol: '',
  from_port: -1,
  to_port: -1,
  ipv4_cidr: '',
  ipv6_cidr: '',
  cloud_target_security_group_id: '',
  memo: '',
  key: random(),
  port: '',
  sourceAddress: AwsSourceAddressType.IPV4,
});

export const awsTitles: (type: SecurityRuleType) => IHead[] = (type) => [
  {
    width: 120,
    title: type === 'ingress' ? '源地址类型' : '目标地址类型',
  },
  {
    width: 120,
    title: type === 'ingress' ? '源地址' : '目标地址',
    memo: '必须指定 CIDR 数据块 或者 安全组 ID',
  },
  {
    width: 120,
    title: '协议',
  },
  {
    width: 120,
    title: '端口',
    memo: '对于 TCP、UDP 协议，允许的端口范围。您可以指定单个端口号（例如 22）或端口号范围（例如7000-8000）',
  },
  {
    width: 120,
    title: '备注',
    memo: '请输入英文描述, 最大不超过256个字符',
    required: false,
  },
  {
    width: 120,
    title: '操作',
    required: false,
  },
];

export const awsSourceAddressTypes = [
  { value: 'ipv4_cidr', label: 'IPv4' },
  { value: 'ipv6_cidr', label: 'IPv6' },
  { value: 'cloud_target_security_group_id', label: '安全组' },
];

export enum AwsSourceAddressType {
  IPV4 = 'ipv4_cidr',
  IPV6 = 'ipv6_cidr',
  Security_Group = 'cloud_target_security_group_id',
}

export const AwsSourceTypeArr = [
  AwsSourceAddressType.IPV4,
  AwsSourceAddressType.IPV6,
  AwsSourceAddressType.Security_Group,
];

export const AwsRenderRow = defineComponent({
  props: {
    vendor: String as PropType<SecurityVendorType>,
    templateData: Object as PropType<{
      ipList: Array<string>;
      ipGroupList: Array<string>;
    }>,
    relatedSecurityGroups: Array as PropType<Array<Object>>,
    removeable: Boolean as PropType<Boolean>,
    value: Object as PropType<Ext<AwsSecurityGroupRule>>,
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

    const handleAdd = () => {
      emit('add');
    };

    const handleRemove = () => {
      emit('remove');
    };

    const handleCopy = () => {
      emit('copy', formModel);
    };

    const handleChange = (val: AwsSecurityGroupRule) => {
      emit('change', val);
    };

    watch(
      () => formModel.protocol,
      (protocol) => {
        if ([AWS_PROTOCOL.ALL, AWS_PROTOCOL.ICMP, AWS_PROTOCOL.ICMPv6].includes(protocol as AWS_PROTOCOL)) {
          formModel.port = AWS_PORT_ALL;
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
          protocolRef.value.getValue(),
          portRef.value.getValue(),
          sourceAddressTypeRef.value.getValue(),
          sourceAddressValRef.value.getValue(),
        ]);
        return cleanObject(formModel);
      },
    });

    return () => (
      <>
        <tr>
          <td>
            <SelectColumn
              list={awsSourceAddressTypes}
              v-model={formModel.sourceAddress}
              ref={sourceAddressTypeRef}
              rules={[
                {
                  validator: (value: string) => Boolean(value),
                  message: '源地址类型不能为空',
                },
                {
                  validator: (value: string) =>
                    (formModel.protocol === 'icmpv6' && value !== AwsSourceAddressType.IPV4) ||
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
              sourceAddressType={formModel.sourceAddress}
              ref={sourceAddressValRef}
              isCidr
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
              disabled={['-1', 'icmp', 'icmpv6'].includes(formModel.protocol)}
              v-model={formModel.port}
              ref={portRef}
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
                  message: '请填写合法的端口号, 注意需要在 1-65535 之间, 若需使用逗号时请注意使用英文逗号,',
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
