import http from '@/http';
import { computed, defineComponent, PropType, ref, watchEffect } from 'vue';
import { Select } from 'bkui-vue';

import { IOption, QueryFilterType, QueryRuleOPEnum } from '@/typings/common';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

const { Option } = Select;

export default defineComponent({
  props: {
    modelValue: String as PropType<string>,
    accountId: String as PropType<string>,
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

    watchEffect(async () => {
      const filter: QueryFilterType = {
        op: 'and',
        rules: [
          {
            field: 'type',
            op: QueryRuleOPEnum.EQ,
            value: 'Microsoft.Resources/resourceGroups',
          },
          {
            field: 'account_id',
            op: QueryRuleOPEnum.EQ,
            value: props.accountId,
          },
        ],
      };

      loading.value = true;
      const result = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/azure/resource_groups/list`, {
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
        clearable={false}
        filterable={true}
        modelValue={selected.value}
        onUpdate:modelValue={(val) => (selected.value = val)}
        loading={loading.value}
        {...{ attrs }}>
        {list.value.map(({ id, name }: IOption) => (
          <Option key={id} value={name} label={name}></Option>
        ))}
      </Select>
    );
  },
});
