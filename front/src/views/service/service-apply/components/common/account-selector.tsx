import http from '@/http';
import { computed, defineComponent, PropType, ref, watchEffect } from 'vue';
import { Select } from 'bkui-vue';

import { IOption } from '@/typings/common';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

const { Option } = Select;

export default defineComponent({
  props: {
    modelValue: String as PropType<string>,
    bizId: Number as PropType<number>,
  },
  emits: ['update:modelValue', 'change'],
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
      if (props.bizId) {
        loading.value = true;
        const result = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/accounts/bizs/${props.bizId}`, {
          params: {
            account_type: 'resource',
          },
        });
        list.value = result?.data ?? [];
        loading.value = false;
      }
    });

    const handleChange = (val: string) => {
      const data = list.value.find(item => item.id === val);
      emit('change', data);
    };

    return () => (
      <Select
        clearable={false}
        filterable={true}
        modelValue={selected.value}
        onUpdate:modelValue={val => selected.value = val}
        onChange={handleChange}
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
