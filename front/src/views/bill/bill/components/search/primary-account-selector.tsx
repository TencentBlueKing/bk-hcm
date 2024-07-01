import { VendorEnum } from '@/common/constant';
import { useSingleList } from '@/hooks/useSingleList';
import { QueryRuleOPEnum } from '@/typings';
import { Select } from 'bkui-vue';
import { PropType, computed, defineComponent, ref, watch } from 'vue';

export default defineComponent({
  props: { modelValue: String as PropType<string>, vendor: Array as PropType<VendorEnum[]> },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const selectedValue = ref(props.modelValue);
    const isMulVendor = computed(() => Array.isArray(props.vendor) && props.vendor.length);

    const { dataList, isDataLoad, handleScrollEnd, handleRefresh } = useSingleList({
      url: '/api/v1/account/root_accounts/list',
      rules: () => {
        if (!props.vendor || !isMulVendor.value) return [];
        return [{ field: 'vendor', op: QueryRuleOPEnum.IN, value: props.vendor }];
      },
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

    watch(
      () => props.vendor,
      () => {
        handleRefresh();
      },
      { deep: true },
    );

    return () => (
      <Select
        v-model={selectedValue.value}
        multiple
        multipleMode='tag'
        collapseTags
        clearable
        onScroll-end={handleScrollEnd}
        loading={isDataLoad.value}
        scrollLoading={isDataLoad.value}>
        {dataList.value.map(({ id, name }) => (
          <Select.Option key={id} id={id} name={name} />
        ))}
      </Select>
    );
  },
});
