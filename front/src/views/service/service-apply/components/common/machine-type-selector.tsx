import http from '@/http';
import { computed, defineComponent, PropType, ref, watch } from 'vue';
import { Select } from 'bkui-vue';

import { formatStorageSize } from '@/common/util';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

const { Option } = Select;

export default defineComponent({
  props: {
    modelValue: String as PropType<string>,
    vendor: String as PropType<string>,
    accountId: String as PropType<string>,
    region: String as PropType<string>,
    zone: String as PropType<string>,
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
      () => props.vendor,
      () => props.accountId,
      () => props.region,
      () => props.zone,
    ], async ([vendor, accountId, region, zone]) => {
      if (!vendor || !accountId || !region || !zone) {
        list.value = [];
        return;
      }

      loading.value = true;
      const result = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/instance_types/list`, {
        account_id: accountId,
        vendor,
        region,
        zone,
      });
      list.value = result?.data ?? [];

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
          list.value.map(({ instance_type, cpu, memory }, index) => (
            <Option
              key={index}
              value={instance_type}
              label={`${instance_type} (${cpu}核CPU，${formatStorageSize(memory * 1024 ** 2)}内存)`}
            >
            </Option>
          ))
        }
      </Select>
    );
  },
});
