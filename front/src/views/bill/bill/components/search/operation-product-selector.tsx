import { VendorEnum } from '@/common/constant';
import { useSingleList } from '@/hooks/useSingleList';
import { QueryRuleOPEnum } from '@/typings';
import { Select } from 'bkui-vue';
import { PropType, defineComponent, ref, watch } from 'vue';

export default defineComponent({
  props: { modelValue: String as PropType<string>, vendor: String as PropType<VendorEnum> },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const selectedValue = ref(props.modelValue);

    const { dataList, isDataLoad, loadDataList, handleScrollEnd } = useSingleList({
      url: '/api/v1/web/operation_products/list',
      rules: () => (props.vendor ? [{ field: 'vendor', op: QueryRuleOPEnum.EQ, value: props.vendor }] : []),
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
        loadDataList();
      },
    );

    return () => (
      <Select
        v-model={selectedValue.value}
        multiple
        multipleMode='tag'
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
