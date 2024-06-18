import { computed, defineComponent, provide, ref } from 'vue';
import { RouterView } from 'vue-router';
import Header from './header';
import './index.scss';

export default defineComponent({
  setup() {
    const currentMonth = ref(new Date());
    const bill_year = computed(() => currentMonth.value.getFullYear());
    const bill_month = computed(() => currentMonth.value.getMonth() + 1);
    provide('currentMonth', currentMonth);
    provide('bill_year', bill_year.value);
    provide('bill_month', bill_month.value);

    return () => (
      <div class='bill-manage-module'>
        <Header />
        <RouterView class='main-container'></RouterView>
      </div>
    );
  },
});
