import { defineComponent, ref } from 'vue';
import { useRoute } from 'vue-router';

import cssModule from './index.module.scss';
import { Button, DatePicker } from 'bkui-vue';
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
  update_time: string;
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
      update_time: '',
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
          {route.name !== 'billAdjust' && (
            <div>
              <div class={cssModule['search-label']}>{t('一级账号')}</div>
              <PrimaryAccountSelector v-model={modal.value.root_account_id} />
            </div>
          )}
          <div>
            <div class={cssModule['search-label']}>{t('运营产品')}</div>
            <OperationProductSelector v-model={modal.value.product_id} />
          </div>
          {['billDetail', 'billAdjust'].includes(route.name as string) && (
            <div>
              <div class={cssModule['search-label']}>{t('二级账号')}</div>
              <SubAccountSelector v-model={modal.value.main_account_id} />
            </div>
          )}
          {route.name === 'billAdjust' && (
            <div>
              <div class={cssModule['search-label']}>{t('更新时间')}</div>
              <DatePicker class={cssModule.datePicker} placeholder='如：2019-01-30 至 2019-01-30' />
            </div>
          )}
        </div>
        <Button theme='primary' class={cssModule['search-button']} onClick={handleSearch}>
          {t('查询')}
        </Button>
        <Button class={cssModule['search-button']} onClick={handleReset}>
          {t('重置')}
        </Button>
      </div>
    );
  },
});
