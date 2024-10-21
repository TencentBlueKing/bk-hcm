import { Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { BILL_BIZS_KEY } from '@/constants';
import { ISearchModal } from '.';

import BusinessSelector from '@/components/business-selector/index.vue';

export const renderProductComp = (modal: Ref<ISearchModal>) => {
  const { t } = useI18n();
  const label = t('业务');

  const render = () => {
    return <BusinessSelector v-model={modal.value.bk_biz_id} clearable multiple urlKey={BILL_BIZS_KEY} base64Encode />;
  };

  return { label, render };
};
