import { VendorEnum } from '@/common/constant';
import { Select } from 'bkui-vue';
import { PropType, defineComponent, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

export default defineComponent({
  props: { modelValue: String as PropType<VendorEnum> },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const { t } = useI18n();
    const selectedValue = ref(props.modelValue);

    const list = ref([
      { id: VendorEnum.TCLOUD, name: t('腾讯云') },
      { id: VendorEnum.AWS, name: t('亚马逊云') },
      { id: VendorEnum.AZURE, name: t('微软云') },
      { id: VendorEnum.GCP, name: t('谷歌云') },
      { id: VendorEnum.HUAWEI, name: t('华为云') },
      { id: VendorEnum.ZENLAYER, name: t('zenlayer') },
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
