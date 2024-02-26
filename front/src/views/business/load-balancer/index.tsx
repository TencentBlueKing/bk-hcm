import { defineComponent } from 'vue';
import { useRoute, useRouter, RouterView } from 'vue-router';
import './index.scss';

export default defineComponent({
  name: 'LoadBalancer',
  setup() {
    const router = useRouter();
    const route = useRoute();

    const TAB_LIST = [
      { routeName: 'loadbalancer-view', label: '负载均衡视角' },
      { routeName: 'target-group-view', label: '目标组视角' },
    ];

    const isActive = (routeName: string) => {
      return routeName === route.name;
    };

    const handleTabChange = (routeName: string) => {
      router.push({ name: routeName });
    };

    return () => (
      <div class='business-loadbalancer-module'>
        <header class='module-header'>
          <section class='title-wrap'>负载均衡</section>
          <section class='tab-list-wrap'>
            {TAB_LIST.map(({ routeName, label }) => {
              return (
                <div
                  key={routeName}
                  class={`tab-item${isActive(routeName) ? ' active' : ''}`}
                  onClick={() => handleTabChange(routeName)}>
                  {label}
                </div>
              );
            })}
          </section>
        </header>
        <section class='module-page-container'>
          <RouterView></RouterView>
        </section>
      </div>
    );
  },
});
