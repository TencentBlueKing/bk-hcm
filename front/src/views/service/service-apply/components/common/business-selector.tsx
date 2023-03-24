import http from '@/http';
import { computed, defineComponent, PropType, ref, watchEffect } from 'vue';
import { Select } from 'bkui-vue';

import { IOption } from '@/typings/common';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

const { Option } = Select;

export default defineComponent({
  props: {
    modelValue: Number as PropType<number>,
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
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
      loading.value = true;
      const result = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/authorized/bizs/list`);
      list.value = result?.data ?? [];
      loading.value = false;
    });

    return () => (
      <Select
        clearable={false}
        filterable={true}
        modelValue={selected.value}
        onUpdate:modelValue={val => selected.value = val}
        loading={loading.value}
      >
        {
          list.value.map(({ id, name }: IOption) => (
            <Option key={id} value={id} label={name}></Option>
          ))
        }
      </Select>
    );
  },
});
