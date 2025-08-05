import { Reactive, ref, watch } from 'vue';
import { Props } from 'bkui-vue/lib/info-box/info-box';

import { InfoBox } from 'bkui-vue';

export const useSideslider = (formModel: Reactive<any>, config: Partial<Props> = {}) => {
  const defaultConfig: Partial<Props> = {
    title: '确认离开当前页？',
    subTitle: '离开会导致未保存信息丢失',
    headerAlign: 'center',
    footerAlign: 'center',
    confirmText: '离开',
    cancelText: '取消',
  };

  const isChange = ref(false);

  const beforeClose = () => {
    if (isChange.value) {
      return new Promise((resolve, reject) => {
        InfoBox({
          ...defaultConfig,
          ...config,
          onConfirm: () => resolve(true),
          onClose: () => reject(false),
        });
      });
    }
    return true;
  };

  watch(
    formModel,
    () => {
      isChange.value = true;
    },
    { deep: true, once: true },
  );

  return { beforeClose };
};
