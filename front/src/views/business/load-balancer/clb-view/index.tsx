import { computed, defineComponent } from 'vue';
import { RouterView, useRoute } from 'vue-router';
import { ResizeLayout } from 'bkui-vue';
import LBTree from './lb-tree/index.vue';
import LbBreadcrumb from '../components/lb-breadcrumb';
import './index.scss';

export default defineComponent({
  name: 'LoadBalancerView',
  setup() {
    const route = useRoute();

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
