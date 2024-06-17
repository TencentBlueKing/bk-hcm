import { defineComponent } from 'vue';
import { useRoute } from 'vue-router';

import cssModule from './index.module.scss';
import { Button } from 'bkui-vue';
import VendorSelector from './vendor-selector';

import { useI18n } from 'vue-i18n';
import PrimaryAccountSelector from './primary-account-selector';
import SubAccountSelector from './sub-account-selector';
import OperationProductSelector from './operation-product-selector';

export default defineComponent({
  setup() {
    const { t } = useI18n();
    const route = useRoute();

    return () => (
      <div class={cssModule['search-container']}>
        <div class={cssModule['search-grid']}>
          {route.name === 'billSummary' && (
            <div>
              <div class={cssModule['search-label']}>{t('云厂商')}</div>
              <VendorSelector />
            </div>
          )}
          <div>
            <div class={cssModule['search-label']}>{t('一级账号')}</div>
            <PrimaryAccountSelector />
          </div>
          <div>
            <div class={cssModule['search-label']}>{t('运营产品')}</div>
            <OperationProductSelector />
          </div>
          {route.name === 'billDetail' && (
            <div>
              <div class={cssModule['search-label']}>{t('二级账号')}</div>
              <SubAccountSelector />
            </div>
          )}
        </div>
        <Button theme='primary' class={cssModule['search-button']}>
          查询
        </Button>
        <Button class={cssModule['search-button']}>重置</Button>
      </div>
    );
  },
});
