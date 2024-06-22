import { VendorEnum } from '@/common/constant';
import { useSingleList } from '@/hooks/useSingleList';
import { Select } from 'bkui-vue';
import { PropType, defineComponent, ref, watch } from 'vue';

export default defineComponent({
  props: {
    modelValue: String as PropType<string>,
    vendor: String as PropType<VendorEnum>,
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const selectedValue = ref(props.modelValue);

    const { dataList, isDataLoad, loadDataList, handleScrollEnd } = useSingleList({
      url: '/api/v1/account/operation_products/list',
      payload: () => ({ filter: undefined }),
      immediate: true,
      disableSort: true,
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
        {dataList.value.map(({ op_product_id, op_product_name }) => (
          <Select.Option key={op_product_id} id={op_product_id} name={`${op_product_name} (${op_product_id})`} />
        ))}
      </Select>
    );
  },
});
