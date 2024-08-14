import { computed, defineComponent, provide, ref } from 'vue';
import { RouterView } from 'vue-router';
import Header from './header';
import dayjs from 'dayjs';
import './index.scss';

export default defineComponent({
  setup() {
    const currentMonth = ref(dayjs().subtract(1, 'month').toDate());
    const bill_year = computed(() => currentMonth.value.getFullYear());
    const bill_month = computed(() => currentMonth.value.getMonth() + 1);
    provide('currentMonth', currentMonth);
    provide('bill_year', bill_year);
    provide('bill_month', bill_month);

    return () => (
      <div class='bill-manage-module'>
        <Header />
        <RouterView class='main-container'></RouterView>
      </div>
    );
  },
});
