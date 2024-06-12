import { defineComponent, provide, ref } from 'vue';
import { RouterLink, RouterView } from 'vue-router';
import { DatePicker } from 'bkui-vue';
import './index.scss';

export default defineComponent({
  setup() {
    const currentMonth = ref(new Date());
    provide('currentMonth', currentMonth);

    const links = ref([
      { name: 'billSummary', title: '账单汇总' },
      { name: 'billDetail', title: '账单明细' },
    ]);

    return () => (
      <div class='bill-manage-module'>
        <header class='header-container'>
          <div class='title-wrap'>
            <div class='title'>云账单管理</div>
            <DatePicker v-model={currentMonth.value} type='month' clearable={false} />
          </div>
          <div class='link-wrap'>
            {links.value.map(({ name, title }) => (
              <RouterLink class='link-item' to={{ name }} activeClass='active'>
                {title}
              </RouterLink>
            ))}
          </div>
        </header>
        <RouterView class='main-container'></RouterView>
      </div>
    );
  },
});
