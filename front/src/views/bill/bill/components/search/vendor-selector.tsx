import { VendorEnum } from '@/common/constant';
import { Select } from 'bkui-vue';
import { PropType, defineComponent, ref, watch } from 'vue';

export default defineComponent({
  props: { modelValue: String as PropType<VendorEnum> },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const selectedValue = ref(props.modelValue);

    const list = ref([
      { id: VendorEnum.TCLOUD, name: '腾讯云' },
      { id: VendorEnum.AWS, name: '亚马逊云' },
      { id: VendorEnum.AZURE, name: '微软云' },
      { id: VendorEnum.GCP, name: '谷歌云' },
      { id: VendorEnum.HUAWEI, name: '华为云' },
      { id: VendorEnum.ZENLAYER, name: 'zenlayer' },
    ]);

    watch(
      () => props.modelValue,
      (val) => {
        selectedValue.value = val;
      },
    );

    watch(selectedValue, (val) => {
      emit('update:modelValue', val);
    });

    return () => (
      <Select v-model={selectedValue.value} multiple multipleMode='tag'>
        {list.value.map(({ id, name }) => (
          <Select.Option key={id} id={id} name={name} />
        ))}
      </Select>
    );
  },
});
