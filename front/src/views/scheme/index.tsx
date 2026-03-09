import { defineComponent } from 'vue';
import { useRouter, useRoute, RouterView } from 'vue-router';
import { MENU_SCHEME_RECOMMENDATION, MENU_SCHEME_LIST, MENU_SCHEME_DETAIL } from '@/constants/menu-symbol';

import './index.scss';

export default defineComponent({
  name: 'ResourceSelection',
  setup() {
    const route = useRoute();
    const router = useRouter();

    const TAB_LIST = [
      { routeName: MENU_SCHEME_RECOMMENDATION, label: '资源选型', icon: 'bkhcm-icon-xuanze' },
      { routeName: MENU_SCHEME_LIST, label: '选型方案', icon: 'bkhcm-icon-bushu' },
    ];

    const isActived = (name: symbol) => {
      if (name === MENU_SCHEME_RECOMMENDATION) {
        return route.name === name;
      }
      return [MENU_SCHEME_LIST, MENU_SCHEME_DETAIL].includes(route.name as symbol);
    };

    const handleTabChange = (routeName: symbol) => {
      router.push({ name: routeName });
    };

    return () => (
      <div class='resource-selection-module'>
        <header class='module-header'>
          <section class='tab-list'>
            {TAB_LIST.map(({ routeName, label, icon }) => {
              return (
                <div
                  class={`tab-item${isActived(routeName) ? ' actived' : ''}`}
                  key={label}
                  onClick={() => handleTabChange(routeName)}>
                  <i class={`hcm-icon ${icon}`}></i>
                  {label}
                </div>
              );
            })}
          </section>
        </header>
        <section class='module-page-container'>
          <RouterView />
        </section>
      </div>
    );
  },
});
