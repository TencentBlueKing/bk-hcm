import http from '@/http';
import { computed, defineComponent, PropType, ref, watch } from 'vue';
import { Select } from 'bkui-vue';

import { formatStorageSize } from '@/common/util';
import { VendorEnum } from '@/common/constant';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

const { Option } = Select;

export default defineComponent({
  props: {
    modelValue: String as PropType<string>,
    vendor: String as PropType<string>,
    accountId: String as PropType<string>,
    region: String as PropType<string>,
    zone: String as PropType<string>,
    bizId: Number as PropType<number>,
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
      () => props.vendor,
      () => props.accountId,
      () => props.region,
      () => props.zone,
    ], async ([vendor, accountId, region, zone], [,,,oldZone]) => {
      if (!vendor || !accountId || !region || (vendor !== VendorEnum.AZURE && !zone)) {
        list.value = [];
        return;
      }

      // AZURE时与zone无关，只需要满足其它条件时请求一次
      if (vendor === VendorEnum.AZURE && zone !== oldZone) {
        return;
      }

      loading.value = true;
      const result = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/bizs/${props.bizId}/instance_types/list`, {
        account_id: accountId,
        vendor,
        region,
        zone,
      });
      list.value = result?.data ?? [];

      loading.value = false;
    });

    const handleChange = (val: string) => {
      const data = list.value.find(item => item.instance_type === val);
      emit('change', data);
    };

    return () => (
      <Select
        filterable={true}
        modelValue={selected.value}
        onUpdate:modelValue={val => selected.value = val}
        loading={loading.value}
        onChange={handleChange}
        {...{ attrs }}
      >
        {
          list.value.map(({ instance_type, cpu, memory, status }, index) => (
            <Option
              key={index}
              value={instance_type}
              label={`${instance_type} (${cpu}核CPU，${formatStorageSize(memory * 1024 ** 2)}内存)${status === 'SELL' ? '可购买' : '已售罄'}`}
            >
            </Option>
          ))
        }
      </Select>
    );
  },
});
