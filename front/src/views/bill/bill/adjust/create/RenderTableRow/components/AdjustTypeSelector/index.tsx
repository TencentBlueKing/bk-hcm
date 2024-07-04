import { defineComponent, ref, watch } from 'vue';
import { SelectColumn } from '@blueking/ediatable';

export default defineComponent({
  props: {
    modelValue: String,
  },
  setup(props, { emit }) {
    const selectedVal = ref(props.modelValue || AdjustTypeEnum.Increase);

    watch(
      () => selectedVal.value,
      (val) => {
        emit('update:modelValue', val);
      },
    );

    watch(
      () => props.modelValue,
      () => {
        selectedVal.value = props.modelValue;
      },
    );

    return () => <SelectColumn list={AdjustTypeList} v-model={selectedVal.value}></SelectColumn>;
  },
});

export enum AdjustTypeEnum {
  Increase = 'increase',
  Decrease = 'decrease',
}

export const AdjustTypeList = [
  {
    label: '增加',
    value: AdjustTypeEnum.Increase,
  },
  {
    label: '减少',
    value: AdjustTypeEnum.Decrease,
  },
];
