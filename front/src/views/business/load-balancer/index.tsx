import { defineComponent } from 'vue';
import { useRoute, useRouter, RouterView } from 'vue-router';
import './index.scss';

export default defineComponent({
  name: 'LoadBalancer',
  setup() {
    const router = useRouter();
    const route = useRoute();

    const TAB_LIST = [
      { path: '/business/loadbalancer/clb-view', label: '负载均衡视角' },
      { path: '/business/loadbalancer/group-view', label: '目标组视角' },
    ];

    const isActive = (path: string) => {
      return route.path.includes(path);
    };

    const handleTabChange = (path: string) => {
      router.push({ path, query: { bizs: route.query.bizs } });
    };

    return () => (
      <div class='business-loadbalancer-module'>
        <header class='module-header'>
          <section class='title-wrap'>负载均衡</section>
          <section class='tab-list-wrap'>
            {TAB_LIST.map(({ path, label }) => {
              return (
                <div
                  key={path}
                  class={`tab-item${isActive(path) ? ' active' : ''}`}
                  onClick={() => handleTabChange(path)}>
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
