import http from '@/http';
import { computed, defineComponent, PropType, ref, watch } from 'vue';
import { Button, Select } from 'bkui-vue';

import { QueryRuleOPEnum } from '@/typings/common';
import { VendorEnum } from '@/common/constant';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import './vpc-selector.scss';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
const { Option } = Select;

export default defineComponent({
  props: {
    modelValue: String as PropType<string>,
    bizId: Number as PropType<number | string>,
    accountId: String as PropType<string>,
    vendor: String as PropType<VendorEnum>,
    region: String as PropType<string>,
    zone: String as PropType<string>,
    isSubnet: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    onRefreshVpcList: {
      type: Function,
      required: false,
    },
  },
  emits: ['update:modelValue', 'change'],
  setup(props, { emit, attrs }) {
    const list = ref([]);
    const loading = ref(false);
    const { isResourcePage, whereAmI } = useWhereAmI();

    const selected = computed({
      get() {
        return props.modelValue;
      },
      set(val) {
        emit('update:modelValue', val);
      },
    });

    watch(
      [() => props.bizId, () => props.accountId, () => props.vendor, () => props.region, () => props.zone],
      async ([bizId, accountId, vendor, region, zone]) => {
        console.log(accountId, region, zone, bizId);
        if (!accountId || !region || !zone || (whereAmI.value === Senarios.business && !bizId)) {
          list.value = [];
          return;
        }
        await refreshList(bizId, accountId, vendor, region);
        props.onRefreshVpcList?.(async () => {
          await refreshList(bizId, accountId, vendor, region);
          handleChange(props.modelValue);
        });
      },
    );

    const handleChange = (val: string) => {
      const data = list.value.find((item) => item.cloud_id === val);
      emit('change', data);
    };

    const refreshList = async (
      bizId: string | number = props.bizId,
      accountId: string = props.accountId,
      vendor: VendorEnum = props.vendor,
      region: string = props.region,
    ) => {
      try {
        loading.value = true;
        const filter = {
          op: 'and',
          rules: [
            {
              field: 'account_id',
              op: QueryRuleOPEnum.EQ,
              value: accountId,
            },
          ],
        };
        if (vendor !== VendorEnum.GCP) {
          filter.rules.push({
            field: 'region',
            op: QueryRuleOPEnum.EQ,
            value: region,
          });
        }
        const url = isResourcePage
          ? `${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/vendors/${props.vendor}/vpcs/with/subnet_count/list`
          : `${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/bizs/${bizId}/vendors/${props.vendor}/vpcs/with/subnet_count/list`;
        const result = await http.post(url, {
          // const result = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vpcs/list`, {
          zone: Array.isArray(props.zone) ? props.zone.join(',') : props.zone,
          filter,
          page: {
            count: false,
            start: 0,
            limit: 50,
          },
        });
        list.value = result?.data?.details ?? [];

        // 用户体验优化项
        const vendorFlag = [VendorEnum.TCLOUD, VendorEnum.AWS].includes(props.vendor);
        // 1.过滤
        let canUseVpcList = null;
        if (!props.isSubnet) {
          canUseVpcList = vendorFlag
            ? list.value.filter(({ current_zone_subnet_count }) => current_zone_subnet_count > 0)
            : list.value.filter(({ subnet_count }) => subnet_count > 0);
        }
        // 2.排序
        vendorFlag
          ? list.value.sort((prev, next) => next.current_zone_subnet_count - prev.current_zone_subnet_count)
          : list.value.sort((prev, next) => next.subnet_count - prev.subnet_count);
        // 3.自动填充
        if (canUseVpcList?.length === 1) {
          selected.value = canUseVpcList[0].cloud_id;
          emit('change', canUseVpcList[0]);
        }
      } finally {
        loading.value = false;
      }
    };

    return () => (
      <div class={'vpc-selector-container'}>
        <Select
          filterable={true}
          modelValue={selected.value}
          onUpdate:modelValue={(val) => (selected.value = val)}
          onChange={handleChange}
          loading={loading.value}
          {...{ attrs }}>
          {list.value.map(({ cloud_id, name, current_zone_subnet_count, subnet_count, extension }) => {
            return (
              <Option
                key={cloud_id}
                value={cloud_id}
                // eslint-disable-next-line max-len
                disabled={
                  !(
                    props.isSubnet ||
                    // eslint-disable-next-line max-len
                    ([VendorEnum.AZURE, VendorEnum.GCP, VendorEnum.HUAWEI].includes(props.vendor) &&
                      subnet_count > 0) ||
                    current_zone_subnet_count > 0
                  )
                }
                label={`${cloud_id} ${name} ${
                  extension?.cidr ? extension?.cidr[0]?.cidr : ''
                } 该VPC共${subnet_count}个子网
                  ${
                    props.vendor === VendorEnum.TCLOUD || props.vendor === VendorEnum.AWS
                      ? `${`该可用区有${current_zone_subnet_count}个子网 ${
                          current_zone_subnet_count === 0 ? '不可用' : '可用'
                        }`}`
                      : ''
                  }`}></Option>
            );
          })}
        </Select>
        {props.region && props.zone && !list.value.length ? (
          <span class={'vpc-selector-list-tip'}>
            该地域无可用的VPC网络，可切换地域，或点击
            <Button
              theme='primary'
              text
              onClick={() => {
                const url = whereAmI.value === Senarios.business ? '/#/business/vpc' : '/#/resource/resource?type=vpc';
                window.open(url, '_blank');
              }}>
              新建
            </Button>
          </span>
        ) : null}
      </div>
    );
  },
});
