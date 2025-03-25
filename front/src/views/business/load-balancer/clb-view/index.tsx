import { computed, defineComponent, provide } from 'vue';
import { RouterView, useRoute } from 'vue-router';
import { useWhereAmI, Senarios } from '@/hooks/useWhereAmI';
import { ResizeLayout } from 'bkui-vue';
import LBTree from './lb-tree/index.vue';
import LbBreadcrumb from '../components/lb-breadcrumb';
import './index.scss';

export default defineComponent({
  name: 'LoadBalancerView',
  setup() {
    const route = useRoute();
    const { whereAmI } = useWhereAmI();

    const createActionName = computed(() => {
      if (whereAmI.value === Senarios.business) return 'biz_clb_resource_create';
      return 'clb_resource_create';
    });
    const deleteActionName = computed(() => {
      if (whereAmI.value === Senarios.business) return 'biz_clb_resource_delete';
      return 'clb_resource_delete';
    });
    provide('createActionName', createActionName);
    provide('deleteActionName', deleteActionName);

    // 面包屑白名单
    const BREADCRUMB_WHITE_LIST = ['lb', 'listener', 'domain'];
    const isBreadcrumbShow = computed(() => {
      return BREADCRUMB_WHITE_LIST.includes(route.meta.type as string);
    });

    return () => (
      <ResizeLayout class='clb-view-page' collapsible initialDivide={300} min={300}>
        {{
          aside: () => <LBTree class='load-balancer-tree' />,
          main: () => (
            <div class='main-container'>
              {isBreadcrumbShow.value && <LbBreadcrumb />}
              {/* 四级路由 */}
              <RouterView />
            </div>
          ),
        }}
      </ResizeLayout>
    );
  },
});
