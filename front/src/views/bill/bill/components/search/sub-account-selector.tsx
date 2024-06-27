import { VendorEnum } from '@/common/constant';
import { useSingleList } from '@/hooks/useSingleList';
import { QueryRuleOPEnum } from '@/typings';
import { SelectColumn } from '@blueking/ediatable';
import { Select } from 'bkui-vue';
import { PropType, defineComponent, ref, watch } from 'vue';

export default defineComponent({
  props: {
    modelValue: String as PropType<string>,
    vendor: String as PropType<VendorEnum>,
    // 是否用于 ediatable
    isEditable: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['update:modelValue'],
  setup(props, { emit, expose }) {
    const selectedValue = ref(props.modelValue);
    const selectRef = ref();

    const { dataList, isDataLoad, loadDataList, handleScrollEnd } = useSingleList({
      url: '/api/v1/account/main_accounts/list',
      rules: () => (props.vendor ? [{ field: 'vendor', op: QueryRuleOPEnum.EQ, value: props.vendor }] : []),
      immediate: true,
    });

    const getValue = () => {
      return selectRef.value.getValue().then(() => selectedValue.value);
    };

    expose({
      getValue,
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

    if (props.isEditable) {
      return () => (
        <SelectColumn
          onScroll-end={handleScrollEnd}
          loading={isDataLoad.value}
          scrollLoading={isDataLoad.value}
          v-model={selectedValue.value}
          list={dataList.value.map(({ id, cloud_id }) => ({
            label: cloud_id as string,
            value: id as string,
            key: id as string,
          }))}
          ref={selectRef}
          rules={[
            {
              validator: (value: string) => Boolean(value),
              message: '二级账号不能为空',
            },
          ]}
        />
      );
    }

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
