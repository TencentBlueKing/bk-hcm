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
    vpcId: String as PropType<string>,
    vendor: String as PropType<string>,
    region: String as PropType<string>,
    accountId: String as PropType<string>,
    zone: Array as PropType<string[]>,
    resourceGroup: String as PropType<string>,
  },
  emits: ['update:modelValue'],
  setup(props, { emit, attrs, expose }) {
    const list = ref([]);
    const loading = ref(false);
    const { isResourcePage, isServicePage } = useWhereAmI();

    expose({ subnetList: list });

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
      () => props.region,
      () => props.vendor,
      () => props.vpcId,
      () => props.accountId,
      () => props.zone,
      () => props.resourceGroup,
    ], async ([bizId, region, vendor, vpcId, accountId, zone]) => {
      if ((!bizId && isServicePage) || !vpcId) {
        list.value = [];
        return;
      }

      loading.value = true;

      const filter = {
        op: 'and',
        rules: [
          {
            field: 'vpc_id',
            op: QueryRuleOPEnum.EQ,
            value: vpcId,
          },
          {
            field: 'account_id',
            op: QueryRuleOPEnum.EQ,
            value: accountId,
          },
          {
            field: 'region',
            op: QueryRuleOPEnum.EQ,
            value: region,
          },
        ],
      };

      if ([VendorEnum.TCLOUD, VendorEnum.AWS].includes(vendor)) {
        filter.rules.push({
          field: 'zone',
          op: QueryRuleOPEnum.EQ,
          value: zone[0] || '',
        });
      }

      // if (vendor === VendorEnum.AZURE) {
      //   filter.rules.push({
      //     field: 'extension.resource_group_name',
      //     op: QueryRuleOPEnum.JSON_EQ,
      //     value: resourceGroup,
      //   });
      // }

      const result = await http.post(isResourcePage
        ? `${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/subnets/with/ip_count/list`
        : `${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/bizs/${bizId}/subnets/with/ip_count/list`, {
        // const result = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/subnets/list`, {
        filter,
        page: {
          count: false,
          start: 0,
          limit: 50,
        },
      });
      list.value = result?.data?.details ?? [];
      loading.value = false;
    });

    return () => (
      <Select
        filterable={true}
        modelValue={selected.value}
        onUpdate:modelValue={val => selected.value = val}
        loading={loading.value}
        {...{ attrs }}
      >
        {
          list.value.map(({ cloud_id, name, ipv4_cidr, available_ip_count }) => (
            <Option key={cloud_id} value={cloud_id} label={`${name} ${ipv4_cidr} ${props.vendor !== VendorEnum.GCP ? `剩余IP ${available_ip_count}` : ''}`}></Option>
          ))
        }
      </Select>
    );
  },
});
