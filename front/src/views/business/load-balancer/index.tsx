import { computed, defineComponent, provide } from 'vue';
import { useRoute, useRouter, RouterView } from 'vue-router';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import './index.scss';

export default defineComponent({
  name: 'LoadBalancer',
  setup() {
    const router = useRouter();
    const route = useRoute();
    const { whereAmI } = useWhereAmI();

    const TAB_LIST = [
      { path: '/business/loadbalancer/clb-view', label: '负载均衡视角' },
      { path: '/business/loadbalancer/group-view', label: '目标组视角' },
    ];

    const createClbActionName = computed(() => {
      if (whereAmI.value === Senarios.business) return 'biz_clb_resource_create';
      return 'clb_resource_create';
    });
    const deleteClbActionName = computed(() => {
      if (whereAmI.value === Senarios.business) return 'biz_clb_resource_delete';
      return 'clb_resource_delete';
    });
    provide('createClbActionName', createClbActionName);
    provide('deleteClbActionName', deleteClbActionName);

    const isActive = (path: string) => {
      return route.path.includes(path);
    };

    const handleTabChange = (path: string) => {
      router.push({ path, query: { [GLOBAL_BIZS_KEY]: route.query[GLOBAL_BIZS_KEY] } });
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
