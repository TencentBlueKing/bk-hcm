import { PropType, defineComponent, ref } from 'vue';

import cssModule from './index.module.scss';
import { Button, DatePicker } from 'bkui-vue';
import VendorSelector from './vendor-selector';
import PrimaryAccountSelector from './primary-account-selector';
import SubAccountSelector from './sub-account-selector';
// import OperationProductSelector from './operation-product-selector';

import { useI18n } from 'vue-i18n';
import { VendorEnum } from '@/common/constant';
import { QueryRuleOPEnum } from '@/typings';
import dayjs from 'dayjs';
import BusinessSelector from '@/components/business-selector/index.vue';
interface ISearchModal {
  vendor: VendorEnum[];
  root_account_id: string[];
  main_account_id: string[];
  product_id: string[];
  updated_at: Date[];
}

type ISearchKeys = 'vendor' | 'root_account_id' | 'main_account_id' | 'product_id' | 'updated_at';

export default defineComponent({
  props: {
    searchKeys: {
      type: Array as PropType<ISearchKeys[]>,
      required: true,
    },
  },
  emits: ['search'],
  setup(props, { emit, expose }) {
    const { t } = useI18n();

    const getDefaultModal = (): ISearchModal => ({
      vendor: [],
      root_account_id: [],
      main_account_id: [],
      product_id: [],
      updated_at: [],
    });
    const modal = ref(getDefaultModal());

    const getRules = () => {
      const { vendor, root_account_id, main_account_id, product_id, updated_at } = modal.value;
      const rules = [
        { field: 'vendor', op: QueryRuleOPEnum.IN, value: vendor },
        { field: 'root_account_id', op: QueryRuleOPEnum.IN, value: root_account_id },
        { field: 'main_account_id', op: QueryRuleOPEnum.IN, value: main_account_id },
        { field: 'product_id', op: QueryRuleOPEnum.IN, value: product_id },
        {
          field: 'updated_at',
          op: QueryRuleOPEnum.GTE,
          value: updated_at[0] ? dayjs(updated_at[0]).format('YYYY-MM-DDTHH:mm:ssZ') : '',
        },
        {
          field: 'updated_at',
          op: QueryRuleOPEnum.LTE,
          value: updated_at[1] ? dayjs(updated_at[1]).format('YYYY-MM-DDTHH:mm:ssZ') : '',
        },
      ];

      return rules.filter((rule) => {
        if (Array.isArray(rule.value)) {
          return rule.value.length;
        }
        return !!rule.value;
      });
    };

    const handleSearch = () => {
      emit('search', getRules());
    };

    const handleReset = () => {
      modal.value = getDefaultModal();
      handleSearch();
    };

    expose({ handleSearch });

    return () => (
      <div class={cssModule['search-container']}>
        <div class={cssModule['search-grid']}>
          {props.searchKeys.includes('vendor') && (
            <div>
              <div class={cssModule['search-label']}>{t('云厂商')}</div>
              <VendorSelector v-model={modal.value.vendor} />
            </div>
          )}
          {props.searchKeys.includes('root_account_id') && (
            <div>
              <div class={cssModule['search-label']}>{t('一级账号')}</div>
              <PrimaryAccountSelector v-model={modal.value.root_account_id} vendor={modal.value.vendor} />
            </div>
          )}
          {props.searchKeys.includes('product_id') && (
            <div>
              <div class={cssModule['search-label']}>{'业务'}</div>
              <BusinessSelector v-model={modal.value.product_id} />
            </div>
          )}
          {props.searchKeys.includes('main_account_id') && (
            <div>
              <div class={cssModule['search-label']}>{t('二级账号')}</div>
              <SubAccountSelector v-model={modal.value.main_account_id} />
            </div>
          )}
          {props.searchKeys.includes('updated_at') && (
            <div>
              <div class={cssModule['search-label']}>{t('更新时间')}</div>
              <DatePicker
                v-model={modal.value.updated_at}
                type='monthrange'
                class={cssModule.datePicker}
                placeholder='如：2019-01-30 至 2019-01-30'
              />
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
