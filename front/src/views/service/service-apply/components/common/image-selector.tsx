import http from '@/http';
import { computed, defineComponent, PropType, ref, watch } from 'vue';
import { Select } from 'bkui-vue';
import { QueryRuleOPEnum } from '@/typings';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

const { Option } = Select;

export default defineComponent({
  props: {
    modelValue: String as PropType<string>,
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
      () => props.vendor,
      () => props.region,
    ], async ([vendor, region]) => {
      if (!vendor || !region) {
        list.value = [];
        return;
      }

      loading.value = true;
      const result = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/images/list`, {
        filter: {
          op: 'and',
          rules: [
            {
              field: 'vendor',
              op: QueryRuleOPEnum.EQ,
              value: vendor,
            },
            {
              field: 'extension.region',
              op: QueryRuleOPEnum.JSON_EQ,
              value: region,
            },
            {
              field: 'type',
              op: QueryRuleOPEnum.EQ,
              value: 'public',
            },
          ],
        },
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
            <Option key={cloud_id} value={cloud_id} label={name}></Option>
          ))
        }
      </Select>
    );
  },
});
