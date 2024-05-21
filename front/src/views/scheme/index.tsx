import { defineComponent } from 'vue';
import { useRouter, useRoute, RouterView } from 'vue-router';

import './index.scss';

export default defineComponent({
  name: 'ResourceSelection',
  setup() {
    const route = useRoute();
    const router = useRouter();

    const TAB_LIST = [
      { routeName: 'scheme-recommendation', label: '资源选型', icon: 'bkhcm-icon-xuanze' },
      { routeName: 'scheme-list', label: '选型方案', icon: 'bkhcm-icon-bushu' },
    ];

    const isActived = (name: string) => {
      if (name === 'scheme-recommendation') {
        return route.name === name;
      }
      return ['scheme-list', 'scheme-detail'].includes(route.name as string);
    };

    const handleTabChange = (routeName: string) => {
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
                  key={routeName}
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
