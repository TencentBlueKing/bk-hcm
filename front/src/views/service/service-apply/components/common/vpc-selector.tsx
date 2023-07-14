import http from '@/http';
import { computed, defineComponent, PropType, ref, watch } from 'vue';
import { Select } from 'bkui-vue';

import { QueryRuleOPEnum } from '@/typings/common';
import { VendorEnum } from '@/common/constant';
import { useWhereAmI } from '@/hooks/useWhereAmI';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

const { Option } = Select;

export default defineComponent({
  props: {
    modelValue: String as PropType<string>,
    bizId: Number as PropType<number | string>,
    accountId: String as PropType<string>,
    vendor: String as PropType<string>,
    region: String as PropType<string>,
    zone: Array as PropType<string[]>,
  },
  emits: ['update:modelValue', 'change'],
  setup(props, { emit, attrs }) {
    const list = ref([]);
    const loading = ref(false);
    const { isResourcePage, isServicePage } = useWhereAmI();

    const selected = computed({
      get() {
        return props.modelValue;
      },
      set(val) {
        emit('update:modelValue', val);
      },
    });

    watch([
      () => props.bizId,
      () => props.accountId,
      () => props.vendor,
      () => props.region,
      () => props.zone,
    ], async ([bizId, accountId, vendor, region, zone]) => {
      if (!accountId || !region || !zone.length || (isServicePage && !bizId)) {
        list.value = [];
        return;
      }

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
          zone: props.zone.join(','),
          filter,
          page: {
            count: false,
            start: 0,
            limit: 50,
          },
        });
        list.value = result?.data?.details ?? [];
      } finally {
        loading.value = false;
      }
    });

    const handleChange = (val: string) => {
      const data = list.value.find(item => item.cloud_id === val);
      emit('change', data);
    };

    return () => (
      <Select
        filterable={true}
        modelValue={selected.value}
        onUpdate:modelValue={val => selected.value = val}
        onChange={handleChange}
        loading={loading.value}
        {...{ attrs }}
      >
        {
          list.value.map(({ cloud_id, name, current_zone_subnet_count, subnet_count, extension }) => (
            <Option key={cloud_id} value={cloud_id}
            // eslint-disable-next-line max-len
            disabled={(props.vendor === VendorEnum.TCLOUD || props.vendor === VendorEnum.AWS) && current_zone_subnet_count === 0}
            label={`${name} ${extension?.cidr ? extension?.cidr[0]?.cidr : ''} 该VPC共${subnet_count}个子网 
            ${(props.vendor === VendorEnum.TCLOUD || props.vendor === VendorEnum.AWS) ? `${`该可用区有${current_zone_subnet_count}个子网 ${current_zone_subnet_count === 0 ? '不可用' : '可用'}`}` : ''}`}></Option>
          ))
        }
      </Select>
    );
  },
});
