import { Ref } from 'vue';
import i18n from '@/language/i18n';
import BusinessSelector from '@/components/business-selector/index.vue';

const { t } = i18n.global;

export const diffModelKey = 'bk_biz_id';
export const getDiffModelLabel = () => t('业务');

export const getDiffSelectorComp = (formModel: any, selectorRef: Ref) => {
  return <BusinessSelector v-model={formModel.bk_biz_id} ref={selectorRef} isEditable />;
};
