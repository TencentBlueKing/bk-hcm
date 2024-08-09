import { Ref, defineComponent, inject, ref, watch } from 'vue';
import { DatePicker } from 'bkui-vue';
import { RouterLink, useRoute } from 'vue-router';
import './index.scss';
import { useI18n } from 'vue-i18n';
import { BILL_BIZS_KEY, BILL_MAIN_ACCOUNTS_KEY } from '@/constants';
import { reqBillsExchangeRateList } from '@/api/bill';
import { QueryRuleOPEnum } from '@/typings';

export default defineComponent({
  setup() {
    const route = useRoute();
    const { t } = useI18n();

    const currentMonth = inject<Ref<Date>>('currentMonth');
    const bill_year = inject<Ref<number>>('bill_year');
    const bill_month = inject<Ref<number>>('bill_month');

    const links = ref([
      { name: 'billSummary', title: t('账单汇总') },
      { name: 'billDetail', title: t('账单明细') },
      { name: 'billAdjust', title: t('账单调整') },
    ]);

    const currentMonthExchangeRate = ref('');

    watch(
      currentMonth,
      async () => {
        const res = await reqBillsExchangeRateList({
          filter: {
            op: QueryRuleOPEnum.AND,
            rules: [
              { field: 'year', op: QueryRuleOPEnum.EQ, value: bill_year.value },
              { field: 'month', op: QueryRuleOPEnum.EQ, value: bill_month.value },
              { field: 'from_currency', op: QueryRuleOPEnum.EQ, value: 'USD' },
              { field: 'to_currency', op: QueryRuleOPEnum.EQ, value: 'CNY' },
            ],
          },
          page: { start: 0, limit: 10, count: false },
        });
        currentMonthExchangeRate.value = res.data?.details[0]?.exchange_rate;
      },
      { immediate: true },
    );

    return () => (
      <div class='header-container'>
        <div class='title-wrap'>
          <div class='title'>{t('云账单管理')}</div>
          <DatePicker v-model={currentMonth.value} type='month' clearable={false} />
          <div class='rate-wrapper'>
            <span
              class='rate-label'
              v-bk-tooltips={{ content: t('当月汇率每月为固定值，由平台统一提供汇率，非实时汇率值。') }}>
              {t('当月汇率')}
            </span>
            ：
            <span class='rate-value'>
              {currentMonthExchangeRate.value
                ? `1$(${t('美金')})=${currentMonthExchangeRate.value}￥(${t('人民币')})`
                : t('本月汇率未提供')}
            </span>
          </div>
        </div>
        <div class='link-wrap'>
          {links.value.map(({ name, title }) => (
            <RouterLink
              class='link-item'
              to={{
                name,
                query: {
                  [BILL_BIZS_KEY]: route.query[BILL_BIZS_KEY],
                  [BILL_MAIN_ACCOUNTS_KEY]: route.query[BILL_MAIN_ACCOUNTS_KEY],
                },
              }}
              activeClass='active'>
              {title}
            </RouterLink>
          ))}
        </div>
      </div>
    );
  },
});
