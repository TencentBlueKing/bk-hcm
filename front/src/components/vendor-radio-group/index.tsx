import { PropType, defineComponent, ref, watch } from 'vue';
import cssModule from './index.module.scss';

import { Button } from 'bkui-vue';
import { BkButtonGroup } from 'bkui-vue/lib/button';
import vendorAzure from '@/assets/image/vendor-azure.svg';
import vendorGCP from '@/assets/image/vendor-gcp.svg';
import vendorAWS from '@/assets/image/vendor-aws.svg';
import vendorHuawei from '@/assets/image/vendor-huawei.svg';
import vendorTcloud from '@/assets/image/vendor-tcloud.svg';

import { useI18n } from 'vue-i18n';
import { VendorEnum } from '@/common/constant';

export default defineComponent({
  props: {
    modelValue: [String, Array] as PropType<VendorEnum | VendorEnum[]>,
    size: {
      type: String as PropType<'small' | 'normal'>,
      default: 'normal',
    },
    disabled: Boolean,
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const { t } = useI18n();

    const vendor = ref(props.modelValue);

    const buttons = ref([
      { label: t('微软云'), value: VendorEnum.AZURE, icon: vendorAzure },
      { label: t('谷歌云'), value: VendorEnum.GCP, icon: vendorGCP },
      { label: t('华为云'), value: VendorEnum.HUAWEI, icon: vendorHuawei },
      { label: t('亚马逊云'), value: VendorEnum.AWS, icon: vendorAWS },
      { label: t('zenlayer'), value: VendorEnum.ZENLAYER, icon: vendorTcloud },
      { label: t('腾讯云'), value: VendorEnum.TCLOUD, icon: vendorTcloud },
    ]);

    watch(vendor, (v) => emit('update:modelValue', v), { deep: true });

    watch(
      () => props.modelValue,
      (v) => (vendor.value = v),
      {
        deep: true,
      },
    );

    return () => (
      <BkButtonGroup
        class={[
          cssModule.group,
          { [cssModule.small]: props.size === 'small', [cssModule.normal]: props.size === 'normal' },
        ]}
        v-model={vendor.value}>
        {buttons.value.map(({ label, value, icon }) => (
          <Button
            class={cssModule.radio}
            selected={vendor.value === value}
            onClick={() => (vendor.value = value)}
            disabled={props.disabled}>
            <img src={icon} alt='' />
            <span>{label}</span>
          </Button>
        ))}
      </BkButtonGroup>
    );
  },
});
