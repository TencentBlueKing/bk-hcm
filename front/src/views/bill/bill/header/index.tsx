import { Ref, defineComponent, inject, ref } from 'vue';
import { DatePicker } from 'bkui-vue';
import { RouterLink, useRoute } from 'vue-router';
import './index.scss';
import { useI18n } from 'vue-i18n';
import { BILL_BIZS_KEY, BILL_MAIN_ACCOUNTS_KEY } from '@/constants';

export default defineComponent({
  setup() {
    const route = useRoute();
    const { t } = useI18n();

    const currentMonth = inject<Ref<Date>>('currentMonth');

    const links = ref([
      { name: 'billSummary', title: t('账单汇总') },
      { name: 'billDetail', title: t('账单明细') },
      { name: 'billAdjust', title: t('账单调整') },
    ]);

    return () => (
      <div class='header-container'>
        <div class='title-wrap'>
          <div class='title'>{t('云账单管理')}</div>
          <DatePicker v-model={currentMonth.value} type='month' clearable={false} />
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
