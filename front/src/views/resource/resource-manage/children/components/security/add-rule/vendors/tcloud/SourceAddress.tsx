import { computed, defineComponent, PropType, ref, watch } from 'vue';
import { SelectColumn, InputColumn } from '@blueking/ediatable';
import { IpType, validateIpCidr } from '../util';
import { TcloudSecurityGroupRule } from '.';
import { AzureSourceAddressType, AzureTargetAddressType } from '../azure';
import { HuaweiSourceAddressType } from '../huawei';

export enum TcloudSourceAddressType {
  TEMPLATE_IP = 'cloud_address_id',
  TEMPLATE_IP_GROUP = 'cloud_address_group_id',
  SECURITY_GROUP = 'cloud_target_security_group_id',
  IPV4 = 'ipv4_cidr',
  IPV6 = 'ipv6_cidr',
}

export default defineComponent({
  props: {
    modelValue: String as PropType<string>,
    sourceAddressType: String as PropType<
      TcloudSourceAddressType | AzureSourceAddressType | AzureTargetAddressType | HuaweiSourceAddressType
    >,
    templateData: Object as PropType<{
      ipList: Array<string>;
      ipGroupList: Array<string>;
    }>,
    relatedSecurityGroups: Array as PropType<Array<Object>>,
    value: Object as PropType<TcloudSecurityGroupRule>,
    isCidr: Boolean as PropType<boolean>,
  },
  emits: ['update:modelValue'],
  setup(props, { emit, expose }) {
    const selectedVal = ref(props.modelValue);
    const instance = ref();

    const list = computed(() => {
      let res: any[] = [];
      if (
        [
          TcloudSourceAddressType.SECURITY_GROUP,
          AzureSourceAddressType.Security_Group,
          AzureTargetAddressType.Security_Group,
          HuaweiSourceAddressType.SECURITY_GROUP,
        ].includes(props.sourceAddressType)
      )
        res = props.relatedSecurityGroups;
      if ([TcloudSourceAddressType.TEMPLATE_IP].includes(props.sourceAddressType as TcloudSourceAddressType))
        res = props.templateData.ipList;
      if ([TcloudSourceAddressType.TEMPLATE_IP_GROUP].includes(props.sourceAddressType as TcloudSourceAddressType))
        res = props.templateData.ipGroupList;

      return res.map((v) => ({ label: v.name, value: v.cloud_id }));
    });

    watch(
      () => selectedVal.value,
      (val) => {
        emit('update:modelValue', val);
      },
    );

    watch(
      () => props.sourceAddressType,
      (_, oldVal) => {
        props.value[oldVal] = '';
        selectedVal.value = '';
      },
    );

    expose({
      getValue: () => instance.value.getValue(),
    });

    return () => (
      <>
        {[
          TcloudSourceAddressType.IPV4,
          TcloudSourceAddressType.IPV6,
          AzureSourceAddressType.IP,
          AzureTargetAddressType.IP,
          HuaweiSourceAddressType.IPV4,
          HuaweiSourceAddressType.IPV6,
          HuaweiSourceAddressType.IP_ADDRESS,
        ].includes(props.sourceAddressType) && (
          <InputColumn
            v-model={selectedVal.value}
            ref={instance}
            rules={[
              {
                message: '请填写对应合法的 IP, 注意区分 IPV4 与 IPV6',
                validator: (val: string) => {
                  if (
                    [AzureSourceAddressType.IP, AzureTargetAddressType.IP, HuaweiSourceAddressType.IP_ADDRESS].includes(
                      props.sourceAddressType,
                    )
                  )
                    return true;
                  const ipType = validateIpCidr(val);
                  if (ipType === IpType.invalid) return false;
                  if (
                    [IpType.ipv4, IpType.ipv4_cidr].includes(ipType) &&
                    ![TcloudSourceAddressType.IPV4, HuaweiSourceAddressType.IPV4].includes(props.sourceAddressType)
                  )
                    return false;
                  if (
                    [IpType.ipv6, IpType.ipv6_cidr].includes(ipType) &&
                    ![TcloudSourceAddressType.IPV6, HuaweiSourceAddressType.IPV6].includes(props.sourceAddressType)
                  )
                    return false;
                  return true;
                },
              },
              {
                message: '填写格式不正确。所有IPv4地址：0.0.0.0/0，所有IPv6地址：0::0/0或::/0',
                validator: (val: string) => {
                  return !['0.0.0.0', '0::0', '::'].includes(val);
                },
              },
              {
                message: '请填写合法的 IP',
                validator: (val: string) => {
                  return validateIpCidr(val) !== IpType.invalid;
                },
              },
              {
                message: '请填写合法的 IP CIDR',
                validator: (val: string) => {
                  if (!props.isCidr) return true;
                  return [IpType.ipv4_cidr, IpType.ipv6_cidr].includes(validateIpCidr(val));
                },
              },
            ]}
          />
        )}

        {[
          TcloudSourceAddressType.SECURITY_GROUP,
          TcloudSourceAddressType.TEMPLATE_IP,
          TcloudSourceAddressType.TEMPLATE_IP_GROUP,
          AzureSourceAddressType.Security_Group,
          AzureTargetAddressType.Security_Group,
          HuaweiSourceAddressType.SECURITY_GROUP,
        ].includes(props.sourceAddressType) && (
          <SelectColumn
            list={list.value}
            v-model={selectedVal.value}
            ref={instance}
            rules={[
              {
                validator: (value: string) => Boolean(value),
                message: '源地址不能为空',
              },
            ]}
          />
        )}
      </>
    );
  },
});
