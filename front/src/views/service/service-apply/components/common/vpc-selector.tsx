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
    accountId: String as PropType<string>,
    vendor: String as PropType<string>,
    region: String as PropType<string>,
  },
  emits: ['update:modelValue', 'change'],
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
      () => props.accountId,
      () => props.vendor,
      () => props.region,
    ], async ([bizId, accountId, vendor, region]) => {
      if (!bizId || !accountId || !region) {
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
        }
        if (vendor !== VendorEnum.GCP) {
          filter.rules.push({
            field: 'region',
            op: QueryRuleOPEnum.EQ,
            value: region,
          });
        }
        const result = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/bizs/${bizId}/vpcs/list`, {
        // const result = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vpcs/list`, {
          filter,
          page: {
            count: false,
            start: 0,
            limit: 500,
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
          list.value.map(({ cloud_id, name }) => (
            <Option key={cloud_id} value={cloud_id} label={name}></Option>
          ))
        }
      </Select>
    );
  },
});
