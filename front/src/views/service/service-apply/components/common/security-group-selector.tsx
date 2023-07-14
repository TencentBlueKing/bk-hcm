import http from '@/http';
import { computed, defineComponent, PropType, ref, watch } from 'vue';
import { Select } from 'bkui-vue';
import { VendorEnum } from '@/common/constant';
import { useWhereAmI } from '@/hooks/useWhereAmI';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

const { Option } = Select;

export default defineComponent({
  props: {
    modelValue: String as PropType<string>,
    bizId: Number as PropType<number | string>,
    accountId: String as PropType<string>,
    region: String as PropType<string>,
    multiple: Boolean as PropType<boolean>,
    vendor: String as PropType<string>,
    vpcId: String as PropType<string>,
  },
  emits: ['update:modelValue'],
  setup(props, { emit, attrs }) {
    const list = ref([]);
    const loading = ref(false);
    const { isServicePage } = useWhereAmI();

    const selected = computed({
      get() {
        return props.modelValue;
      },
      set(val) {
        emit('update:modelValue', val);
      },
    });

    const isSelected = computed(() => {
      if (selected.value) {
        return !!(Object.keys(selected.value).length);
      }
      return false;
    });

    watch([
      () => props.bizId,
      () => props.accountId,
      () => props.region,
      () => props.vpcId,
    ], async ([bizId, accountId, region, vpcId]) => {
      if ((!bizId && isServicePage) || !accountId || !region) {
        list.value = [];
        return;
      }
      loading.value = true;
      // const result = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/bizs/${bizId}/security_groups/list`, {
      const rules = [
        {
          field: 'account_id',
          op: 'eq',
          value: accountId,
        },
        {
          field: 'region',
          op: 'eq',
          value: region,
        },
      ];
      if (props.vendor === VendorEnum.AWS) {
        rules.push({
          field: 'extension.vpc_id',
          op: 'json_eq',
          value: vpcId,
        });
      }
      const result = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/security_groups/list`, {
        filter: {
          op: 'and',
          rules,
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
        multiple={props.multiple}
        loading={loading.value}
        class={isSelected.value && 'security-group-cls'}
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
