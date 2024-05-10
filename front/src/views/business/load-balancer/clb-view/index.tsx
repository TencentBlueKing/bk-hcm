import { computed, defineComponent } from 'vue';
import { RouterView, useRoute } from 'vue-router';
// import components
import { ResizeLayout } from 'bkui-vue';
import LbTree from './lb-tree';
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
      <ResizeLayout class='clb-view-page' collapsible initialDivide={300}>
        {{
          aside: () => (
            <div class='left-container'>
              <LbTree />
            </div>
          ),
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
