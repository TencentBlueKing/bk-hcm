import { computed, defineComponent, ref } from 'vue';
import { RouterView, useRoute } from 'vue-router';
// import components
import LbTree from './lb-tree';
import LbBreadcrumb from '../components/lb-breadcrumb';
import './index.scss';

export default defineComponent({
  name: 'LoadBalancerView',
  setup() {
    const route = useRoute();

    const isAdvancedSearchShow = ref(false);

    // 面包屑白名单
    const BREADCRUMB_WHITE_LIST = ['listener', 'domain'];
    const isBreadcrumbShow = computed(() => {
      return BREADCRUMB_WHITE_LIST.includes(route.meta.type as string);
    });

    return () => (
      <div class='clb-view-page'>
        <div class='left-container'>
          <LbTree />
        </div>
        {isAdvancedSearchShow.value && <div class='advanced-search'>高级搜索</div>}
        <div class='main-container'>
          {isBreadcrumbShow.value && <LbBreadcrumb />}
          {/* 四级路由 */}
          <RouterView />
        </div>
      </div>
    );
  },
});
