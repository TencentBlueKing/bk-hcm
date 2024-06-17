import { defineComponent, provide, ref } from 'vue';
import { RouterView } from 'vue-router';
import Header from './header';
import './index.scss';

export default defineComponent({
  setup() {
    const currentMonth = ref(new Date());
    provide('currentMonth', currentMonth);

    return () => (
      <div class='bill-manage-module'>
        <Header />
        <RouterView class='main-container'></RouterView>
      </div>
    );
  },
});
