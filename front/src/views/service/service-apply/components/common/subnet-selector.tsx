import http from '@/http';
import { computed, defineComponent, PropType, ref, watch } from 'vue';
import { Select } from 'bkui-vue';

import { IOption, QueryRuleOPEnum } from '@/typings/common';
import { VendorEnum } from '@/common/constant';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

const { Option } = Select;

export default defineComponent({
  props: {
    modelValue: String as PropType<string>,
    bizId: Number as PropType<number>,
    vpcId: String as PropType<string>,
    vendor: String as PropType<string>,
    region: String as PropType<string>,
  },
  emits: ['update:modelValue'],
  setup(props, { emit, attrs }) {
    const list = ref([]);
    const loading = ref(false);

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
      () => props.vpcId
    ], async ([bizId, region, vendor, vpcId]) => {
      if (!bizId || !vpcId) {
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
        ],
      };

      if (vendor === VendorEnum.GCP) {
        filter.rules.push({
          field: 'region',
          op: QueryRuleOPEnum.EQ,
          value: region,
        })
      }

      const result = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/bizs/${bizId}/subnets/list`, {
      // const result = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/subnets/list`, {
        filter,
        page: {
          count: false,
          start: 0,
          limit: 500,
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
          list.value.map(({ cloud_id, name }) => (
            <Option key={cloud_id} value={cloud_id} label={`${cloud_id}${name ? `(${name})` : ''}`}></Option>
          ))
        }
      </Select>
    );
  },
});
