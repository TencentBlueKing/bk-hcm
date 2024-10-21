import { PropType, computed, defineComponent, ref, watch } from 'vue';

import cssModule from './index.module.scss';
import { Button, DatePicker } from 'bkui-vue';
import VendorSelector from './vendor-selector';
import PrimaryAccountSelector from './primary-account-selector';
import SubAccountSelector from './sub-account-selector';
// import OperationProductSelector from './operation-product-selector';

import { useI18n } from 'vue-i18n';
import { VendorEnum } from '@/common/constant';
import { QueryRuleOPEnum, RulesItem } from '@/typings';
import dayjs from 'dayjs';
import { BILL_MAIN_ACCOUNTS_KEY } from '@/constants';
import { renderProductComp } from './render-comp.plugin';

export interface ISearchModal {
  vendor: VendorEnum[];
  root_account_id: string[];
  main_account_id: string[];
  product_id: string[];
  bk_biz_id: number[];
  updated_at: Date[];
}

type ISearchKeys = 'vendor' | 'root_account_id' | 'main_account_id' | 'product_id' | 'updated_at';

export default defineComponent({
  props: {
    searchKeys: {
      type: Array as PropType<ISearchKeys[]>,
      required: true,
    },
    vendor: {
      type: Array as PropType<VendorEnum[]>,
      default: [],
    },
    disableSearchHandler: {
      type: Function as PropType<(rules: RulesItem[]) => boolean>,
      default: () => false,
    },
    disabledTipContent: {
      type: String,
      default: '请选择搜索条件',
    },
    // 是否自动选中二级账号
    autoSelectMainAccount: Boolean,
  },
  emits: ['search'],
  setup(props, { emit, expose }) {
    const { t } = useI18n();

    const getDefaultModal = (): ISearchModal => ({
      vendor: props.vendor,
      root_account_id: [],
      main_account_id: [],
      product_id: [],
      bk_biz_id: [],
      updated_at: [],
    });
    const modal = ref(getDefaultModal());
    const rules = computed<RulesItem[]>(() => {
      const { vendor, root_account_id, main_account_id, product_id, bk_biz_id, updated_at } = modal.value;
      const rules = [
        { field: 'vendor', op: QueryRuleOPEnum.IN, value: vendor },
        { field: 'root_account_id', op: QueryRuleOPEnum.IN, value: root_account_id },
        { field: 'main_account_id', op: QueryRuleOPEnum.IN, value: main_account_id },
        { field: 'product_id', op: QueryRuleOPEnum.IN, value: product_id },
        { field: 'bk_biz_id', op: QueryRuleOPEnum.IN, value: bk_biz_id },
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
    });
    const disabledSearch = computed(() => {
      return props.disableSearchHandler(rules.value);
    });

    const { label, render } = renderProductComp(modal);

    const handleSearch = () => {
      // 如果搜索条件判定为空, 不触发搜索
      !disabledSearch.value && emit('search', rules.value);
    };

    const handleReset = () => {
      modal.value = getDefaultModal();
      handleSearch();
    };

    // 云厂商变化, 重置一级账号
    watch(
      () => modal.value.vendor,
      () => (modal.value.root_account_id = []),
      { deep: true },
    );

    // 一级账号变化, 重置二级账号
    watch(
      () => modal.value.root_account_id,
      () => (modal.value.main_account_id = []),
      { deep: true },
    );

    expose({ handleSearch, rules });

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
              <div class={cssModule['search-label']}>{label}</div>
              {render()}
            </div>
          )}
          {props.searchKeys.includes('main_account_id') && (
            <div>
              <div class={cssModule['search-label']}>{t('二级账号')}</div>
              <SubAccountSelector
                v-model={modal.value.main_account_id}
                vendor={modal.value.vendor}
                rootAccountId={modal.value.root_account_id}
                autoSelect={props.autoSelectMainAccount}
                urlKey={BILL_MAIN_ACCOUNTS_KEY}
              />
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
        <Button
          theme='primary'
          class={cssModule['search-button']}
          onClick={handleSearch}
          disabled={disabledSearch.value}
          v-bk-tooltips={{ content: props.disabledTipContent, disabled: !disabledSearch.value }}>
          {t('查询')}
        </Button>
        <Button class={cssModule['search-button']} onClick={handleReset}>
          {t('重置')}
        </Button>
      </div>
    );
  },
});
