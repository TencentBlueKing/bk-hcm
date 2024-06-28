import { VendorEnum } from '@/common/constant';
import { useSingleList } from '@/hooks/useSingleList';
import { QueryRuleOPEnum } from '@/typings';
import { Select } from 'bkui-vue';
import { PropType, computed, defineComponent, ref, watch } from 'vue';

export default defineComponent({
  props: {
    modelValue: String as PropType<string>,
    vendor: Array as PropType<VendorEnum[]>,
    rootAccountId: Array as PropType<string[]>,
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const selectedValue = ref(props.modelValue);
    const rules = computed(() => {
      const rules = [];
      if (props.vendor.length) rules.push({ field: 'vendor', op: QueryRuleOPEnum.IN, value: props.vendor });
      if (props.rootAccountId.length)
        rules.push({ field: 'parent_account_id', op: QueryRuleOPEnum.IN, value: props.rootAccountId });
      return rules;
    });

    const { dataList, isDataLoad, handleScrollEnd, handleRefresh } = useSingleList({
      url: '/api/v1/account/main_accounts/list',
      rules: () => rules.value,
      immediate: true,
    });

    watch(
      () => props.modelValue,
      (val) => {
        selectedValue.value = val;
      },
    );

    watch(selectedValue, (val) => {
      emit('update:modelValue', val);
    });

    watch(rules, () => handleRefresh(), { deep: true });

    return () => (
      <Select
        v-model={selectedValue.value}
        multiple
        multipleMode='tag'
        onScroll-end={handleScrollEnd}
        loading={isDataLoad.value}
        scrollLoading={isDataLoad.value}>
        {dataList.value.map(({ id, cloud_id }) => (
          <Select.Option key={id} id={id} name={cloud_id} />
        ))}
      </Select>
    );
  },
});
