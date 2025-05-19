import { computed, defineComponent, PropType, ref, watch } from 'vue';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useSingleList } from '@/hooks/useSingleList';
import { QueryRuleOPEnum, RulesItem } from '@/typings/common';
import { ResourceTypeEnum } from '@/common/resource-constant';
import { VendorEnum } from '@/common/constant';
import { ISubnetItem } from '../../cvm/children/SubnetPreviewDialog';

import RightTurnLine from 'bkui-vue/lib/icon/right-turn-line';

export default defineComponent({
  props: {
    modelValue: String as PropType<string>,
    bizId: Number as PropType<number | string>,
    vpcId: String as PropType<string>,
    vendor: String as PropType<string>,
    region: String as PropType<string>,
    accountId: String as PropType<string>,
    zone: [String, Array<String>] as PropType<string | string[]>,
    resourceGroup: String as PropType<string>,
    handleChange: Function as PropType<(data: ISubnetItem) => void>,
    clearable: { type: Boolean, default: true },
    resourceType: String as PropType<ResourceTypeEnum>,
    optionDisabled: {
      type: Function as PropType<(subnetItem: ISubnetItem) => boolean>,
      default: () => false,
    },
  },
  emits: ['update:modelValue'],
  setup(props, { emit, attrs, expose }) {
    const { getBusinessApiPath, isServicePage } = useWhereAmI();

    const selected = computed({
      get() {
        return props.modelValue;
      },
      set(val) {
        emit('update:modelValue', val);
      },
    });

    const url = `/api/v1/web/${getBusinessApiPath()}subnets/with/ip_count/list`;
    const rules = ref<RulesItem[]>([]);
    const { dataList, isDataLoad, handleReset, handleRefresh } = useSingleList<ISubnetItem>({
      url,
      rules: () => rules.value,
      rollRequestConfig: { enabled: true, limit: 50 },
    });

    const handleChange = (cloud_id: string) => {
      const data = dataList.value.find((item) => item.cloud_id === cloud_id);
      typeof props.handleChange === 'function' && props.handleChange(data);
    };

    const getSubnetsData = async (
      bizId: string | number,
      region: string,
      vendor: string,
      vpcId: string,
      accountId: string,
      zone: string | string[],
    ) => {
      if ((!bizId && isServicePage) || !vpcId) {
        handleReset();
        return;
      }

      const filter: RulesItem[] = [
        { field: 'vpc_id', op: QueryRuleOPEnum.EQ, value: vpcId },
        { field: 'account_id', op: QueryRuleOPEnum.EQ, value: accountId },
        { field: 'region', op: QueryRuleOPEnum.EQ, value: region },
      ];

      if ([VendorEnum.TCLOUD, VendorEnum.AWS].includes(vendor as VendorEnum)) {
        if (Array.isArray(zone)) {
          zone.length > 0 && filter.push({ field: 'zone', op: QueryRuleOPEnum.IN, value: zone });
        } else {
          // CLB可能zone字段为空，但CVM一定不为空
          zone && filter.push({ field: 'zone', op: QueryRuleOPEnum.EQ, value: zone });
        }
      }

      // if (vendor === VendorEnum.AZURE) {
      //   filter.rules.push({
      //     field: 'extension.resource_group_name',
      //     op: QueryRuleOPEnum.JSON_EQ,
      //     value: resourceGroup,
      //   });
      // }

      // 更新搜索条件
      rules.value = filter;

      handleRefresh();
    };

    watch(
      [
        () => props.bizId,
        () => props.region,
        () => props.vendor,
        () => props.vpcId,
        () => props.accountId,
        () => props.zone,
        () => props.resourceGroup,
      ],
      ([bizId, region, vendor, vpcId, accountId, zone]) => {
        if (props.resourceType === ResourceTypeEnum.CLB && region && vendor && vpcId && accountId) {
          getSubnetsData(bizId, region, vendor, vpcId, accountId, zone);
        } else {
          if (region && vendor && vpcId && accountId && zone) {
            getSubnetsData(bizId, region, vendor, vpcId, accountId, zone);
          }
        }
      },
      { immediate: true },
    );

    const optionRender = () => {
      return dataList.value.map((subnet) => {
        const { cloud_id, name, ipv4_cidr, ipv6_cidr, available_ip_count } = subnet;
        const ipv4CidrStr = ipv4_cidr ? ` ${ipv4_cidr.join(',')}` : '';
        const ipv6CidrStr = ipv6_cidr ? ` ${ipv6_cidr.join(',')}` : '';
        const label =
          props.vendor !== VendorEnum.GCP
            ? `${cloud_id} ${name}${ipv4CidrStr}${ipv6CidrStr} 剩余IP:${available_ip_count}`
            : `${cloud_id} ${name}${ipv4CidrStr}${ipv6CidrStr}`;

        return <bk-option key={cloud_id} value={cloud_id} label={label} disabled={props.optionDisabled(subnet)} />;
      });
    };

    const userGuideRender = () => {
      if (props.vpcId && !dataList.value.length && !isDataLoad.value) {
        return (
          <div class={'subnet-selector-tips'}>
            <span class={'subnet-create-tips'}>{'所选的VPC，在当前区无可用的子网，可切换VPC或'}</span>
            <bk-button
              class='mr8'
              text
              theme='primary'
              onClick={() => {
                const url = '/#/resource/resource?type=subnet';
                window.open(url, '_blank');
              }}>
              新建子网
            </bk-button>
            <bk-button
              text
              onClick={() => {
                getSubnetsData(props.bizId, props.region, props.vendor, props.vpcId, props.accountId, props.zone);
              }}>
              <RightTurnLine fill='#3A84FF' />
            </bk-button>
          </div>
        );
      }
      return null;
    };

    expose({ subnetList: dataList });

    return () => (
      <div>
        <bk-select
          {...{ attrs }}
          v-model={selected.value}
          loading={isDataLoad.value}
          filterable
          clearable={props.clearable}
          onChange={handleChange}>
          {optionRender()}
        </bk-select>
        {/* 用户指引 */}
        {userGuideRender()}
      </div>
    );
  },
});
