import { defineComponent, ref } from 'vue';
// import components
import LbTree from './lb-tree';
import LbBreadcrumb from '../components/lb-breadcrumb';
import AllClbsManager from './all-clbs-manager';
import SpecificListenerManager from './specific-listener-manager';
import SpecificClbManager from './specific-clb-manager';
import SpecificDomainManager from './specific-domain-manager';
import './index.scss';

type NodeType = 'all' | 'lb' | 'listener' | 'domain';

export default defineComponent({
  name: 'LoadBalancerView',
  setup() {
    const isAdvancedSearchShow = ref(false);

    const componentMap = {
      all: <AllClbsManager />,
      lb: <SpecificClbManager />,
      listener: <SpecificListenerManager />,
      domain: <SpecificDomainManager />,
    };
    const renderComponent = (type: NodeType) => {
      return componentMap[type];
    };

    const activeType = ref<NodeType>('all');

    // 面包屑白名单
    const BREADCRUMB_WHITE_LIST = ['listener', 'domain'];

    return () => (
      <div class='clb-view-page'>
        <div class='left-container'>
          <LbTree v-model:activeType={activeType.value} />
        </div>
        {isAdvancedSearchShow.value && <div class='advanced-search'>高级搜索</div>}
        <div class='main-container'>
          {BREADCRUMB_WHITE_LIST.includes(activeType.value) && <LbBreadcrumb />}
          {renderComponent(activeType.value)}
        </div>
      </div>
    );
  },
});
