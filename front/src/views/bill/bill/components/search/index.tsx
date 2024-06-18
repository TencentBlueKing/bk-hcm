import { defineComponent, ref } from 'vue';
import { useRoute } from 'vue-router';

import cssModule from './index.module.scss';
import { Button } from 'bkui-vue';
import VendorSelector from './vendor-selector';
import PrimaryAccountSelector from './primary-account-selector';
import SubAccountSelector from './sub-account-selector';
import OperationProductSelector from './operation-product-selector';

import { useI18n } from 'vue-i18n';
import { VendorEnum } from '@/common/constant';

interface ISearchModal {
  vendor: VendorEnum[];
  root_account_id: string[];
  main_account_id: string[];
  product_id: string[];
}

export default defineComponent({
  emits: ['search'],
  setup(_, { emit }) {
    const { t } = useI18n();
    const route = useRoute();

    const getDefaultModal = (): ISearchModal => ({
      vendor: [],
      root_account_id: [],
      main_account_id: [],
      product_id: [],
    });
    const modal = ref(getDefaultModal());

    const handleSearch = () => {
      emit('search', modal.value);
    };

    const handleReset = () => {
      modal.value = getDefaultModal();
      handleSearch();
    };

    return () => (
      <div class={cssModule['search-container']}>
        <div class={cssModule['search-grid']}>
          {route.name === 'billSummary' && (
            <div>
              <div class={cssModule['search-label']}>{t('云厂商')}</div>
              <VendorSelector v-model={modal.value.vendor} />
            </div>
          )}
          <div>
            <div class={cssModule['search-label']}>{t('一级账号')}</div>
            <PrimaryAccountSelector v-model={modal.value.root_account_id} />
          </div>
          <div>
            <div class={cssModule['search-label']}>{t('运营产品')}</div>
            <OperationProductSelector v-model={modal.value.product_id} />
          </div>
          {route.name === 'billDetail' && (
            <div>
              <div class={cssModule['search-label']}>{t('二级账号')}</div>
              <SubAccountSelector v-model={modal.value.main_account_id} />
            </div>
          )}
        </div>
        <Button theme='primary' class={cssModule['search-button']} onClick={handleSearch}>
          查询
        </Button>
        <Button class={cssModule['search-button']} onClick={handleReset}>
          重置
        </Button>
      </div>
    );
  },
});
