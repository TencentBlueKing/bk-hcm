import { defineComponent, ref } from 'vue';
// import components
import LbTree from './lb-tree';
import AllClbsManager from './all-clbs-manager';
import SpecificListenerManager from './specific-listener-manager';
import SpecificClbManager from './specific-clb-manager';
import SpecificDomainManager from './specific-domain-manager';
import './index.scss';

type NodeType = 'all' | 'load_balancers' | 'listeners' | 'domains';

export default defineComponent({
  name: 'LoadBalancerView',
  setup() {
    const isAdvancedSearchShow = ref(false);

    const componentMap = {
      all: <AllClbsManager />,
      load_balancers: <SpecificClbManager />,
      listeners: <SpecificListenerManager />,
      domain: <SpecificDomainManager />,
    };
    const renderComponent = (type: NodeType) => {
      return componentMap[type];
    };

    const activeType = ref<NodeType>('all');

    return () => (
      <div class='clb-view-page'>
        <div class='left-container'>
          <LbTree v-model:activeType={activeType.value} />
        </div>
        {isAdvancedSearchShow.value && <div class='advanced-search'>高级搜索</div>}
        <div class='main-container'>{renderComponent(activeType.value)}</div>
      </div>
    );
  },
});
